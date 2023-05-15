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
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index documents to meilisearch for full-text-search",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		db, err := storage.NewDatabase(config.C.Database)
		if err != nil {
			logrus.Fatalf("Connect to database: %v", err)
		}
		defer db.Close()

		engine, err := search.NewEngine(db, &config.C.Meilisearch)
		if err != nil {
			logrus.Fatalf("Init search engine: %v", err)
		}

		users, err := db.UserStore.GetUsers()
		if err != nil {
			logrus.Errorf("get users: %v", err)
			return
		}

		paging := storage.Paging{
			Offset: 0,
			Limit:  200,
		}

		for _, v := range *users {
			docs, err := db.DocumentStore.GetNeedsIndexing(v.Id, paging)
			if err != nil {
				logrus.Warningf("get indexing documents got user: %v", err)
				continue
			}

			logrus.Infof("index %d documents for useer %s", len(*docs), v.Name)
			err = engine.IndexDocuments(docs, v.Id)
			if err != nil {
				logrus.Warningf("index documents: %v", err)
			}
		}
	},
}
