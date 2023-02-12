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
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/process"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

type Api struct {
	echo          *echo.Echo
	publicRouter  *echo.Group
	apiRouter     *echo.Group
	privateRouter *echo.Group
	adminRouter   *echo.Group

	cors    http.Handler
	db      *storage.Database
	search  *search.Engine
	process *process.Manager
}

// NewApi initializes new api instance. It connects to database and opens http port.
func NewApi(database *storage.Database) (*Api, error) {
	api := &Api{
		db:   database,
		echo: echo.New(),
	}

	api.echo.Use(middleware.RequestID())
	api.echo.Use(loggingMiddlware())
	api.echo.Use(middleware.Recover())
	api.echo.Use(middleware.CORS())
	api.echo.HTTPErrorHandler = httpErrorHandler

	api.echo.Server.ReadTimeout = time.Second * 30
	api.echo.Server.WriteTimeout = time.Second * 30

	var err error
	api.search, err = search.NewEngine(database, &config.C.Meilisearch)
	if err != nil {
		return api, err
	}

	api.process, err = process.NewManager(database, api.search)
	if err != nil {
		return api, err
	}

	api.addRoutesV2()
	return api, err
}

func (a *Api) Serve() error {
	err := a.process.Start()
	if err != nil {
		return err
	}

	go func() {
		addr := fmt.Sprintf("%s:%d", config.C.Api.Host, config.C.Api.Port)
		logrus.Infof("listen http on %s", addr)
		err := a.echo.Start(addr)
		if err != nil && err != http.ErrServerClosed {
			a.echo.Logger.Fatal("shutting down")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logrus.Info("stop server")
	if err := a.echo.Shutdown(ctx); err != nil {
		a.echo.Logger.Fatal(err)
	}
	logrus.Info("server stopped")
	return nil
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

func (a *Api) getVersionV2(c echo.Context) error {
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
	return c.JSON(http.StatusOK, v)
}

func (a *Api) getSupportedFileTypes(c echo.Context) error {
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
	return c.JSON(http.StatusOK, mimes)
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	govalidator.TagMap["metadata"] = govalidator.Validator(func(str string) bool {
		return !(strings.Contains(str, ";") || strings.Contains(str, ":") || strings.Contains(str, "\n"))
	})
	govalidator.TagMap["safefilename"] = govalidator.Validator(func(str string) bool {
		return str == govalidator.SafeFileName(str)
	})
}

//go:embed swaggerdocs/swagger.json
var swaggerJson string

func serverSwaggerDoc(c echo.Context) error {
	return c.File(swaggerJson)
}
