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
	"tryffel.net/go/virtualpaper/config"
)

const (
	tokenClaimUserid = "user_id"
)

func getUserId(req *http.Request) (int, bool) {
	ctx := req.Context()
	userId := ctx.Value("user_id")
	id, ok := userId.(int)
	return id, ok
}

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
			err = respBadRequest(w, "Invalid token", nil)
			return
		}

		userId, err := validateToken(parts[1], config.C.Api.Key)
		if userId == "" || err != nil {
			respBadRequest(w, "invalid token", nil)
		}

		if userId != "" && err == nil {
			numId, err := strconv.Atoi(userId)
			if err != nil {
				logrus.Errorf("user id is not numerical: %v", err)
				respInternalError(w)
				return
			}
			ctx := r.Context()
			userCtx := context.WithValue(ctx, "user_id", numId)
			ctxReq := r.WithContext(userCtx)
			next.ServeHTTP(w, ctxReq)
			return
		}
		respBadRequest(w, "invalid token", nil)
	})
}

// newToken issues a new token for user_id
// if ExpireDuration == 0, disable expiration
func newToken(userId string, privateKey string) (string, error) {
	var token *jwt.Token = nil
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		tokenClaimUserid: userId,
	})

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
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		user := claims[tokenClaimUserid].(string)
		return user, nil
	}
	return "", fmt.Errorf("invalid token")
}

type loginBody struct {
	Username string `valid:"alphanum,required"`
	Password string `valid:"required"`
}

type loginResponse struct {
	UserId int
	Token  string
}

func (a *Api) login(resp http.ResponseWriter, req *http.Request) {
	headers := req.Header
	token := headers.Get("Authorization")
	if len(token) > 0 {
		err := respBadRequest(resp, "already logged in", nil)
		if err != nil {
			logrus.Error(err)
		}
		return
	}

	dto := &loginBody{}
	err := unMarshalBody(req, dto)
	if err != nil {
		respBadRequest(resp, err.Error(), nil)
		return
	}

	userId, err := a.db.UserStore.TryLogin(dto.Username, dto.Password)
	if userId == -1 || err != nil {
		respUnauthorized(resp)
		return
	}

	token, err = newToken(strconv.Itoa(userId), config.C.Api.Key)
	if err != nil {
		logrus.Errorf("Create new token: %v", err)
		respInternalError(resp)
		return
	}

	respBody := &loginResponse{
		UserId: userId,
		Token:  token,
	}

	respOk(resp, respBody)
}
