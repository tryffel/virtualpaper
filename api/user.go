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
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

// swagger:model UserPreferences
type UserPreferences struct {
	// user
	Id                  int        `json:"user_id"`
	Name                string     `json:"user_name"`
	Email               string     `json:"email"`
	UpdatedAt           int64      `json:"updated_at"`
	CreatedAt           int64      `json:"created_at"`
	DocumentsCount      int64      `json:"documents_count"`
	DocumentsSize       int64      `json:"documents_size"`
	DocumentsSizeString string     `json:"documents_size_string"`
	IsAdmin             bool       `json:"is_admin"`
	StopWords           []string   `json:"stop_words"`
	Synonyms            [][]string `json:"synonyms"`
}

func (u *UserPreferences) copyUser(userPref *models.UserPreferences) {
	u.Id = userPref.UserId
	u.Name = userPref.UserName
	u.Email = userPref.Email
	u.UpdatedAt = userPref.UpdatedAt.Unix() * 1000
	u.CreatedAt = userPref.CreatedAt.Unix() * 1000
	u.DocumentsCount = int64(userPref.DocumentCount)
	u.DocumentsSize = int64(userPref.DocumentsSize)
	u.DocumentsSizeString = models.GetPrettySize(u.DocumentsSize)
	u.IsAdmin = userPref.IsAdmin
	u.StopWords = userPref.StopWords
	u.Synonyms = userPref.Synonyms
}

func (a *Api) getUserPreferences(c echo.Context) error {
	// swagger:route GET /api/v1/preferences/user Preferences GetPreferences
	// Get user preferences
	// responses:
	//   200: RespUserPreferences
	//   304: RespNotModified
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError
	//
	ctx := c.(UserContext)
	preferences, err := a.userService.GetPreferences(getContext(c), ctx.UserId)
	if err != nil {
		return err
	}
	userPref := &UserPreferences{}
	userPref.copyUser(preferences)
	return c.JSON(http.StatusOK, userPref)
}

// swagger:model UserPreferences
type ReqUserPreferences struct {
	StopWords []string   `json:"stop_words" valid:"optional"`
	Synonyms  [][]string `json:"synonyms" valid:"optional"`
	Email     string     `json:"email" valid:"email,optional"`
}

func (a *Api) updateUserPreferences(c echo.Context) error {
	ctx := c.(UserContext)

	dto := &ReqUserPreferences{}
	err := unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	pref := &models.UserPreferences{
		UserId:        ctx.UserId,
		UserName:      ctx.User.Name,
		Email:         dto.Email,
		UpdatedAt:     time.Time{},
		CreatedAt:     time.Time{},
		DocumentCount: 0,
		DocumentsSize: 0,
		IsAdmin:       ctx.User.IsAdmin,
		StopWords:     dto.StopWords,
		Synonyms:      dto.Synonyms,
	}

	err = a.userService.UpdatePreferences(getContext(ctx), pref)
	if err != nil {
		return err
	}

	userPref, err := a.userService.GetPreferences(getContext(c), ctx.UserId)
	if err != nil {
		return err

	}
	return c.JSON(http.StatusOK, userPref)
}

func (a *Api) GetUsers(c echo.Context) error {
	ctx := c.(UserContext)
	users, err := a.userService.GetUsers(getContext(ctx))
	if err != nil {
		return err
	}
	return resourceList(c, users, len(*users))
}
