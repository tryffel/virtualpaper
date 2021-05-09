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
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"net/http"
	"path"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/process"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

type Api struct {
	server *http.Server
	// baseRouter server static files and other public content as well as private endpoints
	baseRouter *mux.Router
	// privateRouter routes only authenticated endpoints
	privateRouter *mux.Router

	adminRouter *mux.Router

	cors    http.Handler
	db      *storage.Database
	search  *search.Engine
	process *process.Manager
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

	var err error
	api.search, err = search.NewEngine(database)
	if err != nil {
		return api, err
	}

	api.process, err = process.NewManager(database, api.search)
	if err != nil {
		return api, err
	}

	api.privateRouter = api.baseRouter.PathPrefix("/api/v1").Subrouter()
	api.adminRouter = api.baseRouter.PathPrefix("/api/v1/admin").Subrouter()
	api.addRoutes()
	return api, err
}

func (a *Api) addRoutes() {
	if len(config.C.Api.CorsHosts) > 0 {
		a.baseRouter.Use(a.corsHeader)
	}

	a.baseRouter.Use(LoggingMiddleware)
	a.baseRouter.HandleFunc("/api/v1/auth/login", a.login).Methods(http.MethodPost)
	a.baseRouter.HandleFunc("/api/v1/version", a.getVersion).Methods(http.MethodGet)

	a.privateRouter.Use(a.authorizeUser)
	a.privateRouter.HandleFunc("/documents", a.getDocuments).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/undefined", a.getEmptyDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/show", a.getDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id:[a-zA-Z0-9-]{30,40}}", a.getDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}", a.updateDocument).Methods(http.MethodPut)
	a.privateRouter.HandleFunc("/documents/{id}/preview", a.getDocumentPreview).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/jobs", a.getDocumentLogs).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents", a.uploadFile).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/create", a.uploadFile).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/undefined", a.uploadFile).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/create", a.getEmptyDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/content", a.getDocumentContent).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/download", a.downloadDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/metadata", a.updateDocumentMetadata).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/{id}/process", a.requestDocumentProcessing).Methods(http.MethodPost)

	a.privateRouter.HandleFunc("/jobs", a.GetJob).Methods(http.MethodGet)

	a.privateRouter.HandleFunc("/tags", a.getTags).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/tags/{id}", a.getTag).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/tags", a.createTag).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/tags/create", a.createTag).Methods(http.MethodPost)

	a.privateRouter.HandleFunc("/metadata/keys", a.getMetadataKeys).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/metadata/keys", a.addMetadataKey).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/metadata/keys/{id}", a.getMetadataKey).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/metadata/keys/{id}/values", a.getMetadataKeyValues).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/metadata/keys/{id}/values", a.addMetadataValue).Methods(http.MethodPost)

	a.privateRouter.HandleFunc("/documents/stats", a.getUserDocumentStatistics).Methods(http.MethodGet)

	a.privateRouter.HandleFunc("/processing/rules", a.addUserRule).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/processing/rules", a.getUserRules).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/processing/rules/{id}", a.getUserRule).Methods(http.MethodGet)

	a.privateRouter.HandleFunc("/preferences/user", a.getUserPreferences).Methods(http.MethodGet)

	a.privateRouter.HandleFunc("/filetypes", a.getSupportedFileTypes).Methods(http.MethodGet)

	a.adminRouter.Use(a.authorizeUser, a.authorizeAdmin)
	a.adminRouter.HandleFunc("/documents/process", a.forceDocumentProcessing).Methods(http.MethodPost)
	a.adminRouter.HandleFunc("/documents/process", a.getDocumentProcessQueue).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/admin/systeminfo", a.getSystemInfo).Methods(http.MethodGet)

	if config.C.Api.StaticContentPath != "" {
		logrus.Debugf("Serve static files")
		a.baseRouter.Handle("/", http.FileServer(http.Dir(config.C.Api.StaticContentPath)))
		a.baseRouter.PathPrefix("/static").
			Handler(http.StripPrefix("/static/",
				http.FileServer(http.Dir(path.Join(config.C.Api.StaticContentPath, "static")))))
	}

}

func (a *Api) Serve() error {
	err := a.process.Start()
	if err != nil {
		return err
	}
	return a.server.ListenAndServe()
}

// VersionResponse contains general server info.
type VersionResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// MimeTypesSupportedResponse conatains info on mime types that server can extract.
type MimeTypesSupportedResponse struct {
	Names     []string `json:"names"`
	Mimetypes []string `json:"mimetypes"`
}

func (a *Api) getVersion(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/version Public GetVersion
	// Get server version
	//
	// responses:
	//   200:
	v := &VersionResponse{
		Name:    "VirtualPaper",
		Version: config.Version,
		Commit:  config.Commit,
	}
	respOk(resp, v)
}

func (a *Api) getSupportedFileTypes(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/filetypes Public GetFileTypes
	// Get supported file types
	//
	// responses:
	//   200:

	mimetypes, filetypes := process.SupportedFileTypes()

	mimes := &MimeTypesSupportedResponse{
		Names:     filetypes,
		Mimetypes: mimetypes,
	}

	respOk(resp, mimes)
}

func (a *Api) getEmptyResp(resp http.ResponseWriter, req *http.Request) {
	respOk(resp, nil)
}

type LoggingWriter struct {
	resp   http.ResponseWriter
	status int
	length int
}

func (l *LoggingWriter) WriteHeader(status int) {
	l.length = 0
	l.status = status
	l.resp.WriteHeader(status)
}

func (l *LoggingWriter) Write(b []byte) (int, error) {
	l.length = len(b)
	if l.status == 0 {
		l.status = 200
	}
	return l.resp.Write(b)
}

func (l *LoggingWriter) Header() http.Header {
	return l.resp.Header()
}

// LogginMiddlware Provide logging for requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger := &LoggingWriter{
			resp: w,
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

		if config.C.Logging.HttpLog != nil {
			config.C.Logging.HttpLog.WithFields(fields).Infof("http")
		} else {
			logrus.WithFields(fields).Infof("http")
		}
	})
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}
