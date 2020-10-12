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
	db            *storage.Database
}

// NewApi initializes new api instance. It connects to database and opens http port.
func NewApi() (*Api, error) {
	api := &Api{
		baseRouter: mux.NewRouter(),
	}

	api.server = &http.Server{
		Handler:      api.baseRouter,
		Addr:         fmt.Sprintf("%s:%d", config.C.Api.Host, config.C.Api.Port),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
	}

	api.privateRouter = api.baseRouter.PathPrefix("/api/v1").Subrouter()
	api.addRoutes()

	var err error
	api.db, err = storage.NewDatabase()

	if err != nil {
		return api, fmt.Errorf("initialize database: %v", err)
	}
	return api, nil
}

func (a *Api) addRoutes() {
	a.baseRouter.HandleFunc("/api/v1/auth/login", a.login).Methods(http.MethodPost)
	a.baseRouter.HandleFunc("/api/v1/version", a.getVersion).Methods(http.MethodGet)
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

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}
