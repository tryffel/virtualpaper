package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/storage"
	"tryffel.net/go/virtualpaper/storage/migration"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run migrations or init empty application",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		db, err := storage.NewDatabase(config.C.Database)
		if err != nil {
			logrus.Fatalf("Connect to database: %v", err)
		}
		defer db.Close()

		currentSchema, err := migration.CurrentVersion(db.Engine())
		if err != nil {

			if strings.Contains(err.Error(), "relation \"schemas\" does not exist") {
				err = runMigration(db)
				if err != nil {
					logrus.Errorf("Init db: %v", err)
				}
				return
			}
			logrus.Fatalf("Get db schema level: %v", err)
		}

		if currentSchema.Level == config.SchemaVersion {
			logrus.Info("Database schema up to date")
			return
		}

		if currentSchema.Level > config.SchemaVersion {
			logrus.Fatalf("Database schema level too high! Supported level %d, but db already has %d",
				config.SchemaVersion, currentSchema.Level)
			return
		}

		if currentSchema.Level < config.SchemaVersion {
			err = runMigration(db)
			if err != nil {
				logrus.Errorf("Migrate: %v", err)
			}
		}
	},
}

func runMigration(db *storage.Database) error {
	return migration.Migrate(db.Engine(), migration.Migrations)
}
