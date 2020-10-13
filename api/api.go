/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package api

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/storage"
)

type Api struct {
	server *http.Server
	// baseRouter server static files and other public content as well as private endpoints
	baseRouter *mux.Router
	// privateRouter routes only authenticated endpoints
	privateRouter *mux.Router

	cors http.Handler
	db   *storage.Database
}

// NewApi initializes new api instance. It connects to database and opens http port.
func NewApi(database *storage.Database) (*Api, error) {
	api := &Api{
		baseRouter: mux.NewRouter(),
		db:         database,
	}

	c := cors.AllowAll()
	api.cors = c.Handler(api.baseRouter)

	api.server = &http.Server{
		Handler:      api.cors,
		Addr:         fmt.Sprintf("%s:%d", config.C.Api.Host, config.C.Api.Port),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
	}

	api.privateRouter = api.baseRouter.PathPrefix("/api/v1").Subrouter()

	api.addRoutes()

	return api, nil
}

func (a *Api) addRoutes() {
	//a.baseRouter.Use(a.corsHeader)
	a.baseRouter.Use(LoggingMiddleware)
	a.baseRouter.HandleFunc("/api/v1/auth/login", a.login).Methods(http.MethodPost)
	a.baseRouter.HandleFunc("/api/v1/version", a.getVersion).Methods(http.MethodGet)

	a.privateRouter.Use(a.authorizeUser)
	a.privateRouter.HandleFunc("/documents", a.getDocuments).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/show", a.getDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}", a.getDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/jobs", a.getDocumentLogs).Methods(http.MethodGet)

	a.baseRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/build/static/"))))
	a.baseRouter.Handle("/", http.FileServer(http.Dir("./frontend/build/")))

}

func (a *Api) Serve() error {
	return a.server.ListenAndServe()
}

type VersionResponse struct {
	Name    string
	Version string
}

func (a *Api) getVersion(resp http.ResponseWriter, req *http.Request) {
	v := &VersionResponse{
		Name:    "VirtualPaper",
		Version: config.Version,
	}
	respOk(resp, v)
}

type LoggingWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (l *LoggingWriter) WriteHeader(status int) {
	l.length = 0
	l.status = status
	l.ResponseWriter.WriteHeader(status)
}

func (l *LoggingWriter) Write(b []byte) (int, error) {
	l.length = len(b)
	if l.status == 0 {
		l.status = 200
	}
	return l.ResponseWriter.Write(b)
}

// LogginMiddlware Provide logging for requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger := &LoggingWriter{
			ResponseWriter: w,
		}
		next.ServeHTTP(logger, r)

		duration := time.Since(start).String()
		verb := r.Method
		url := r.RequestURI

		fields := make(map[string]interface{})
		fields["verb"] = verb
		fields["request"] = url
		fields["duration"] = duration
		fields["status"] = logger.status
		fields["length"] = logger.length
		logrus.WithFields(fields).Infof("http")
	})
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}
