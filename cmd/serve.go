package cmd

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/process"
	"tryffel.net/go/virtualpaper/storage"
	"tryffel.net/go/virtualpaper/storage/migration"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run server",
	Long: "Run Virtualpaper in server mode. Open http server and serve api as well as frontend " +
		"and start processing new documents. By default, migrate database to new version (if update exists).",

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

		schemaErr, _ := checkCorrectSchemaVersion(db)
		if schemaErr != nil {
			if migrationNeeded && !noAutoMigrateDb {
				logrus.Warningf("Start migrating database")
				err := migration.Migrate(db.Engine(), migration.Migrations)
				if err != nil {
					logrus.Fatalf("database migrations failed: %v", err)
				}

			} else {
				logrus.Fatalf("check database version: %v", schemaErr)
			}
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

var noAutoMigrateDb = false
var migrationNeeded = false

func init() {
	serveCmd.PersistentFlags().BoolVarP(&noAutoMigrateDb, "no-migrate", "m", false,
		"Disable automatic database migrations on startup")
}

func checkCorrectSchemaVersion(db *storage.Database) (err error, current int) {
	logrus.Debugf("check database version")

	var version migration.Schema

	version, err = migration.CurrentVersion(db.Engine())
	if err != nil {
		return
	}

	current = version.Level

	if version.Level == 0 && version.Success == 0 {
		err = errors.New("database needs initializing. Please run 'migrate' first.")
		migrationNeeded = true
		return
	}

	if version.Success == 0 {
		err = errors.New("last migration has failed")
		return
	}

	if version.Level == config.SchemaVersion {
		err = nil
		logrus.Debugf("database version ok")
		return
	}

	if version.Level > config.SchemaVersion {
		err = fmt.Errorf("database schema version unsupported: v%d, supported: v%d",
			version.Level, config.SchemaVersion)
		return
	}

	if version.Level < config.SchemaVersion {
		err = fmt.Errorf("database schema needs migrating to version %d", config.SchemaVersion)
		migrationNeeded = true
		return
	}
	err = nil
	return
}
