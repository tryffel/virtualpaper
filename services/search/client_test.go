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

import (
	"reflect"
	"testing"
)

func Test_buildSynonyms(t *testing.T) {
	type args struct {
		synonyms [][]string
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			args: args{[][]string{
				{
					"a",
					"b",
					"c",
					"d",
					"e",
				},
				{
					"f",
				},
				{
					"ab",
					"ac",
				},
			}},
			want: map[string][]string{
				"a":  {"b", "c", "d", "e"},
				"b":  {"a", "c", "d", "e"},
				"c":  {"a", "b", "d", "e"},
				"d":  {"a", "b", "c", "e"},
				"e":  {"a", "b", "c", "d"},
				"ab": {"ac"},
				"ac": {"ab"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildSynonyms(tt.args.synonyms); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildSynonyms() = %v, want %v", got, tt.want)
			}
		})
	}
}
