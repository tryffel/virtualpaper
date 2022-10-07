/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2022  Tero Vierimaa
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
	"strings"
	"testing"
	"time"

	"tryffel.net/go/virtualpaper/models"
)

func Test_matchDate(t *testing.T) {

	type args struct {
		token string
	}
	tests := []struct {
		name  string
		args  args
		want  valueMatchStatus
		want1 string
		want2 time.Time
		want3 time.Time
	}{
		{
			name:  "no suggestion",
			args:  args{"asdf"},
			want:  valueMatchStatusInvalid,
			want1: "",
			want2: time.Time{},
			want3: time.Time{},
		},
		{
			name:  "parse date1",
			args:  args{"2022-07-01"},
			want:  valueMatchStatusOk,
			want1: "",
			want2: models.MidnightForDate(time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)),
			want3: models.MidnightForDate(time.Date(2022, 7, 2, 0, 0, 0, 0, time.UTC)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3 := matchDate(tt.args.token)
			if got != tt.want {
				t.Errorf("matchDate() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("matchDate() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("matchDate() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("matchDate() got3 = %v, want %v", got3, tt.want3)
			}

		})
	}
}

type metadata struct {
	keys     []string
	metadata map[string][]models.Metadata
}

func newMetadata() *metadata {
	m := &metadata{
		keys:     []string{},
		metadata: make(map[string][]models.Metadata),
	}
	m.addKey("class")
	m.addKey("author")
	m.addKey("authentic")
	m.addKey("topic")
	m.addKey("datasource")

	m.addValue("class", "paper")
	m.addValue("class", "invoice")

	m.addValue("author", "doyle")
	m.addValue("author", "dustin")
	m.addValue("author", "dubai")

	m.addValue("authentic", "true")
	m.addValue("authentic", "false")

	m.addValue("topic", "science")
	m.addValue("topic", "technology")
	m.addValue("topic", "logic")
	m.addValue("topic", "arts")
	return m
}

func (m *metadata) addKey(key string) {
	m.keys = append(m.keys, key)
	m.metadata[key] = []models.Metadata{}
}

func (m *metadata) addValue(key, value string) {
	array := m.metadata[key]
	array = append(array, models.Metadata{Key: key, Value: value})
	m.metadata[key] = array
}

func (m *metadata) queryKeys(key string) []string {
	results := []string{}

	for _, v := range m.keys {
		if key == "" {
			results = append(results, v)
		} else if strings.Contains(v, key) {
			results = append(results, v)
		}
	}
	return results
}

func (m *metadata) queryValues(key, value string) []string {
	results := []string{}

	keyValues, ok := m.metadata[key]
	if !ok {
		return results
	}

	for _, v := range keyValues {
		if value == "" {
			results = append(results, v.Value)
		} else if strings.Contains(v.Value, value) {
			results = append(results, v.Value)
		}
	}
	return results
}

/*
func Test_suggest(t *testing.T) {
	metadata := newMetadata()

	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want *QuerySuggestions
	}{
		{
			name: "fts, no suggestions",
			args: args{"one day"},
			want: &QuerySuggestions{[]string{}, "one day"},
		},
		{
			name: "fts, suggest date and metadata key",
			args: args{"one da"},
			want: &QuerySuggestions{[]string{"date:", "datasource:"}, "one "},
		},
		{
			name: "fts, suggest metadata keys",
			args: args{"one auth"},
			want: &QuerySuggestions{[]string{"author:", "authentic:"}, "one "},
		},
		{
			name: "fts suggest metadata key",
			args: args{"one autho"},
			want: &QuerySuggestions{[]string{"author:"}, "one "},
		},
		{
			name: "fts: suggest semicolon after key",
			args: args{"one author"},
			want: &QuerySuggestions{[]string{"author:"}, "one "},
		},
		{
			name: "suggest metadata inside parantheses",
			args: args{"one author ( "},
			want: &QuerySuggestions{[]string{"class:", "author:", "authentic:", "topic:", "datasource:"}, "one author ( "},
		},
		{
			name: "autocomplete metadata inside parantheses",
			args: args{"one author ( aut "},
			want: &QuerySuggestions{[]string{"author:", "authentic:"}, "one author ( "},
		},
		{
			name: "suggest metadata and operators inside parantheses",
			args: args{"one author ( author:value "},
			want: &QuerySuggestions{[]string{"AND", "OR", "NOT", ")"}, "one author ( author:value "},
		},
		{
			name: "suggest date ",
			args: args{"one date:"},
			want: &QuerySuggestions{[]string{"today", "yesterday"}, "one date:"},
		},
		{
			name: "autocomplete date",
			args: args{"one date:to"},
			want: &QuerySuggestions{[]string{"today"}, "one date:"},
		},
		{
			name: "don't autocomplete date",
			args: args{"one date:an"},
			want: &QuerySuggestions{[]string{}, "one date:"},
		},
		{
			name: "suggest metadata value",
			args: args{"one author:"},
			want: &QuerySuggestions{[]string{"doyle", "dustin", "dubai"}, "one author:"},
		},
		{
			name: "autocomplete metadata value",
			args: args{"one author:du"},
			want: &QuerySuggestions{[]string{"dustin", "dubai"}, "one author:"},
		},
		{
			name: "autocomplete metadata value fuzzy",
			args: args{"one topic:log"},
			want: &QuerySuggestions{[]string{"technology", "logic"}, "one topic:"},
		},
		{
			name: "metadata value match no suggestion",
			args: args{"one topic:technology"},
			want: &QuerySuggestions{[]string{}, "one topic:technology"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := suggest(tt.args.query, metadata); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("suggest() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
