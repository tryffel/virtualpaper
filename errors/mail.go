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

package errors

import (
	"context"
	e "errors"
	"fmt"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/services/mail"
)

// MailEnabled returns true if error mailing is enabled
func MailEnabled() bool {
	return mail.MailEnabled() && config.C.Mail.ErrorRecipient != ""
}

// SendMail sends error as mail to recipient. If recipient == "", send mail
// to default recipient from config file.
func SendMail(ctx context.Context, err error, recipient string) error {
	if recipient == "" {
		recipient = config.C.Mail.ErrorRecipient
		if recipient == "" {
			return e.New("no default error recipient defined")
		}
	}

	msg := fmt.Sprintf("Virtualpaper (version %v) caught an error.\nTimestamp: %s\n",
		config.Version, time.Now().String())
	if vpErr, ok := err.(Error); ok {
		if len(vpErr.Stack) == 0 {
			vpErr.SetStack()
		}
		msg += fmt.Sprintf("error type: %s\nerror: %s\nmessage: %s\n\nStack trace: \n%s",
			vpErr.ErrType, vpErr.Err.Error(), vpErr.ErrMsg, string(vpErr.Stack))
	} else {
		stack := getStack(8)
		msg += fmt.Sprintf("uncaught error: %s\nstack:\n\n%s", err.Error(), stack)
	}
	return mail.SendMail(ctx, "Caught an error in Virtualpaper", msg, recipient)
}
