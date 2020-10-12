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
		db, err := storage.NewDatabase()
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
			logrus.Info("Already up to date")
			return
		}

		if currentSchema.Level > config.SchemaVersion {
			logrus.Warningf("Schema level too high!")
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
