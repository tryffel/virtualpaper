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

package mail

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"time"
	"tryffel.net/go/virtualpaper/config"
	log "tryffel.net/go/virtualpaper/util/logger"
)

// SendMail sends mail.
func SendMail(ctx context.Context, subject, msg string, recipient string) error {
	if !MailEnabled() {
		return fmt.Errorf("mail not configured")
	}

	log.Context(ctx).WithField("recipient", recipient).WithField("subject", subject).Infof("Send mail")
	start := time.Now()

	host := fmt.Sprintf("%s:%d", config.C.Mail.Host, config.C.Mail.Port)
	auth := smtp.PlainAuth("", config.C.Mail.Username, config.C.Mail.Password, config.C.Mail.Host)
	fullMsg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", recipient, subject, msg)
	err := smtp.SendMail(host, auth, config.C.Mail.From, []string{recipient}, []byte(fullMsg))
	took := time.Now().Sub(start)

	millisec := took.Milliseconds()
	if millisec > 2000 {
		log.Context(ctx).Warnf("Sending mail took %.2f s", float32(millisec)/1000)
	}

	if err == nil {
		return nil
	}
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("something went wrong when sending mail, please check mail config is valid: %v", err)
	}
	if err != nil {
		return fmt.Errorf("smtp: %v", err)
	}
	return nil
}

// MailEnabled returns true if error mailing is enabled
func MailEnabled() bool {
	return config.C.Mail.Host != ""
}
