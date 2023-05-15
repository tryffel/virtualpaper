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

package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/mileusna/useragent"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/mail"
	"tryffel.net/go/virtualpaper/models"
)

const (
	tokenClaimUserid = "user_id"
	tokenClaimsId    = "token_id"
)

func (a *Api) authorizeUserV2() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (authErr error) {
			e := errors.ErrUnauthorized
			e.ErrMsg = "invalid token"
			authErr = e

			key := c.Request().Header.Get("Authorization")
			if key == "" {
				return
			}

			parts := strings.Split(key, " ")
			if len(parts) != 2 {
				return
			}

			userId, tokenKey, err := validateToken(parts[1], config.C.Api.Key)
			if userId == "" || err != nil {
				return
			}

			token, err := a.db.AuthStore.GetToken(tokenKey, true)
			if err != nil {
				if errors.Is(err, errors.ErrRecordNotFound) {
					return authErr
				}
				return fmt.Errorf("get token from database: %v", err)
			}

			if token.HasExpired() {
				return authErr
			}

			if userId != "" {
				numId, err := strconv.Atoi(userId)
				if err != nil {
					c.Logger().Error("user id is not numerical", numId)
					return echo.ErrInternalServerError
				}

				user, err := a.db.UserStore.GetUser(numId)
				if err != nil {
					return
				}

				if !user.IsActive {
					return
				}

				ctx := UserContext{
					Context:  Context{c},
					Admin:    user.IsAdmin,
					UserId:   numId,
					User:     user,
					TokenKey: token.Key,
				}
				return next(ctx)
			}
			return
		}
	}
}

func (a *Api) ConfirmAuthorizedToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, ok := c.(UserContext)
			if !ok {
				c.Logger().Error("no UserContext found")
				return echo.ErrInternalServerError
			}
			token, err := a.db.AuthStore.GetToken(ctx.TokenKey, false)
			if err != nil {
				return err
			}
			if token.ConfirmationExpired() {
				logrus.Infof("user's token needs confirmation, token %d", token.Id)
				err := errors.ErrUnauthorized
				err.ErrMsg = "authentication required"
				return err
			}
			return next(c)
		}
	}
}

func (a *Api) corsHeader(next http.Handler) http.Handler {
	hosts := config.C.Api.CorsHostList()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Access-Control-Allow-Origin", hosts)
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		next.ServeHTTP(w, r)
	})
}

// newToken issues a new token for user_id
// if ExpireDuration == 0, disable expiration
func newToken(userId string, tokenId string, privateKey string) (string, error) {
	var token *jwt.Token = nil

	claims := jwt.MapClaims{
		tokenClaimUserid: userId,
		tokenClaimsId:    tokenId,
	}

	if config.C.Api.TokenExpire != 0 {
		claims["nbf"] = time.Now().Unix()
		claims["exp"] = time.Now().Add(config.C.Api.TokenExpire).Unix()
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(privateKey))
	if err != nil {
		return tokenString, err
	}
	return tokenString, nil
}

// validateToken validates and parses user id and token id. If valid, return user_id, else return error description
// Return user_id, nonce, error
func validateToken(tokenString string, privateKey string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(privateKey), nil
	})

	if err != nil {
		e, ok := err.(*jwt.ValidationError)
		if ok {
			if e.Inner == nil {
				e := errors.ErrInvalid
				e.ErrMsg = "invalid token"
				return "", "", e
			}
			if e.Inner.Error() == "Token is expired" {
				logrus.Debugf("token expired")
				e := errors.ErrInvalid
				e.ErrMsg = "token expired"
				return "", "", e
			}

		} else {
			e := errors.ErrInvalid
			e.ErrMsg = "invalid token"
			return "", "", e
		}
		return "", "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		user := claims[tokenClaimUserid].(string)
		rawToken := claims[tokenClaimsId]
		if rawToken == nil {
			e := errors.ErrInvalid
			e.ErrMsg = "invalid token"
			return "", "", e
		}
		expired := !claims.VerifyNotBefore(time.Now().Unix(), config.C.Api.TokenExpire != 0)
		if expired {
			logrus.Debugf("token expired")
			e := errors.ErrInvalid
			e.ErrMsg = "token expired"
			return "", "", e
		}
		return user, rawToken.(string), nil
	}
	return "", "", fmt.Errorf("invalid token")
}

type LoginRequest struct {
	Username string `valid:"username,required"`
	Password string `valid:"required"`
}

type LoginResponse struct {
	UserId int
	Token  string
}

func (a *Api) LoginV2(c echo.Context) error {
	// swagger:route POST /api/v1/version Authentication Login
	// Login
	//
	// responses:
	//   200:

	req := c.Request()
	headers := req.Header
	token := headers.Get("Authorization")
	if len(token) > 0 {
		return c.String(http.StatusNotModified, "already logged in")
	}

	dto := &LoginRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	userId, err := a.db.UserStore.TryLogin(dto.Username, dto.Password)
	remoteAddr := getRemoteAddr(req)
	if userId == -1 || err != nil {
		logrus.Infof("Failed login attempt for user %s from remote %s", dto.Username, remoteAddr)
		// request takes about the same time with invalid password & invalid user
		time.Sleep(time.Millisecond*1100 + time.Duration(int(rand.Float64()*1000))*time.Millisecond)
		return echo.ErrUnauthorized
	}

	c.Logger().Infof("User %d '%s' logged in from %s", userId, dto.Username, remoteAddr)
	authToken := &models.Token{
		Id:            0,
		UserId:        userId,
		IpAddr:        c.RealIP(),
		LastConfirmed: time.Now(),
		LastSeen:      time.Now(),
	}

	ua := useragent.Parse(c.Request().Header.Get("User-Agent"))
	authToken.Name = fmt.Sprintf("%s %s, %s %s", ua.OS, ua.OSVersion, ua.Name, ua.Version)

	if config.C.Api.TokenExpireSec != 0 {
		authToken.ExpiresAt = time.Now().Add(config.C.Api.TokenExpire)
	}

	err = authToken.Init()
	if err != nil {
		return fmt.Errorf("init token: %v", err)
	}
	err = a.db.AuthStore.InsertToken(authToken)
	if err != nil {
		return fmt.Errorf("save auth token to database: %v", err)
	}

	token, err = newToken(strconv.Itoa(userId), authToken.Key, config.C.Api.Key)
	if err != nil {
		c.Logger().Errorf("CreateJob new token: %v", err)
	}

	user, err := a.db.UserStore.GetUser(userId)
	if user.Email != "" {
		logrus.Debugf("Send email for logged in user to %s", user.Email)

		msg := fmt.Sprintf(`User logged in

ip address: %s,
user agent: %s,
`, remoteAddr, req.Header.Get("user-agent"))

		err := mail.SendMail("User logged in", msg, user.Email)
		if err != nil {
			logrus.Errorf("send logged-in email to user: %s: %v", user.Email, err)
		}
	}
	respBody := &LoginResponse{
		UserId: userId,
		Token:  token,
	}

	return c.JSON(http.StatusOK, respBody)
}

type ResetPasswordRequest struct {
	Token    string `json:"token" valid:"minstringlength(4)"`
	Id       int    `json:"id" valid:"required"`
	Password string `json:"password" valid:"required"`
}

func (a *Api) Logout(c echo.Context) error {
	ctx, ok := c.(UserContext)
	if !ok {
		// should not happen, user is required to authenticate
		return c.JSON(200, map[string]string{"status": "ok"})
	}
	err := a.db.AuthStore.RevokeToken(ctx.TokenKey)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return c.JSON(200, map[string]string{"status": "ok"})
		}
		return fmt.Errorf("delete token from db: %v", err)
	}
	return c.JSON(200, map[string]string{"status": "ok"})
}

func (a *Api) ResetPassword(c echo.Context) error {
	// swagger:route POST /api/v1/auth/reset-password Authentication Reset password
	// ResetPassword
	//
	// responses:
	//   200:

	req := c.Request()
	dto := &ResetPasswordRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err = ValidatePassword(dto.Password); err != nil {
		return err
	}

	token, err := a.db.UserStore.GetPasswordResetTokenByHash(dto.Id)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			e := errors.ErrForbidden
			logrus.Warningf("password reset token not found by id %d", dto.Id)
			return e
		}
		return err
	}

	if token.HasExpired() {
		logrus.Warningf("user %d attempted to change password with expired reset token %d, expired at %s",
			token.UserId, token.Id, token.ExpiresAt)
		e := errors.ErrInvalid
		e.ErrMsg = "Token has expired. Please create a new reset link."
		return e
	}

	match, err := token.TokenMatches(dto.Token)
	if err != nil {
		logrus.Warningf("user %d attempted to change password with bad reset token %d: %v", token.UserId, token.Id, err)
		return fmt.Errorf("compare token to hash: %v", err)
	}
	if !match {
		e := errors.ErrForbidden
		e.ErrMsg = "Invalid token"
		return e
	}

	user, err := a.db.UserStore.GetUser(token.UserId)
	if err != nil {
		logrus.Errorf("user %d not found for password reset token %d", token.UserId, token.Id)
	}

	if user != nil {
		err = user.SetPassword(dto.Password)
		if err != nil {
			return fmt.Errorf("set new password: %v", err)
		}

		err = a.db.UserStore.Update(user)
		if err != nil {
			return fmt.Errorf("update user's passowrd: %v", err)
		}
	}

	logrus.Warningf("Reset user's (%d) password with reset token %d", user.Id, token.Id)

	logrus.Warningf("CreateJob password reset token %d for user %d, expires at %s", token.Id, user.Id, token.ExpiresAt)
	err = a.db.UserStore.DeletePasswordResetToken(token.Id)
	if err != nil {
		logrus.Errorf("delete password token %d: %v", token.Id, err)
	}
	return c.JSON(200, "ok")
}

type ForgottenPasswordRequest struct {
	Email string `json:"email" valid:"email"`
}

func (a *Api) CreateResetPasswordToken(c echo.Context) error {
	// swagger:route POST /api/v1/auth/reset-password-token Authentication Forgot password
	// ForgotPassword
	//
	// responses:
	//   200:

	req := c.Request()
	dto := &ForgottenPasswordRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = err.Error()
		e.Err = err
		return e
	}

	user, err := a.db.UserStore.GetUserByEmail(dto.Email)
	userOk := user != nil && err == nil
	if err != nil {
		logrus.Warningf("user by email '%s' not found when creating reset password link", dto.Email)
	}

	// don't allow inactive users to reset passwords
	if user != nil && userOk {
		if !user.IsActive {
			logrus.Warningf("user %d is not active, refure to send password reset link", user.Id)
			userOk = false
		}
		if user.Email == "" {
			logrus.Warningf("user %d does not have valid email, cannot send password reset link", user.Id)
			userOk = false
		}
	}

	token := &models.PasswordResetToken{}
	rawToken, hash, err := newPasswordToken()
	if err != nil {
		return fmt.Errorf("generate random string for password reset token: %v", err)
	}
	token.Token = hash

	msg := map[string]string{"status": "Password reset email has been sent"}

	if !userOk {
		// about the time it would take to save the token in db
		time.Sleep(time.Millisecond * 2)
		return c.JSON(200, msg)
	}

	// token is valid for 24 hours
	token.ExpiresAt = time.Now().Add(time.Hour * 24)
	token.Update()
	token.CreatedAt = token.UpdatedAt

	token.UserId = user.Id
	err = a.db.UserStore.AddPasswordResetToken(token)
	if err != nil {
		return fmt.Errorf("save password reset token: %v", err)
	}

	logrus.Warningf("CreateJob password reset token %d for user %d, expires at %s", token.Id, user.Id, token.ExpiresAt)
	go mail.ResetPassword(user.Email, rawToken, token.Id)
	return c.JSON(200, msg)
}

type AuthConfirmationRequest struct {
	Password string `json:"password" valid:"stringlength(8|150)"`
}

func (a *Api) ConfirmAuthentication(c echo.Context) error {
	req := c.Request()
	dto := &AuthConfirmationRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = err.Error()
		e.Err = err
		return e
	}
	user := c.(UserContext)
	userId, err := a.db.UserStore.TryLogin(user.User.Name, dto.Password)
	remoteAddr := getRemoteAddr(req)
	if userId == -1 || err != nil {
		logrus.Infof("Failed authentication confirmation for user %d, token %s, from remote %s", user.UserId, user.TokenKey, remoteAddr)
		return echo.ErrForbidden
	}

	token, err := a.db.AuthStore.GetToken(user.TokenKey, true)
	if err != nil {
		return err
	}

	token.LastConfirmed = time.Now()
	err = a.db.AuthStore.UpdateTokenConfirmation(user.TokenKey, token.LastConfirmed)
	if err != nil {
		return err
	}
	logrus.Infof("Authentication confirmation successful for user %d, token %s, from remote %s", user.UserId, user.TokenKey, remoteAddr)
	return c.JSON(200, "")
}

// get new password token, returns token, hashed token and error
func newPasswordToken() (string, string, error) {
	rawToken, err := config.RandomStringCrypt(80)
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

func ValidatePassword(password string) error {
	err := errors.ErrInvalid
	if len(password) < 8 || len(password) > 100 {
		err.ErrMsg = "password's length must be 8 - 100 characters"
		return err
	}
	return nil
}
