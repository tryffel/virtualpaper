package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"tryffel.net/go/virtualpaper/config"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of VirtualPaper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s - %s\n", config.Version, config.Commit)
	},
}
