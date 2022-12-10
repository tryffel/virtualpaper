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
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/mail"
)

const (
	tokenClaimUserid = "user_id"
)

func (a *Api) authorizeUserV2() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (authErr error) {
			authErr = echo.ErrUnauthorized

			key := c.Request().Header.Get("Authorization")
			if key == "" {
				return
			}

			parts := strings.Split(key, " ")
			if len(parts) != 2 {
				return
			}

			userId, err := validateToken(parts[1], config.C.Api.Key)
			if userId == "" || err != nil {
				return
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
					Context: Context{c},
					Admin:   user.IsAdmin,
					UserId:  numId,
					User:    user,
				}
				return next(ctx)
			}
			return
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
func newToken(userId string, privateKey string) (string, error) {
	var token *jwt.Token = nil

	claims := jwt.MapClaims{
		tokenClaimUserid: userId,
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

// validateToken validates and parses user id. If valid, return user_id, else return error description
// Return user_id, nonce, error
func validateToken(tokenString string, privateKey string) (string, error) {
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
				return "", e
			}
			if e.Inner.Error() == "Token is expired" {
				logrus.Debugf("token expired")
				e := errors.ErrInvalid
				e.ErrMsg = "token expired"
				return "", e
			}

		} else {
			e := errors.ErrInvalid
			e.ErrMsg = "invalid token"
			return "", e
		}
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		user := claims[tokenClaimUserid].(string)
		expired := !claims.VerifyNotBefore(time.Now().Unix(), config.C.Api.TokenExpire != 0)
		if expired {
			logrus.Debugf("token expired")
			e := errors.ErrInvalid
			e.ErrMsg = "token expired"
			return "", e
		}
		return user, nil
	}
	return "", fmt.Errorf("invalid token")
}

type LoginRequest struct {
	Username string `valid:"alphanum,required"`
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

	dto.Username = strings.ToLower(dto.Username)
	userId, err := a.db.UserStore.TryLogin(dto.Username, dto.Password)
	remoteAddr := getRemoteAddr(req)
	if userId == -1 || err != nil {
		logrus.Infof("Failed login attempt for user %s from remote %s", dto.Username, remoteAddr)
		return echo.ErrUnauthorized
	}

	c.Logger().Infof("User %d '%s' logged in from %s", userId, dto.Username, remoteAddr)
	token, err = newToken(strconv.Itoa(userId), config.C.Api.Key)
	if err != nil {
		c.Logger().Errorf("Create new token: %v", err)
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
