package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/process"
	"tryffel.net/go/virtualpaper/storage"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		err := config.InitLogging()
		if err != nil {
			logrus.Fatalf("init log: %v", err)
			return
		}
		defer config.DeinitLogging()

		db, err := storage.NewDatabase()
		if err != nil {
			logrus.Fatalf("connect to database: %v", err)
			return
		}
		server, err := api.NewApi(db)
		if err != nil {
			logrus.Fatalf("init server: %v", err)
			return
		}

		process.Init()
		defer process.Deinit()

		err = server.Serve()
		if err != nil {
			logrus.Fatalf("start server: %v", err)
		}
	},
}
