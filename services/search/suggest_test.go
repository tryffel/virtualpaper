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
			want1: "2022-07-01",
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

	m.addKey("whitespace key")
	m.addValue("whitespace key", "whitespace value")

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

func (m *metadata) queryKeys(key string, prefix string, suffix string) []string {
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
func (m *metadata) queryLangs(key string) []string {
	return []string{"fi", "en"}
}

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
			name: "empty query",
			args: args{""},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "name", Type: "key", Hint: ""},
				{Value: "description", Type: "key", Hint: ""},
				{Value: "content", Type: "key", Hint: ""},
				{Value: "date", Type: "key", Hint: ""},
				{Value: "lang", Type: "key", Hint: ""},
				{Value: "owner", Type: "key", Hint: ""},
				{Value: "class", Type: "metadata", Hint: ""},
				{Value: "author", Type: "metadata", Hint: ""},
				{Value: "authentic", Type: "metadata", Hint: ""},
				{Value: "topic", Type: "metadata", Hint: ""},
				{Value: "datasource", Type: "metadata", Hint: ""},
				{Value: "whitespace key", Type: "metadata", Hint: ""},
			}, Prefix: "", ValidQuery: false},
		},
		{
			name: "fts, no suggestions",
			args: args{"one day"},
			want: &QuerySuggestions{Suggestions: []Suggestion{}, Prefix: "one day", ValidQuery: false},
		},

		{
			name: "fts, suggest date and metadata key",
			args: args{"one da"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "date:", Type: "key", Hint: ""},
				{Value: "datasource:", Type: "metadata", Hint: ""},
			}, Prefix: "one ", ValidQuery: false},
		},
		{
			name: "fts, suggest metadata keys",
			args: args{"one auth"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "author:", Type: "metadata", Hint: ""},
				{Value: "authentic:", Type: "metadata", Hint: ""},
			}, Prefix: "one ", ValidQuery: false},
		},
		{
			name: "fts suggest metadata key",
			args: args{"one autho"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "author:", Type: "metadata", Hint: ""},
			}, Prefix: "one ", ValidQuery: false},
		},
		{
			name: "fts: suggest semicolon after key",
			args: args{"one author"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "author:", Type: "metadata", Hint: ""},
				{Value: "author:doyle", Type: "metadata", Hint: ""},
				{Value: "author:dustin", Type: "metadata", Hint: ""},
				{Value: "author:dubai", Type: "metadata", Hint: ""},
			}, Prefix: "one ", ValidQuery: false},
		},
		{
			name: "suggest metadata inside parantheses",
			args: args{"one author ( "},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "class:", Type: "metadata", Hint: ""},
				{Value: "author:", Type: "metadata", Hint: ""},
				{Value: "authentic:", Type: "metadata", Hint: ""},
				{Value: "topic:", Type: "metadata", Hint: ""},
				{Value: "datasource:", Type: "metadata", Hint: ""},
				{Value: `"whitespace key":`, Type: "metadata", Hint: ""},
			}, Prefix: "one author ( ", ValidQuery: false},
		},
		{
			name: "autocomplete metadata inside parantheses",
			args: args{"one author ( aut "},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "author:", Type: "metadata", Hint: ""},
				{Value: "authentic:", Type: "metadata", Hint: ""},
			}, Prefix: "one author ( ", ValidQuery: false},
		},
		{
			name: "autocomplete metadata with whitespace",
			args: args{"one space "},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: `"whitespace key":`, Type: "metadata", Hint: ""},
			}, Prefix: "one ", ValidQuery: false},
		},
		{
			name: "suggest metadata and operators inside parantheses",
			args: args{"one author ( author:value "},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "AND", Type: "operand", Hint: ""},
				{Value: "OR", Type: "operand", Hint: ""},
				{Value: "NOT", Type: "operand", Hint: ""},
				{Value: ")", Type: "operand", Hint: ""},
				// TODO: fix prefix removing metadata value
			}, Prefix: "one author ( author ", ValidQuery: false},
		},
		{
			name: "suggest date ",
			args: args{"one date:"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "today", Type: "key", Hint: ""},
				{Value: "yesterday", Type: "key", Hint: ""},
				{Value: "week", Type: "key", Hint: ""},
				{Value: "month", Type: "key", Hint: ""},
				{Value: "year", Type: "key", Hint: ""},
				{Value: time.Now().Format("2006"), Type: "key", Hint: ""},
				{Value: time.Now().Format("2006-1"), Type: "key", Hint: ""},
				{Value: time.Now().Format("2006-1-2"), Type: "key", Hint: ""},
				{Value: time.Now().AddDate(-2, 0, 0).Format("2006") + "|", Type: "key", Hint: ""},
				{Value: time.Now().AddDate(-2, 0, 0).Format("2006") + "|today", Type: "key", Hint: ""},
				{Value: time.Now().AddDate(-5, 0, 0).Format("2006") + "|" + time.Now().AddDate(-1, 0, 0).Format("2006"), Type: "key", Hint: ""},
			}, Prefix: "one date:", ValidQuery: false},
		},
		{
			name: "autocomplete date",
			args: args{"one date:to"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "today", Type: "key", Hint: ""},
			}, Prefix: "one date:", ValidQuery: false},
		},
		{
			name: "don't autocomplete date",
			args: args{"one date:an"},
			want: &QuerySuggestions{Suggestions: []Suggestion{}, Prefix: "one date:an", ValidQuery: false},
		},
		{
			name: "suggest metadata value",
			args: args{"one author:"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "doyle", Type: "metadata", Hint: ""},
				{Value: "dustin", Type: "metadata", Hint: ""},
				{Value: "dubai", Type: "metadata", Hint: ""},
			}, Prefix: "one author:", ValidQuery: false},
		},
		{
			name: "autocomplete metadata value",
			args: args{"one author:du"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "dustin", Type: "metadata", Hint: ""},
				{Value: "dubai", Type: "metadata", Hint: ""},
			}, Prefix: "one author:", ValidQuery: false},
		},
		{
			name: "autocomplete metadata value with whitespace",
			args: args{`one "whitespace key":"val"`},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: `"whitespace value"`, Type: "metadata", Hint: ""},
			}, Prefix: `one "whitespace key":`, ValidQuery: false},
		},
		{
			name: "autocomplete metadata value fuzzy",
			args: args{"one topic:log"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "technology", Type: "metadata", Hint: ""},
				{Value: "logic", Type: "metadata", Hint: ""},
			}, Prefix: "one topic:", ValidQuery: false},
		},
		{
			name: "metadata value match no suggestion",
			args: args{"one topic:technology"},
			want: &QuerySuggestions{Suggestions: []Suggestion{}, Prefix: "one topic:technology", ValidQuery: false},
		},
		{
			name: "lang",
			args: args{"one lan"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "lang:", Type: "key"},
			}, Prefix: "one ", ValidQuery: false},
		},
		{
			name: "owner",
			args: args{"own"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "owner:", Type: "key"},
			}, Prefix: " ", ValidQuery: false},
		},
		{
			name: "owner value",
			args: args{"owner:"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "me", Type: "key"},
				{Value: "anyone", Type: "key"},
				{Value: "others", Type: "key"},
			}, Prefix: "owner:", ValidQuery: false},
		},
		{
			name: "shared",
			args: args{"sha"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "shared:", Type: "key"},
			}, Prefix: " ", ValidQuery: false},
		},
		{
			name: "shared value",
			args: args{"shared:"},
			want: &QuerySuggestions{Suggestions: []Suggestion{
				{Value: "yes", Type: "key"},
				{Value: "no", Type: "key"},
			}, Prefix: "shared:", ValidQuery: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggest(tt.args.query, metadata)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("suggest() = %v, want %v", got, tt.want)
			}
		})
	}
}
