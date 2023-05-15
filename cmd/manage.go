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
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/services/mail"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage server and users",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var mailTestCmd = &cobra.Command{
	Use:   "test-mail",
	Short: "Test sending mail",
	Long: "Send test mail. Default recipient is config.mail.error_recipient, " +
		"which can be overridden with flag --recipient",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		err := config.InitLogging()
		if err != nil {
			logrus.Fatalf("init log: %v", err)
			return
		}
		defer config.DeinitLogging()

		var recipient string
		if testMailRecipient != "" {
			recipient = testMailRecipient
		} else {
			recipient = config.C.Mail.ErrorRecipient
		}

		err = mail.SendMail("A test mail from Virtualpaper", "Mail configuration seemd to be valid.",
			config.C.Mail.ErrorRecipient)
		if err == nil {
			fmt.Printf("Sent mail to %s\n", recipient)
		} else {
			logrus.Fatalf("Mail send failed: %v", err)
		}
	},
}

var testMailRecipient string

func init() {
	manageCmd.AddCommand(mailTestCmd)
	mailTestCmd.PersistentFlags().StringVarP(&testMailRecipient, "recipient", "r", "",
		"Recipient to send test mail to")
}
