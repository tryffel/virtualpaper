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
	"time"
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

type Text string

func (t Text) Value() (driver.Value, error) {
	return string(t), nil
}

func (t Text) String() string {
	return string(t)
}

func (t *Text) Scan(src interface{}) error {
	if src == nil {
		*t = ""
		return nil
	}

	isStr, ok := src.(string)
	if ok {
		*t = Text(isStr)
		return nil
	}

	return fmt.Errorf("unknown type: %v", src)
}

func MidnightForDate(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type Language struct {
	Id   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type Lang string

func (l *Lang) Scan(src interface{}) error {
	if src == nil {
		*l = ""
		return nil
	}
	if str, ok := src.(string); ok {
		*l = Lang(str)
		return nil
	}
	return fmt.Errorf("invalid type: %v, expected string", src)
}

func (l Lang) Value() (driver.Value, error) {
	if l == "" {
		return nil, nil
	}
	return string(l), nil
}

func (l Lang) String() string {
	return string(l)
}
