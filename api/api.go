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
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
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

func (a *Api) Serve() error {
	err := a.process.Start()
	if err != nil {
		return err
	}

	logrus.Infof("listen http on %s", a.server.Addr)
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
	//   200: RespVersion
	v := &VersionResponse{
		Name:    "VirtualPaper",
		Version: config.Version,
		Commit:  config.Commit,
	}
	respOk(resp, v)
}

func (a *Api) getSupportedFileTypes(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/filetypes Public GetFileTypes
	// Get supported file types.
	// Returns a list of valid name endings and a list of mime types.
	//
	// responses:
	//   200: RespFileTypes

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

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

//go:embed swaggerdocs/swagger.json
var swaggerJson string

func serverSwaggerDoc(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("content-type", "application/json")
	_, _ = resp.Write([]byte(swaggerJson))
}
