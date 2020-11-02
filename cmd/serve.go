package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/process"
	"tryffel.net/go/virtualpaper/storage"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := storage.NewDatabase()
		if err != nil {
			logrus.Errorf("connect to database: %v", err)
			return
		}
		server, err := api.NewApi(db)
		if err != nil {
			logrus.Errorf("init server: %v", err)
			return
		}

		process.Init()
		defer process.Deinit()

		err = server.Serve()
		if err != nil {
			logrus.Errorf("start server: %v", err)
		}
	},
}
