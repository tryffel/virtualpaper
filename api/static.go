package api

import (
	"github.com/sirupsen/logrus"
	"io/fs"
	"net/http"
	"tryffel.net/go/virtualpaper/frontend"
)

func staticServer() http.Handler {
	htmlContent, err := fs.Sub(frontend.StaticFiles, "dist")
	if err != nil {
		logrus.Panic(err)
	}
	return http.FileServer(http.FS(htmlContent))
}

func static() fs.FS {
	htmlContent, err := fs.Sub(frontend.StaticFiles, "dist")
	if err != nil {
		logrus.Panic(err)
	}
	return htmlContent
}
