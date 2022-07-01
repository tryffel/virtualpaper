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

import (
	"database/sql/driver"
	"fmt"
	"math"
	"time"
)

// Modeler is a basic interface for all models.
type Modeler interface {
	// Update marks model as updated.
	Update()
	// FilterAttributes returns list of attributes that can be used for filtering.
	FilterAttributes() []string

	// SortAttributes returns list of attributes that can be used for sorting.
	SortAttributes() []string

	SortNoCase() []string
}

type Timestamp struct {
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (t *Timestamp) Update() {
	t.UpdatedAt = time.Now()
}

func (t *Timestamp) FilterAttributes() []string {
	return []string{"created_at", "updated_at"}
}

func (t *Timestamp) SortAttributes() []string {
	return t.FilterAttributes()
}

func (t *Timestamp) SortNoCase() []string {
	return []string{}
}

// Int as an integer that accepts null values from database.
type Int int64

func (i *Int) Scan(src interface{}) error {
	if src == nil {
		*i = 0
		return nil
	}

	if isInt64, ok := src.(int64); ok {
		*i = Int(isInt64)
	} else if intArray, ok := src.([]uint8); ok {
		var val int64 = 0
		for i := int64(0); i < int64(len(intArray)); i++ {
			if intArray[i] < 48 || intArray[i] > 57 {
				return fmt.Errorf("not ascii number: %d", intArray[i])
			}
			val += int64(intArray[i]-48) * int64(math.Pow10(int(i)+1))
		}
		*i = Int(val)
	} else {
		return fmt.Errorf("unknown type: %T", src)
	}
	return nil
}

func (i Int) Value() (driver.Value, error) {
	return int64(i), nil
}
