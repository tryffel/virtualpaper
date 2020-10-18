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

package models

import "time"

type Document struct {
	*Timestamp
	Id             int       `db:"id"`
	UserId         int       `db:"user_id"`
	Name           string    `db:"name"`
	Content        string    `db:"content"`
	Filename       string    `db:"filename"`
	Preview        string    `db:"preview"`
	Hash           string    `db:"hash"`
	IndexedAt      time.Time `db:"indexed_at"`
	AwaitsIndexing bool      `db:"awaits_indexing"`
}
