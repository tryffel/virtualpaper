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
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
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

func (a *Api) authorizeUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		key := r.Header.Get("Authorization")
		if key == "" {
			var err error
			err = respUnauthorized(w)
			if err != nil {
				logrus.Errorf("send unauthorized to user: %v", err)
			}
			return
		}

		parts := strings.Split(key, " ")
		if len(parts) != 2 {
			respBadRequest(w, "Invalid token", nil)
			return
		}

		userId, err := validateToken(parts[1], config.C.Api.Key)
		if userId == "" || err != nil {
			respError(w, err, "authorize user middleware")
			return
		}

		if userId != "" {
			numId, err := strconv.Atoi(userId)
			if err != nil {
				logrus.Errorf("user id is not numerical: %v", err)
				respInternalError(w)
				return
			}

			user, err := a.db.UserStore.GetUser(numId)
			if err != nil {
				respError(w, err, "api.authorizeUser")
				return
			}

			if !user.IsActive {
				logrus.Debugf("refuse to serve user %d, who is not active", user.Id)
				respError(w, errors.ErrForbidden, "api.authorizeUser")
				return
			}

			ctx := r.Context()
			userCtx := context.WithValue(ctx, "user_id", numId)
			userCtx = context.WithValue(userCtx, "user", user)
			ctxReq := r.WithContext(userCtx)
			next.ServeHTTP(w, ctxReq)
			return
		}
		respBadRequest(w, "invalid token", nil)
	})
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

func (a *Api) login(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/version Authentication Login
	// Login
	//
	// responses:
	//   200:
	headers := req.Header
	token := headers.Get("Authorization")
	if len(token) > 0 {
		respBadRequest(resp, "already logged in", nil)
		return
	}

	dto := &LoginRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		respBadRequest(resp, err.Error(), nil)
		return
	}

	dto.Username = strings.ToLower(dto.Username)
	userId, err := a.db.UserStore.TryLogin(dto.Username, dto.Password)
	remoteAddr := getRemoteAddr(req)
	if userId == -1 || err != nil {
		logrus.Infof("Failed login attempt for user %s from remote %s", dto.Username, remoteAddr)
		respUnauthorized(resp)
		return
	}

	logrus.Infof("User %d '%s' logged in from %s", userId, dto.Username, remoteAddr)
	token, err = newToken(strconv.Itoa(userId), config.C.Api.Key)
	if err != nil {
		logrus.Errorf("Create new token: %v", err)
		respInternalError(resp)
		return
	}

	user, err := a.db.UserStore.GetUser(userId)
	if err != nil {
		logrus.Errorf("get user: %v", err)
	} else {
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
	}

	respBody := &LoginResponse{
		UserId: userId,
		Token:  token,
	}

	respOk(resp, respBody)
}
