package services

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services/mail"
	"tryffel.net/go/virtualpaper/storage"
	"tryffel.net/go/virtualpaper/util/logger"
)

type AuthService struct {
	db *storage.Database
}

func NewAuthService(db *storage.Database) *AuthService {
	return &AuthService{
		db: db,
	}
}

func (service *AuthService) Login(ctx context.Context, username, password, userAgent, ipAddrs string) (*models.Token, error) {
	userId, err := service.db.UserStore.TryLogin(username, password)
	if err != nil {
		return nil, err
	}
	authToken := &models.Token{
		Id:            0,
		UserId:        userId,
		IpAddr:        ipAddrs,
		Name:          userAgent,
		LastConfirmed: time.Now(),
		LastSeen:      time.Now(),
	}
	if config.C.Api.TokenExpireSec != 0 {
		authToken.ExpiresAt = time.Now().Add(config.C.Api.TokenExpire)
	}
	err = authToken.Init()
	if err != nil {
		return nil, fmt.Errorf("init token: %v", err)
	}
	err = service.db.AuthStore.InsertToken(authToken)
	if err != nil {
		return nil, fmt.Errorf("persist auth token: %v", err)
	}

	logger.Entry(ctx).WithField("user", username).WithField("remoteAddr", ipAddrs).Infof("User logged in")
	user, err := service.db.UserStore.GetUser(userId)
	if err != nil {
		return nil, fmt.Errorf("get user: %v", err)
	}
	if user.Email != "" {
		logger.Context(ctx).Infof("Send email for logged in user to %s", user.Email)

		msg := fmt.Sprintf(`User logged in

ip address: %s,
user agent: %s,
`, ipAddrs, userAgent)

		err := mail.SendMail(ctx, "User logged in", msg, user.Email)
		if err != nil {
			logger.Context(ctx).Errorf("send logged-in email to user: %s: %v", user.Email, err)
		}
	}
	return authToken, nil
}

func (service *AuthService) GetUserByToken(ctx context.Context, tokenKey string, userId int) (user *models.User, token *models.Token, tokenError error) {
	token, err := service.db.AuthStore.GetToken(tokenKey, true)
	if err != nil {
		tokenError = err
		return
	}

	if token.HasExpired() {
		tokenError = errors.ErrUnauthorized
		return
	}

	user, err = service.db.UserStore.GetUser(userId)
	if err != nil {
		tokenError = err
		return
	}

	if !user.IsActive {
		tokenError = errors.ErrInvalid
		return
	}
	return
}

func (service *AuthService) ConfirmAuthToken(ctx context.Context, tokenKey string) error {
	token, err := service.db.AuthStore.GetToken(tokenKey, false)
	if err != nil {
		return err
	}

	if token.ConfirmationExpired() {
		logger.Context(ctx).WithField("token", token.Id).Infof("auth token needs confirmation")
		return errors.ErrInvalid
	}
	return nil
}

func (service *AuthService) ResetPassword(ctx context.Context, request *PasswordReset) error {
	token, err := service.db.UserStore.GetPasswordResetTokenByHash(request.Id)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.ErrForbidden
		}
		return err
	}

	if token.HasExpired() {
		logger.Context(ctx).Warnf("user %d attempted to change password with expired reset token %d, expired at %s",
			token.UserId, token.Id, token.ExpiresAt)
		e := errors.ErrInvalid
		e.ErrMsg = "Token has expired. Please create a new reset link."
		return e
	}

	match, err := token.TokenMatches(request.Token)
	if err != nil {
		logger.Context(ctx).Warnf("user %d attempted to change password with bad reset token %d: %v", token.UserId, token.Id, err)
		return fmt.Errorf("compare token to hash: %v", err)
	}
	if !match {
		e := errors.ErrForbidden
		e.ErrMsg = "Invalid token"
		return e
	}

	user, err := service.db.UserStore.GetUser(token.UserId)
	if err != nil {
		logger.Context(ctx).Errorf("user %d not found for password reset token %d", token.UserId, token.Id)
	}

	if user != nil {
		err = user.SetPassword(request.Password)
		if err != nil {
			return fmt.Errorf("set new password: %v", err)
		}

		err = service.db.UserStore.Update(user)
		if err != nil {
			return fmt.Errorf("update user's passowrd: %v", err)
		}
	}

	logger.Context(ctx).Infof("Reset user's (%d) password with reset token %d", user.Id, token.Id)
	err = service.db.UserStore.DeletePasswordResetToken(token.Id)
	if err != nil {
		logger.Context(ctx).Errorf("delete password token %d: %v", token.Id, err)
	}
	return nil
}

func (service *AuthService) RevokeToken(ctx context.Context, tokenKey string) error {
	return service.db.AuthStore.RevokeToken(tokenKey)
}

func (service *AuthService) CreateResetPasswordToken(ctx context.Context, email string) error {
	user, err := service.db.UserStore.GetUserByEmail(email)
	userOk := user != nil && err == nil
	if err != nil {
		logger.Context(ctx).Warnf("user by email '%s' not found when creating reset password link", email)
	}

	// don't allow inactive users to reset passwords
	if user != nil && userOk {
		if !user.IsActive {
			logger.Context(ctx).Warnf("user %d is not active, refuse to send password reset link", user.Id)
			userOk = false
		}
		if user.Email == "" {
			logger.Context(ctx).Warnf("user %d does not have valid email, cannot send password reset link", user.Id)
			userOk = false
		}
	}

	token := &models.PasswordResetToken{}
	rawToken, hash, err := newPasswordToken()
	if err != nil {
		return fmt.Errorf("generate random string for password reset token: %v", err)
	}
	token.Token = hash

	if !userOk {
		// about the time it would take to save the token in db
		time.Sleep(time.Millisecond * 2)
		return nil
	}

	// token is valid for 24 hours
	token.ExpiresAt = time.Now().Add(time.Hour * 24)
	token.Update()
	token.CreatedAt = token.UpdatedAt

	token.UserId = user.Id
	err = service.db.UserStore.AddPasswordResetToken(token)
	if err != nil {
		return fmt.Errorf("save password reset token: %v", err)
	}
	logger.Context(ctx).Warnf("Create password reset token %d for user %d, expires at %s", token.Id, user.Id, token.ExpiresAt)
	go mail.ResetPassword(ctx, user.Email, rawToken, token.Id)
	return nil
}

func (service *AuthService) ConfirmAuthentication(ctx context.Context, user *models.User, password, remoteAddr, tokenKey string) error {
	userId, err := service.db.UserStore.TryLogin(user.Name, password)
	if userId == -1 || err != nil {
		logger.Context(ctx).Infof("Failed authentication confirmation for user %d, token %s, from remote %s", user.Id, tokenKey, remoteAddr)
		return errors.ErrForbidden
	}

	token, err := service.db.AuthStore.GetToken(tokenKey, true)
	if err != nil {
		return err
	}

	token.LastConfirmed = time.Now()
	err = service.db.AuthStore.UpdateTokenConfirmation(tokenKey, token.LastConfirmed)
	if err != nil {
		return err
	}
	logger.Context(ctx).Infof("Authentication confirmation successful for user %d, token %s, from remote %s", user.Id, tokenKey, remoteAddr)
	return nil
}

type PasswordReset struct {
	Token    string
	Id       int
	Password string
}

// get new password token, returns token, hashed token and error
func newPasswordToken() (string, string, error) {
	rawToken, err := config.RandomStringCrypt(70)
	if err != nil {
		return "", "", fmt.Errorf("generate random token: %v", err)
	}
	hash, err := hashPasswordResetToken(rawToken)
	if err != nil {
		return "", "", fmt.Errorf("hash token: %v", err)
	}
	return rawToken, hash, nil
}

func hashPasswordResetToken(token string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(token), 14)
	if err != nil {
		return "", fmt.Errorf("hash token: %v", err)
	}
	return string(bytes), nil
}
