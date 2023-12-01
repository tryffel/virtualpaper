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
)

func (a *Api) getUserDocumentStatistics(c echo.Context) error {
	// swagger:route GET /api/v1/documents/stats Documents GetUserDocumentStatistics
	// Get document statistics
	//
	// responses:
	//   200: RespDocumentStatistics
	//   304: RespNotModified
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	ctx := c.(UserContext)

	stats, err := a.documentService.GetStatistics(getContext(c), ctx.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, stats)
}
