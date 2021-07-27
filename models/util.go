/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2021  Tero Vierimaa
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

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// GetSize returns human-formatted size
func GetPrettySize(bytes int64) string {
	if bytes < 1024 {
		return strconv.Itoa(int(bytes))
	}
	size := float64(bytes)
	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%.2f KiB", size)
	}
	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%.2f MiB", size)
	}
	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%.2f GiB", size)
	}
	return fmt.Sprintf("%f B", size)

}

type IntId uint64

func (i IntId) Value() (driver.Value, error) {
	if i == 0 {
		return nil, nil
	}
	return int64(i), nil
}

// Scan scans duration from postgres string: (00:00:00). Only hours-minutes-seconds are supported.
func (i *IntId) Scan(src interface{}) error {
	if src == nil {
		*i = 0
		return nil
	}

	isInt64, ok := src.(int64)
	if ok {
		*i = IntId(isInt64)
		return nil
	}
	return fmt.Errorf("unknown type: %v", src)
}
