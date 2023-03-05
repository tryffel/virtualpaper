package mail

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
)

func ResetPassword(email string, token string, tokenId int) {
	textFmt := `Password reset for Virtualpaper

Someone requested resetting password to Virtualpaper with this email address.
To reset your password click the link: %s
If you did not request password reset link, no further actions are required.
`

	link := fmt.Sprintf("%s/#/auth/reset-password?token=%s&id=%d", config.C.Api.PublicUrl, token, tokenId)

	//logrus.Debugf("password reset link: %s", link)
	text := fmt.Sprintf(textFmt, link)

	err := SendMail("Password reset link for Virtualpaper", text, email)
	if err != nil {
		logrus.Errorf("send password reset email for user %s: %v", email, err)
	}
	logrus.Infof("password reset link sent for email %s", email)
}
