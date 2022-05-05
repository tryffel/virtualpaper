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
	"reflect"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
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

func Test_getSortParams(t *testing.T) {
	type args struct {
		req   *http.Request
		model models.Modeler
	}

	emptyReq, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	simpleReq, _ := http.NewRequest(http.MethodGet, "http://localhost?sort=id&order=desc", nil)
	altReq, _ := http.NewRequest(http.MethodGet, "http://localhost?sort=name&order=asc", nil)
	skipReq, _ := http.NewRequest(http.MethodGet, "http://localhost?id=asc&name=desc&created_at=ab", nil)

	tests := []struct {
		name    string
		args    args
		want    []storage.SortKey
		wantErr bool
	}{
		{
			args: args{
				req:   emptyReq,
				model: &TestStruct{},
			},
			want:    []storage.SortKey{},
			wantErr: false,
		},
		{
			args: args{
				req:   simpleReq,
				model: &TestStruct{},
			},
			want: []storage.SortKey{
				{Key: "id", Order: true},
			},
			wantErr: false,
		},
		{
			args: args{
				req:   altReq,
				model: &TestStruct{},
			},
			want: []storage.SortKey{
				{Key: "name", Order: false},
			},
			wantErr: false,
		},
		{
			args: args{
				req:   skipReq,
				model: &TestStruct{},
			},
			want:    []storage.SortKey{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSortParams(tt.args.req, tt.args.model)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSortParams() = %v, want %v", got, tt.want)
			}
			if tt.wantErr != (err != nil) {
				t.Errorf("getSortParams() = %v, want error %v", err, tt.wantErr)
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
