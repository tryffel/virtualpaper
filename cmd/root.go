package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
	"tryffel.net/go/virtualpaper/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "virtualpaper",
	Short: "Virtualpaper",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(manageCmd)
	rootCmd.AddCommand(indexCmd)
}

func initConfig() {

	logrus.SetLevel(logrus.DebugLevel)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {

		configDir, err := os.UserConfigDir()
		if err != nil {
			logrus.Errorf("Cannot determine config directory: %v", err)
			configDir, err = homedir.Dir()
			if err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
		}
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config.toml")
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvPrefix("virtualpaper")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("Read config file: %v", err)
	}
	err := config.ConfigFromViper()
	if err != nil {
		logrus.Fatalf("Read config file: %v", err)
	}

	err = config.InitConfig()
	if err != nil {
		logrus.Fatalf("Init config: %v", err)
	}
}
