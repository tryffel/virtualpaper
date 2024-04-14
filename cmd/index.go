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

package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services"
	"tryffel.net/go/virtualpaper/storage"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Schedule indexing documents to search engine",
	Long: `Schedules indexing for all documents in the system. 

The server needs to run to do index the documents.
This command will only mark the documents for scheduling. 
If server is currently running in the background, it should start processing the request shortly.
Otherwise server can be started after calling this command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		db, err := storage.NewDatabase(config.C.Database)
		if err != nil {
			logrus.Fatalf("Connect to database: %v", err)
		}
		defer db.Close()

		users, err := db.UserStore.GetUsers()
		if err != nil {
			logrus.Errorf("get users: %v", err)
			return
		}

		failed := 0
		success := 0

		service := services.NewAdminService(db, nil, nil)
		for _, v := range *users {
			logrus.Infof("index documents for user %s", v.Name)
			err = service.ForceProcessingByUser(cmd.Context(), v.Id, []models.ProcessStep{models.ProcessFts})
			if err != nil {
				logrus.Errorf("schedule indexing for user: (id: %d - %s): %v", v.Id, v.Name, err)
				failed += 1
			} else {
				success += 1
			}
		}

		if failed == 0 {
			logrus.Infof("Successfully scheduled indexing for %d users", success)
		} else {
			logrus.Errorf("Scheduling indexing failed for %d users, see above for errors", failed)
			os.Exit(1)
		}
	},
}
