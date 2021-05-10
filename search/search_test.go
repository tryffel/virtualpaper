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

package search

import "testing"

func Test_parseFilter(t *testing.T) {
	type args struct {
		filter string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{filter: ""},
			want: "",
		},
		{
			args: args{filter: `class:book AND (name:"james bond" OR name:"agatha christie")`},
			want: `metadata="class:book" AND ( metadata="name:james_bond" OR metadata="name:agatha_christie" )`,
		},
		{
			args: args{filter: `class:book NOT author:"james bond"`},
			want: `metadata="class:book" NOT metadata="author:james_bond"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseFilter(tt.args.filter); got != tt.want {
				t.Errorf("parseFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
