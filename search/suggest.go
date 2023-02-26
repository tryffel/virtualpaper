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
	"fmt"
	"regexp"
	"strings"
	"time"

	"tryffel.net/go/virtualpaper/models"

	"github.com/meilisearch/meilisearch-go"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/storage"
)

// total maximum for suggestions
const MaxSuggestions = 50

// max for either metadate keys or values
const MaxSuggestMetadata = 10

func (e *Engine) SuggestSearch(userId int, query string) (*QuerySuggestions, error) {
	metadata := &metadataSuggest{
		db:     e.db,
		userId: userId,
	}
	return suggest(query, metadata), nil
}

type metadataSuggest struct {
	db     *storage.Database
	userId int
}

func (m *metadataSuggest) queryKeys(key, prefix, suffix string) []string {
	keys, err := m.db.MetadataStore.GetUserKeysCached(m.userId)
	if err != nil {
		logrus.Error(err)
		return []string{}
	}

	data := []string{}

	for i, v := range *keys {
		if i > MaxSuggestMetadata {
			break
		}

		if key == "" {
			data = append(data, prefix+v.Key+suffix)
		} else if strings.Contains(v.Key, key) {
			data = append(data, prefix+v.Key+suffix)
		}
	}
	return data
}

func (m *metadataSuggest) queryValues(key, value string) []string {
	values, err := m.db.MetadataStore.GetUserKeyValuesCached(m.userId, key)
	if err != nil {
		logrus.Error(err)
		return []string{}
	}
	data := []string{}
	for _, v := range *values {
		if len(data) > MaxSuggestMetadata {
			break
		}
		if value == "" {
			data = append(data, v.Value)

		} else if strings.Contains(v.Value, value) {
			data = append(data, v.Value)
		}
	}
	return data
}

type searchQuery struct {
	RawQuery       string
	Query          string
	Name           string
	Description    string
	Content        string
	DateBefore     time.Time
	DateAfter      time.Time
	MetadataQuery  []string
	MetadataString string
	Suggestions    []string
}

func (s *searchQuery) addSuggestion(text string) {
	// sort ?
	s.Suggestions = append(s.Suggestions, text)
}

func (s *searchQuery) prepareMeiliQuery(userId int, sort storage.SortKey, paging storage.Paging) *meilisearch.SearchRequest {

	request := &meilisearch.SearchRequest{
		Offset:                int64(paging.Offset),
		Limit:                 int64(paging.Limit),
		AttributesToRetrieve:  []string{"document_id", "name", "content", "description", "date", "mimetype"},
		AttributesToCrop:      []string{"content"},
		CropLength:            1000,
		AttributesToHighlight: []string{"name"},
		PlaceholderSearch:     false,
	}
	filter := s.MetadataString
	if s.Query == "" {
		request.PlaceholderSearch = true
	}

	//facets := make([]interface{}, 0)
	//if d.Tag != "" {
	//facets = append(facets, fmt.Sprintf(`tags:%s`, d.Tag))
	//}

	datefilters := []string{}
	if !s.DateAfter.IsZero() {
		datefilters = append(datefilters, fmt.Sprintf("date >= %d", s.DateAfter.Unix()))
		logrus.Tracef("search after %s", s.DateAfter.Format("2006-1-2"))
	}
	if !s.DateBefore.IsZero() {
		datefilters = append(datefilters, fmt.Sprintf("date < %d", s.DateBefore.Unix()))
		logrus.Tracef("search before %s", s.DateBefore.Format("2006-1-2"))

	}
	if len(datefilters) > 0 {
		filter = strings.Join(append(datefilters, filter), " AND ")
	}
	filter = strings.TrimSuffix(filter, " AND ")
	//logrus.Infof("filter: %s", filter)
	// TODO : fix
	//if sort.Key != "" {
	//request.Sort = []string{sort.Key + " :" + strings.ToLower(sort.SortOrder())}
	//}

	if s.Name != "" {
		q := fmt.Sprintf(`name="%s"`, s.Name)
		filter = strings.Join([]string{filter, q}, " AND ")
	} else if s.Description != "" {
		q := fmt.Sprintf(`description="%s"`, s.Description)
		filter = strings.Join([]string{filter, q}, " AND ")
	} else if s.Content != "" {
		q := fmt.Sprintf(`content="%s"`, s.Content)
		filter = strings.Join([]string{filter, q}, " AND ")
	}

	if filter != "" {
		// don't set empty filter, it will block all results
		request.Filter = filter
	}

	if sort.Key != "" && sort.Key != "id" {
		request.Sort = []string{sort.Key + ":" + strings.ToLower(sort.SortOrder())}
		//logrus.Info("sort: ", request.Sort)
	}
	return request
}

const (
	SuggestionTypeMetadata = "metadata"
	SuggestionTypeOperand  = "operand"
	SuggestionTypeKey      = "key"
)

type Suggestion struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Hint  string `json:"hint"`
}

// QuerySuggestions contains the current normalized query and suggestions.
// Concatenating the prefix with any suggestion results in valid query.
// e.g. query 'some data'
// results in prefix 'some' and suggestion: ['metadata:', 'datavalue:']
type QuerySuggestions struct {
	Suggestions []Suggestion `json:"suggestions"`
	Prefix      string       `json:"prefix"`
	ValidQuery  bool         `json:"valid_query"`
}

func (q *QuerySuggestions) addSuggestion(s ...Suggestion) {
	q.Suggestions = append(q.Suggestions, s...)
}

func (q *QuerySuggestions) addSuggestionValues(value, suggestionType, hint string) {
	q.Suggestions = append(q.Suggestions, Suggestion{value, suggestionType, hint})
}

// create suggestions. Does not actually validate the query, just analyzes last token(s).
func suggest(query string, metadata metadataQuerier) *QuerySuggestions {
	qs := &QuerySuggestions{Suggestions: []Suggestion{}}
	if query == "" || query == " " {
		qs.Suggestions = suggestEmpty(metadata)
		return qs
	}

	normalized := strings.ToLower(query)
	tokens := tokenizeFilter(normalized)
	inParantheses := false

	normalizedTokens := []string{}
	if len(tokens) > 1 {
		normalizedTokens = tokens[:len(tokens)-1]
	} else if len(tokens) == 1 {
		normalizedTokens = []string{}
	}

	lastToken := tokens[len(tokens)-1]
	if lastToken == "(" {
		inParantheses = true

		// suggest metadata keys
		keys := metadata.queryKeys("", "", ":")
		for _, v := range keys {
			qs.addSuggestionValues(escapeMetadataKey(v)+":", SuggestionTypeMetadata, "")
		}
		qs.Prefix = query
		return qs
	} else {
		for i := len(tokens) - 1; i > 0; i-- {
			if tokens[i] == "(" {
				//qs.addSuggestion(")")
				inParantheses = true
				break
			}
		}
	}

	keys := []string{"name", "description", "content", "date"}
	operators := []string{"AND", "OR", "NOT"}

	parts := strings.Split(lastToken, ":")
	if len(parts) == 1 {
		// no value yet, suggest key
		for _, v := range keys {
			if strings.Contains(v, parts[0]) {
				qs.addSuggestionValues(v+":", SuggestionTypeKey, "")
			}
		}

		metadataKeys := metadata.queryKeys(parts[0], "", "")
		for _, v := range metadataKeys {
			if strings.Contains(v, " ") {
				v = `"` + v + `"`
			}
			qs.addSuggestionValues(v+":", SuggestionTypeMetadata, "")
		}

		// suggest values too
		values := metadata.queryValues(parts[0], "")
		for i, v := range values {
			if i > 5 {
				// show only 5 keys per key when still typing key
				break
			}
			qs.addSuggestionValues(parts[0]+":"+escapeMetadataValue(v), SuggestionTypeMetadata, "")
		}

		qs.Prefix = strings.Join(normalizedTokens, " ") + " "
	} else if len(parts) == 2 {
		tokenPrefix := escapeMetadataKey(parts[0])
		addWhiteSpace := true
		// suggest value
		if inParantheses && parts[0] != "" {
			// key-value must be non empty before closing parantheses

			for _, v := range operators {
				qs.addSuggestionValues(v, SuggestionTypeOperand, "")
			}
			qs.addSuggestionValues(")", SuggestionTypeOperand, "")
			qs.ValidQuery = false
		}
		if parts[0] == "date" {
			dateSuggestions := suggestDate(parts[1])
			tokenPrefix = "date:"
			if len(dateSuggestions) > 0 {
				//tokenPrefix = "date:"
				addWhiteSpace = false
				for _, v := range dateSuggestions {
					qs.addSuggestionValues(v, SuggestionTypeKey, "")
				}
			}
		} else {

			values := metadata.queryValues(parts[0], parts[1])
			perfectMatch := false
			for _, v := range values {
				if parts[1] == v {
					// don't suggest if perfect match
					tokenPrefix = strings.Join(parts, ":")
					perfectMatch = true
					break
				}
				qs.addSuggestionValues(escapeMetadataValue(v), SuggestionTypeMetadata, "")
			}
			if len(values) > 0 && !perfectMatch {
				tokenPrefix = tokenPrefix + ":"
				addWhiteSpace = false
			} else {

			}
		}
		qs.Prefix = strings.Join(append(normalizedTokens, tokenPrefix), " ")
		if addWhiteSpace {
			qs.Prefix += " "
		}
	}
	if qs.Prefix == "" {
		if len(tokens) < 2 {
			if len(qs.Suggestions) > 0 {
				qs.Prefix = ""
			} else {
				qs.Prefix = query
			}
		} else {
			qs.Prefix = strings.Join(normalizedTokens, " ")
		}
	}

	if len(qs.Suggestions) == 0 {
		qs.Prefix = query
		if query[len(query)-1] == ' ' {
			qs.Suggestions = suggestEmpty(metadata)
		}
	}

	// remove suggestions that are already in query
	for i, v := range qs.Suggestions {
		// doesn't seem to actually remove the items bc of prefixing vs value.
		if lastToken == v.Value {
			if i == 0 {
				qs.Suggestions = qs.Suggestions[i+1:]
			} else if i == len(qs.Suggestions)-1 {
				qs.Suggestions = qs.Suggestions[:i]
			} else {
				qs.Suggestions = append(qs.Suggestions[:i], qs.Suggestions[i+1:]...)
			}
			// item removed from array, compensate index
			i -= 1
		}
	}

	if len(qs.Suggestions) > MaxSuggestions {
		qs.Suggestions = qs.Suggestions[:MaxSuggestions]
	}
	return qs
}

func suggestEmpty(metadata metadataQuerier) []Suggestion {

	keys := []string{"name", "description", "content", "date"}
	results := metadata.queryKeys("", "", ":")

	suggestions := make([]Suggestion, 0, len(keys)+len(results))
	//values := make([]string, 0, len(keys)+len(results))

	for _, v := range keys {
		suggestions = append(suggestions, Suggestion{v, SuggestionTypeKey, ""})
	}
	for _, v := range results {
		suggestions = append(suggestions, Suggestion{v, SuggestionTypeMetadata, ""})
	}

	return suggestions
}

type valueMatchStatus int

const (
	valueMatchStatusInvalid = iota
	valueMatchStatusIncomplete
	valueMatchStatusOk
)

func suggestDate(token string) []string {
	suggestions := []string{}
	keys := []string{"today", "yesterday", "week", "month", "year"}

	now := time.Now().UTC()

	dateCandidates := []string{
		now.Format("2006"),
		now.Format("2006-1"),
		now.Format("2006-1-2"),
		now.AddDate(-2, 0, 0).Format("2006") + "|",
		now.AddDate(-2, 0, 0).Format("2006") + "|today",
		now.AddDate(-5, 0, 0).Format("2006") + "|" + now.AddDate(-1, 0, 0).Format("2006"),
	}

	if token == "" {
		return append(keys, dateCandidates...)
	}

	for _, v := range keys {
		if strings.Contains(v, token) {
			suggestions = append(suggestions, v)
		}
	}

	return suggestions
}

// matchDate tries to validate and autocomplete date filter
func matchDate(token string) (valueMatchStatus, string, time.Time, time.Time) {
	if token == "" {
		return valueMatchStatusIncomplete, "", time.Time{}, time.Time{}
	}
	if token == "today" {
		return valueMatchStatusOk, "today", models.MidnightForDate(time.Now()), models.MidnightForDate(time.Now().AddDate(0, 0, 1))
	} else if token == "yesterday" {
		return valueMatchStatusOk, "yesterday", models.MidnightForDate(time.Now().AddDate(0, 0, -1)), models.MidnightForDate(time.Now())
	} else if token == "week" {
		return valueMatchStatusOk, "week", models.MidnightForDate(time.Now().AddDate(0, 0, -7)), models.MidnightForDate(time.Now())
	} else if token == "month" {
		return valueMatchStatusOk, "month", models.MidnightForDate(time.Now().AddDate(0, -1, 0)), models.MidnightForDate(time.Now())
	} else if token == "year" {
		return valueMatchStatusOk, "year", models.MidnightForDate(time.Now().AddDate(-1, 0, 0)), models.MidnightForDate(time.Now())
	}

	separator := "|"
	splits := strings.Split(token, separator)
	if len(splits) == 1 {
		t := parseDateFromLayout(token, false, false)
		if !t.IsZero() {
			endDate := parseDateFromLayout(token, true, false)
			return valueMatchStatusOk, token, t, endDate
		}
	}

	if len(splits) == 2 {
		startD := parseDateFromLayout(splits[0], false, false)

		if splits[1] == "" {
			splits[1] = "today"
		}

		stopD := parseDateFromLayout(splits[1], true, false)
		if startD.IsZero() || stopD.IsZero() {
			return valueMatchStatusInvalid, token, startD, stopD
		}
		return valueMatchStatusOk, token, startD, stopD
	}

	layouts := []string{
		//"1/2/2006",
		//"2.1.2006",
		"2006-1-2",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, token)
		if err == nil {
			return valueMatchStatusOk, token, t.UTC(), t.UTC()
		}
	}

	return valueMatchStatusInvalid, "", time.Time{}, time.Time{}
}

var regexYear = regexp.MustCompile(`^(\d{4})$`)
var regexYearMonth = regexp.MustCompile(`^(\d{4}-\d{1,2})$`)
var regexDate = regexp.MustCompile(`^(\d{4}-\d{1,2}-\d{1,2})$`)

func parseDateFromLayout(token string, addDigit bool, removeDigit bool) time.Time {

	if token == "today" {
		return models.MidnightForDate(time.Now())
	} else if token == "yesterday" {
		return models.MidnightForDate(time.Now().AddDate(0, 0, -1))
	} else if token == "week" {
		return models.MidnightForDate(time.Now().AddDate(0, 0, 7))
	} else if token == "month" {
		return models.MidnightForDate(time.Now().AddDate(0, 1, 0))
	} else if token == "year" {
		return models.MidnightForDate(time.Now().AddDate(1, 0, 0))
	}

	t, err := time.Parse("2006", "2006")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(t.Format("2006-01-02"))
	layout := ""
	addYears := 0
	addMonths := 0
	addDays := 0
	if regexYear.MatchString(token) {
		addYears = 1
		layout = "2006"
	} else if regexYearMonth.MatchString(token) {
		addMonths = 1
		layout = "2006-1"
	} else if regexDate.MatchString(token) {
		addDays = 1
		layout = "2006-1-2"
	}

	if layout != "" {
		date, err := time.Parse(layout, token)
		if err == nil {
			if addDigit {
				date = date.AddDate(addYears, addMonths, addDays)
			}
			if removeDigit {
				date = date.AddDate(-addYears, -addMonths, -addDays)
			}
			return models.MidnightForDate(date)
		}
	}
	return time.Time{}
}

type metadataQuerier interface {
	queryKeys(key string, prefis string, suffix string) []string
	queryValues(key, value string) []string
}
