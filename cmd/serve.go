package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"tryffel.net/go/virtualpaper/api"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		server, err := api.NewApi()
		if err != nil {
			logrus.Errorf("init server: %v", err)
			return
		}

		server.Serve()

	},
}
