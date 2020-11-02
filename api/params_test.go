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
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/storage"
)

func Test_getPaging(t *testing.T) {
	type args struct {
		buildReq func() *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    storage.Paging
		wantErr bool
	}{
		{
			name: "no paging",
			args: args{buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost/api/v1", nil)
				_ = req.ParseForm()
				return req
			}},
			want: storage.Paging{
				Offset: 0,
				Limit:  10,
			},
			wantErr: false,
		},
		{
			name: "simple paging",
			args: args{buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost/api/v1", nil)
				_ = req.ParseForm()
				req.Form.Add("page", "1")
				req.Form.Add("page_size", "10")
				return req
			}},
			want: storage.Paging{
				Offset: 0,
				Limit:  10,
			},
			wantErr: false,
		},
		{
			name: "large paging",
			args: args{buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost/api/v1", nil)
				_ = req.ParseForm()
				req.Form.Add("page", "10")
				req.Form.Add("page_size", "100")
				return req
			}},
			want: storage.Paging{
				Offset: 900,
				Limit:  100,
			},
			wantErr: false,
		},
		{
			name: "limit paging",
			args: args{buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost/api/v1", nil)
				_ = req.ParseForm()
				req.Form.Add("page", "10")
				req.Form.Add("page_size", strconv.Itoa(config.MaxRows+1))
				return req
			}},
			want: storage.Paging{
				Offset: config.MaxRows * 9,
				Limit:  config.MaxRows,
			},
			wantErr: false,
		},
		{
			name: "invalid paging",
			args: args{buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://localhost/api/v1", nil)
				_ = req.ParseForm()
				req.Form.Add("page", "-2")
				req.Form.Add("page_size", "-2")
				return req
			}},
			want: storage.Paging{
				Offset: 0,
				Limit:  10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPaging(tt.args.buildReq())
			if (err != nil) != tt.wantErr {
				t.Errorf("getPaging() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPaging() got = %v, want %v", got, tt.want)
			}
		})
	}
}
