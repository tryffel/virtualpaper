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
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/services"
	log "tryffel.net/go/virtualpaper/util/logger"
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

			userId, tokenKey, err := validateToken(c, parts[1], config.C.Api.Key)
			if userId == "" || err != nil {
				return
			}

			userNumId := 0
			if userId != "" {
				userNumId, err = strconv.Atoi(userId)
				if err != nil {
					c.Logger().Error("user id is not numerical", userId)
					return echo.ErrInternalServerError
				}
			}

			user, token, err := a.authService.GetUserByToken(getContext(c), tokenKey, userNumId)
			if err != nil {
				if errors.Is(err, errors.ErrRecordNotFound) {
					return authErr
				} else if errors.Is(err, errors.ErrUnauthorized) {
					return authErr
				}
				return fmt.Errorf("get token from database: %v", err)
			}
			ctx := UserContext{
				Context: Context{Context: c, pagination: pageParams{
					Page:     1,
					PageSize: 20,
				},
					sort: SortKey{},
				},
				Admin:    user.IsAdmin,
				UserId:   userNumId,
				User:     user,
				TokenKey: token.Key,
			}
			return next(ctx)
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

			err := a.authService.ConfirmAuthToken(getContext(c), ctx.TokenKey)
			if err != nil {
				if errors.Is(err, errors.ErrInvalid) {
					err := errors.ErrInvalid
					err.ErrMsg = "authentication required"
					return err
				}
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
func validateToken(c echo.Context, tokenString string, privateKey string) (string, string, error) {
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
				log.Debugf(getContext(c), "token expired")
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
			c.Logger().Infof("token expired")
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

	ctx := getContext(c)
	remoteAddr := getRemoteAddr(req)

	ua := useragent.Parse(c.Request().Header.Get("User-Agent"))
	userAgent := fmt.Sprintf("%s %s, %s %s", ua.OS, ua.OSVersion, ua.Name, ua.Version)

	authToken, err := a.authService.Login(getContext(c), dto.Username, dto.Password, userAgent, c.RealIP())
	if err != nil {
		log.Entry(ctx).WithField("user", dto.Username).WithField("remoteAddr", remoteAddr).Infof("Failed login attempt")
		// request takes about the same time with invalid password & invalid user
		time.Sleep(time.Millisecond*1100 + time.Duration(int(rand.Float64()*1000))*time.Millisecond)
		return echo.ErrUnauthorized
	}

	log.Entry(ctx).WithField("user", dto.Username).WithField("remoteAddr", remoteAddr).Infof("User logged in")
	token, err = newToken(strconv.Itoa(authToken.UserId), authToken.Key, config.C.Api.Key)
	if err != nil {
		c.Logger().Errorf("Create new token: %v", err)
		return respInternalErrorV2(c, err)
	}
	respBody := &LoginResponse{
		UserId: authToken.UserId,
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
	err := a.authService.RevokeToken(getContext(c), ctx.TokenKey)
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

	err = a.authService.ResetPassword(getContext(c), &services.PasswordReset{
		Token:    dto.Token,
		Id:       dto.Id,
		Password: dto.Password,
	})
	if err != nil {
		return err
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

	err = a.authService.CreateResetPasswordToken(getContext(c), dto.Email)
	if err != nil {
		return err
	}
	msg := map[string]string{"status": "Password reset email has been sent"}
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
	remoteAddr := getRemoteAddr(req)
	err = a.authService.ConfirmAuthentication(getContext(c), user.User, dto.Password, remoteAddr, user.TokenKey)
	if err != nil {
		return err
	}
	return c.JSON(200, "")
}

func ValidatePassword(password string) error {
	err := errors.ErrInvalid
	if len(password) < 8 || len(password) > 100 {
		err.ErrMsg = "password's length must be 8 - 100 characters"
		return err
	}
	return nil
}
