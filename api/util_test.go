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
	"net/http"
	"testing"
	"time"
)

func Test_getUserId(t *testing.T) {
	type args struct {
		req *http.Request
	}

	emptyReq, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	validReq, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	ctx := context.WithValue(validReq.Context(), "user_id", 10)
	validReq = validReq.WithContext(ctx)

	tests := []struct {
		name  string
		args  args
		want  int
		want1 bool
	}{
		{
			args:  args{emptyReq},
			want:  0,
			want1: false,
		},
		{
			args:  args{validReq},
			want:  10,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getUserId(tt.args.req)
			if got != tt.want {
				t.Errorf("getUserId() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getUserId() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type TestStruct struct {
	Id        int
	Name      string
	CreatedAt time.Time
}

func (t *TestStruct) Update() {}

func (t *TestStruct) FilterAttributes() []string {
	return []string{"id", "name", "created_at"}
}

func (t *TestStruct) SortAttributes() []string {
	return t.FilterAttributes()
}

func (t *TestStruct) SortNoCase() []string {
	return []string{"name"}
}
