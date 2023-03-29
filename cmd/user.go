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
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"tryffel.net/go/virtualpaper/api"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

// ReadUserInput reads value from stdin. Name is printed like 'Enter <name>. If mask is true, input is masked.
func readUserInput(name string, mask bool) (string, error) {
	fmt.Print("Enter ", name, ": ")
	var val string
	var err error
	if mask {
		// needs cast for windows
		raw, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", fmt.Errorf("failed to read user input: %v", err)
		}
		val = string(raw)
		fmt.Println()
	} else {
		reader := bufio.NewReader(os.Stdin)
		val, err = reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read user input: %v", err)
		}
	}
	val = strings.Trim(val, "\n\r")
	return val, nil
}

var addUserCmd = &cobra.Command{
	Use:   "add-user",
	Short: "Add new user. Enter username, password and whether to make user administrator.",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		err := config.InitLogging()
		if err != nil {
			logrus.Fatalf("init log: %v", err)
			return
		}
		defer config.DeinitLogging()
		db, err := storage.NewDatabase(config.C.Database)
		if err != nil {
			logrus.Fatalf("Connect to database: %v", err)
		}
		defer db.Close()

		if userName == "" {
			userName, err = readUserInput("username", false)
			if userName == "" {
				logrus.Fatalf("username cannot be empty")
			}
		}

		if password == "" {
			var firstPw string
			var secondPw string
			firstPw, err = readUserInput("password", true)
			secondPw, err = readUserInput("repeat password", true)
			if firstPw != secondPw {
				logrus.Fatalf("passwords do not match.")
			}
			err = api.ValidatePassword(firstPw)
			if err != nil {
				logrus.Fatalf("error: %v", err)
			}

			password = firstPw
		}

		var admin bool
		if strings.ToLower(optAdmin) == "true" {
			admin = true
		} else if strings.ToLower(optAdmin) == "false" {
			admin = false
		} else {
			for {
				isAdmin, err := readUserInput("user is administrator (y/n)", false)
				if err != nil {
					logrus.Errorf("error reading input: %v", err)
				}
				if isAdmin == "" || isAdmin == "n" {
					admin = false
					break
				} else if isAdmin == "y" {
					admin = true
					break
				} else {
					logrus.Errorf("enter either y or n")
				}
			}
		}

		user := &models.User{}
		user.Name = userName
		err = user.SetPassword(password)
		if err != nil {
			logrus.Errorf("set password: %v", err)
		}
		user.IsAdmin = admin
		user.IsActive = true

		err = db.UserStore.AddUser(user)
		if err != nil {
			logrus.Error(err)
		} else {
			if admin {
				logrus.Infof("Created admin user (id:%d) - %s", user.Id, user.Name)
			} else {
				logrus.Infof("Created user (id:%d) %s", user.Id, user.Name)
			}

		}

		logrus.Infof("init search engine index for new user")
		_, err = search.NewEngine(db, &config.C.Meilisearch)
		if err != nil {
			logrus.Fatalf("connect to search engine: %v", err)
		}
		logrus.Infof("Done")
	},
}

var resetPwCMd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset user password",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		err := config.InitLogging()
		if err != nil {
			logrus.Fatalf("init log: %v", err)
			return
		}
		defer config.DeinitLogging()

		db, err := storage.NewDatabase(config.C.Database)
		if err != nil {
			logrus.Fatalf("Connect to database: %v", err)
		}
		defer db.Close()
		userName, err := readUserInput("username", false)
		if userName == "" {
			logrus.Fatalf("username cannot be empty")
		}
		firstPw, err := readUserInput("new password", true)
		secondPw, err := readUserInput("repeat password", true)
		if firstPw != secondPw {
			logrus.Fatalf("passwords do not match.")
		}
		err = api.ValidatePassword(firstPw)
		if err != nil {
			logrus.Fatalf("error: %v", err)
		}

		user, err := db.UserStore.GetUserByName(userName)
		if err != nil {
			logrus.Fatalf("user not found: %v", err)
		}

		err = user.SetPassword(firstPw)
		if err != nil {
			logrus.Fatalf("set new password: %v", err)
		}

		err = db.UserStore.Update(user)
		if err != nil {
			logrus.Fatalf("update user: %v", err)
		}
		logrus.Infof("Password updated")
	},
}

var userName = ""
var password = ""
var optAdmin = ""

func init() {
	manageCmd.AddCommand(addUserCmd)
	manageCmd.AddCommand(resetPwCMd)

	addUserCmd.PersistentFlags().StringVarP(&optAdmin, "admin", "a", "",
		"Make user an administrator")
	addUserCmd.PersistentFlags().StringVarP(&userName, "username", "U", "",
		"New username")
	addUserCmd.PersistentFlags().StringVarP(&password, "password", "P", "",
		"New password")
}

func init() {

}
