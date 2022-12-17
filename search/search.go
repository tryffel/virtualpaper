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

package search

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/meilisearch/meilisearch-go"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

// DocumentFilter defines filter for searching/filtering documents
type DocumentFilter struct {
	Query    string    `json:"q"`
	Tag      string    `json:"tag"`
	After    time.Time `json:"after"`
	Before   time.Time `json:"before"`
	Metadata string    `json:"metadata"`
	Sort     string    `json:"sort"`
	SortMode string    `json:"sort_mode"`
}

func (d *DocumentFilter) buildRequest(paging storage.Paging) *meilisearch.SearchRequest {
	request := &meilisearch.SearchRequest{
		Offset:                int64(paging.Offset),
		Limit:                 int64(paging.Limit),
		AttributesToRetrieve:  []string{"document_id", "name", "content", "description", "date"},
		AttributesToCrop:      []string{"content"},
		CropLength:            1000,
		AttributesToHighlight: []string{"content", "name", "description"},
		Matches:               false,
		FacetsDistribution:    nil,
		PlaceholderSearch:     false,
	}
	filter := ""
	if d.Query == "" {
		request.PlaceholderSearch = true
	}

	facets := make([]interface{}, 0)
	if d.Tag != "" {
		facets = append(facets, fmt.Sprintf(`tags:%s`, d.Tag))
	}
	if !d.After.IsZero() {
		filter += fmt.Sprintf("date > %d", d.After.Add(-time.Hour*24).Unix())
	}
	if !d.Before.IsZero() {
		if filter != "" {
			filter += " AND "
		}
		filter += fmt.Sprintf("date < %d", d.Before.Add(time.Hour*24).Unix())
	}
	if d.Metadata != "" {
		metadata := parseMetadataFilter(d.Metadata)
		if filter != "" {
			filter += " AND "
		}
		filter += metadata
	}
	if filter != "" {
		request.Filter = filter
	}

	if d.Sort != "" && d.Sort != "id" {
		if d.SortMode == "" {
			d.SortMode = "desc"
		}
		request.Sort = []string{d.Sort + ":" + d.SortMode}
	}
	return request
}

// SearchDocumentsV1 searches documents for given user. Query can be anything. If field="", search in any field,
// else search only specified field
func (e *Engine) SearchDocumentsV1(userId int, query *DocumentFilter, paging storage.Paging) ([]*models.Document, int, error) {

	request := query.buildRequest(paging)
	logrus.Debugf("Meilisearch query: %v", request)

	docs := make([]*models.Document, 0)

	res, err := e.client.Index(indexName(userId)).Search(query.Query, request)
	if err != nil {
		if meiliError, ok := err.(*meilisearch.Error); ok {
			if meiliError.StatusCode == 400 {
				logrus.Debugf("meilisearch invalid query: %v", meiliError)
				// invalid query
				userError := errors.ErrInvalid
				userError.ErrMsg = "Invalid query"
				return nil, 0, userError
			} else {
				logrus.Errorf("meilisearch error: %v", meiliError)
			}
		}
		return docs, 0, err
	}
	if len(res.Hits) == 0 {
		return docs, 0, nil
	}

	docs = make([]*models.Document, len(res.Hits))

	for i, v := range res.Hits {
		isMap, ok := v.(map[string]interface{})
		if ok {
			doc := &models.Document{}
			doc.Id = getString("document_id", isMap)
			doc.Name = getString("name", isMap)
			doc.Content = getString("content", isMap)
			doc.Description = getString("description", isMap)
			doc.Date = time.Unix(int64(getInt("date", isMap)), 0)
			docs[i] = doc

			formatted := isMap["_formatted"]
			if formattedMap, ok := formatted.(map[string]interface{}); ok {
				name := getString("name", formattedMap)
				if name != "" {
					doc.Name = name
				}
				content := getString("content", formattedMap)
				if content != "" {
					doc.Content = content
				}
			}

		}
	}
	// If there are only filters and no query, meilisearch returns larger nbHits, probably count of all documents,
	// which is incorrect for given filter.
	nHits := int(res.NbHits)
	return docs, nHits, nil
}

// SearchDocumentsV1 searches documents for given user. Query can be anything. If field="", search in any field,
// else search only specified field
func (e *Engine) SearchDocumentsNew(userId int, query string, sort storage.SortKey, paging storage.Paging) ([]*models.Document, int, error) {

	qs, err := parseFilter(query)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = err.Error()
		return nil, 0, e
	}

	request := qs.prepareMeiliQuery(userId, sort, paging)
	logrus.Debugf("Meilisearch query: %s, %v", qs.Query, request.Filter)

	docs := make([]*models.Document, 0)

	res, err := e.client.Index(indexName(userId)).Search(qs.Query, request)
	if err != nil {
		if meiliError, ok := err.(*meilisearch.Error); ok {
			if meiliError.StatusCode == 400 {
				logrus.Debugf("meilisearch invalid query: %v", meiliError)
				// invalid query
				userError := errors.ErrInvalid
				userError.ErrMsg = "Invalid query"
				return nil, 0, userError
			} else {
				logrus.Errorf("meilisearch error: %v", meiliError)
			}
		}
		return docs, 0, err
	}
	if len(res.Hits) == 0 {
		return docs, 0, nil
	}

	docs = make([]*models.Document, len(res.Hits))

	for i, v := range res.Hits {
		isMap, ok := v.(map[string]interface{})
		if ok {
			doc := &models.Document{}
			doc.Id = getString("document_id", isMap)
			doc.Name = getString("name", isMap)
			doc.Content = getString("content", isMap)
			doc.Description = getString("description", isMap)
			doc.Date = time.Unix(int64(getInt("date", isMap)), 0)
			docs[i] = doc

			formatted := isMap["_formatted"]
			if formattedMap, ok := formatted.(map[string]interface{}); ok {
				name := getString("name", formattedMap)
				if name != "" {
					doc.Name = name
				}
				content := getString("content", formattedMap)
				if content != "" {
					doc.Content = content
				}
			}

		}
	}
	// If there are only filters and no query, meilisearch returns larger nbHits, probably count of all documents,
	// which is incorrect for given filter.
	nHits := int(res.NbHits)
	return docs, nHits, nil
}

func getString(key string, container map[string]interface{}) string {
	val, ok := container[key].(string)
	if !ok {
		return ""
	}
	return val
}

func getInt(key string, container map[string]interface{}) int {
	// int
	intVal, ok := container[key].(int)
	if ok {
		return intVal
	}
	int64Val, ok := container[key].(int64)
	if ok {
		return int(int64Val)
	}
	float32Val, ok := container[key].(float32)
	if ok {
		return int(float32Val)
	}
	float64Val, ok := container[key].(float64)
	if ok {
		return int(float64Val)
	}
	return 0
}

// tokenizes filter: 'a:misc and (topic:"unknown topic")' -> [a:misc, and, (, topic:unknown topic, )]
func tokenizeFilter(filter string) []string {
	if filter == "" {
		return []string{}
	}

	tokens := make([]string, 0, 10)
	escapeChar := '"'
	inEscape := false
	textLeft := filter
	token := ""

	for {
		if textLeft == "" {
			if token != "" {
				tokens = append(tokens, token)
			}
			break
		}
		character, width := utf8.DecodeRuneInString(textLeft)
		if inEscape {
			if character == escapeChar {
				inEscape = false
				textLeft = textLeft[width:]
			} else {
				token += string(character)
				textLeft = textLeft[width:]
			}
		} else {
			if character == escapeChar {
				inEscape = true
				textLeft = textLeft[width:]
			} else if character == ' ' {
				// next token
				tokens = append(tokens, token)
				textLeft = textLeft[width:]
				token = ""
			} else if character == ')' || character == '(' {
				if token != "" {
					tokens = append(tokens, token)
					token = ""
				}
				tokens = append(tokens, string(character))
				textLeft = textLeft[width:]
				textLeft = strings.TrimLeft(textLeft, " ")
			} else {
				token += string(character)
				textLeft = textLeft[width:]
			}
		}
	}

	return tokens
}

// parse user filter into meilisearch metadata filter
func parseMetadataFilter(filter string) string {
	tokens := tokenizeFilter(filter)
	for i, v := range tokens {
		if strings.Contains(v, ":") {
			v = strings.Replace(v, " ", "_", -1)
			v = `metadata="` + v + `"`
			tokens[i] = v
		}
	}
	return strings.Join(tokens, " ")
}

func parseFilter(filter string) (*searchQuery, error) {
	sq := &searchQuery{RawQuery: filter}
	tokens := tokenizeFilter(strings.ToLower(filter))

	operators := []string{"and", "or", "not", "(", ")"}
	metadataQuery := []string{}

	textQuery := []string{}

	matchers := map[string]parseFunc{
		"date":        parseDate,
		"name":        parseName,
		"content":     parseContent,
		"description": parseDescription,
	}

	tokensLeft := tokens
	removeToken := func() {
		tokensLeft = tokensLeft[1:]
	}

	maxIterations := len(tokens)
	iteration := 0
	for iteration < maxIterations && len(tokensLeft) > 0 {
		iteration += 1
		token := tokensLeft[0]
		splits := strings.Split(tokensLeft[0], ":")
		if len(splits) == 1 {
			found := false
			for _, v := range operators {
				if token == v {
					metadataQuery = append(metadataQuery, strings.ToUpper(v))
					removeToken()
					found = true
					break
				}
			}
			if found {
				continue
			}
			textQuery = append(textQuery, token)
			removeToken()
			continue
		}
		found := false
		for key, matcher := range matchers {

			if splits[0] == key {
				ok := matcher(splits[1], sq)
				if !ok {
					return sq, fmt.Errorf("invalid query: %v", token)
				} else {
					removeToken()
					found = true
					break
				}
			}

		}
		if found {
			continue
		}

		metadataFilter := fmt.Sprintf(`metadata="%s:%s"`, normalizeMetadataKey(splits[0]), normalizeMetadataValue(splits[1]))
		metadataQuery = append(metadataQuery, metadataFilter)
		removeToken()
	}
	sq.MetadataQuery = metadataQuery
	sq.MetadataString = strings.Join(metadataQuery, " ")
	sq.Query = strings.Join(textQuery, " ")
	return sq, nil
}

type parseFunc func(value string, sq *searchQuery) bool

func parseDate(value string, sq *searchQuery) bool {
	status, _, startT, endT := matchDate(value)
	if status == valueMatchStatusOk {
		sq.DateAfter = startT
		sq.DateBefore = endT
		return true
	}
	return false
}

func parseName(value string, sq *searchQuery) bool {
	sq.Name = value
	return true
}

func parseContent(value string, sq *searchQuery) bool {
	sq.Content = value
	return true
}

func parseDescription(value string, sq *searchQuery) bool {
	sq.Description = value
	return true
}

func normalizeMetadataKey(key string) string {
	return strings.Replace(key, " ", "_", -1)
}

func normalizeMetadataValue(value string) string {
	return normalizeMetadataKey(value)
}
