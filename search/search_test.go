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
	"time"
	"tryffel.net/go/virtualpaper/models"
)

func Test_tokenizeFilter(t *testing.T) {
	type args struct {
		filter string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{`a author:doyle OR (topic:"misc topic" AND author:doyle) and one more`},
			want: []string{"a", "author:doyle", "OR", "(", "topic:misc topic", "AND", "author:doyle", ")", "and", "one", "more"},
		},
		{
			args: args{`a "complex key":"complex value"`},
			want: []string{"a", "complex key:complex value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tokenizeFilter(tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenizeFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFilter(t *testing.T) {
	type args struct {
		filter string
	}
	tests := []struct {
		name    string
		args    args
		want    *searchQuery
		wantErr bool
	}{
		{
			name: "simple fts",
			args: args{"test one"},
			want: &searchQuery{
				RawQuery:      "test one",
				Query:         "test one",
				MetadataQuery: []string{},
			},
			wantErr: false,
		},
		{
			name: "simple key:value",
			args: args{"simple key:value"},
			want: &searchQuery{
				RawQuery:       "simple key:value",
				Query:          "simple",
				MetadataQuery:  []string{`metadata="key:value"`},
				MetadataString: `metadata="key:value"`,
			},
			wantErr: false,
		},
		{
			name: "multi word key values",
			args: args{`simple "key 2":"complex value" AND key:value`},
			want: &searchQuery{
				RawQuery:       `simple "key 2":"complex value" AND key:value`,
				Query:          "simple",
				MetadataQuery:  []string{`metadata="key_2:complex_value"`, "AND", `metadata="key:value"`},
				MetadataString: `metadata="key_2:complex_value" AND metadata="key:value"`,
			},
			wantErr: false,
		},
		{
			name: "multiple metadata",
			args: args{"simple key:value AND another:value more search"},
			want: &searchQuery{
				RawQuery:       "simple key:value AND another:value more search",
				Query:          "simple more search",
				MetadataQuery:  []string{`metadata="key:value"`, "AND", `metadata="another:value"`},
				MetadataString: `metadata="key:value" AND metadata="another:value"`,
			},
			wantErr: false,
		},
		{
			name: "multiple metadata with parentheses",
			args: args{"simple key:value AND (another:value OR key:value) more search"},
			want: &searchQuery{
				RawQuery:       "simple key:value AND (another:value OR key:value) more search",
				Query:          "simple more search",
				MetadataQuery:  []string{`metadata="key:value"`, "AND", "(", `metadata="another:value"`, "OR", `metadata="key:value"`, ")"},
				MetadataString: `metadata="key:value" AND ( metadata="another:value" OR metadata="key:value" )`,
			},
			wantErr: false,
		},
		{
			name: "date today",
			args: args{"date:today"},
			want: &searchQuery{
				RawQuery:       "date:today",
				Query:          "",
				MetadataQuery:  []string{},
				MetadataString: "",
				DateBefore:     models.MidnightForDate(time.Now().AddDate(0, 0, 1).Local()),
				DateAfter:      models.MidnightForDate(time.Now().AddDate(0, 0, 0).Local()),
			},
			wantErr: false,
		},
		{
			name: "date yesterday",
			args: args{"date:yesterday"},
			want: &searchQuery{
				RawQuery:       "date:yesterday",
				Query:          "",
				MetadataQuery:  []string{},
				MetadataString: "",
				DateBefore:     models.MidnightForDate(time.Now().AddDate(0, 0, 0).Local()),
				DateAfter:      models.MidnightForDate(time.Now().AddDate(0, 0, -1).Local()),
			},
			wantErr: false,
		},
		{
			name: "name",
			args: args{`name:"docname two"`},
			want: &searchQuery{
				RawQuery:       `name:"docname two"`,
				MetadataQuery:  []string{},
				MetadataString: "",
				Name:           "docname two",
			},
			wantErr: false,
		},
		{
			name: "content",
			args: args{`content:"one two three"`},
			want: &searchQuery{
				RawQuery:       `content:"one two three"`,
				MetadataQuery:  []string{},
				MetadataString: "",
				Content:        "one two three",
			},
			wantErr: false,
		},
		{
			name: "description",
			args: args{`description:"one two three"`},
			want: &searchQuery{
				RawQuery:       `description:"one two three"`,
				MetadataQuery:  []string{},
				MetadataString: "",
				Description:    "one two three",
			},

			wantErr: false,
		},
		{
			name: "combined",
			args: args{`fts test date:today class:paper OR class:invoice"`},
			want: &searchQuery{
				RawQuery:       `fts test date:today class:paper OR class:invoice"`,
				Query:          "fts test",
				MetadataQuery:  []string{`metadata="class:paper"`, "OR", `metadata="class:invoice"`},
				MetadataString: `metadata="class:paper" OR metadata="class:invoice"`,
				DateBefore:     models.MidnightForDate(time.Now().Local().AddDate(0, 0, 1)),
				DateAfter:      models.MidnightForDate(time.Now().Local().AddDate(0, 0, 0)),
			},

			wantErr: false,
		},
		{
			name: "date: year",
			args: args{`date:2022`},
			want: &searchQuery{
				RawQuery:       `date:2022`,
				Query:          "",
				MetadataQuery:  []string{},
				MetadataString: "",
				DateBefore:     timeFromDate(2023, 1, 1),
				DateAfter:      timeFromDate(2022, 1, 1),
			},

			wantErr: false,
		},
		{
			name: "date: month range",
			args: args{`date:2022|2022-06`},
			want: &searchQuery{
				RawQuery:       `date:2022|2022-06`,
				Query:          "",
				MetadataQuery:  []string{},
				MetadataString: "",
				DateBefore:     timeFromDate(2022, 7, 1),
				DateAfter:      timeFromDate(2022, 1, 1),
			},

			wantErr: false,
		},
		{
			name: "date: year range",
			args: args{`date:2020|today`},
			want: &searchQuery{
				RawQuery:       `date:2020|today`,
				Query:          "",
				MetadataQuery:  []string{},
				MetadataString: "",
				DateBefore:     models.MidnightForDate(time.Now().AddDate(0, 0, 0)),
				DateAfter:      timeFromDate(2020, 1, 1),
			},

			wantErr: false,
		},
		{
			name: "date: day-to-yesterday",
			args: args{`date:2015-6-12|yesterday`},
			want: &searchQuery{
				RawQuery:       `date:2015-6-12|yesterday`,
				Query:          "",
				MetadataQuery:  []string{},
				MetadataString: "",
				DateBefore:     models.MidnightForDate(time.Now().AddDate(0, 0, -1)),
				DateAfter:      timeFromDate(2015, 6, 12),
			},

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFilter(tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func timeFromDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
