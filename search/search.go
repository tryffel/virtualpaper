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
	"github.com/meilisearch/meilisearch-go"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
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
		metadata := parseFilter(d.Metadata)
		if filter != "" {
			filter += " AND "
		}
		filter += metadata
	}
	if filter != "" {
		request.Filter = filter
	}

	if d.Sort != "" {
		if d.SortMode == "" {
			d.SortMode = "desc"
		}
		request.Sort = []string{d.Sort + ":" + d.SortMode}
	}
	return request
}

// SearchDocuments searches documents for given user. Query can be anything. If field="", search in any field,
// else search only specified field
func (e *Engine) SearchDocuments(userId int, query *DocumentFilter, paging storage.Paging) ([]*models.Document, int, error) {

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
	if len(res.Hits) < paging.Limit {
		nHits = len(res.Hits)
	}
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

// parse user filter into meilisearch metadata filter
func parseFilter(filter string) string {
	if filter == "" {
		return filter
	}

	inEscape := false
	textLeft := filter

	// sweep through the filter, remove whitespaces in sentences with underscore
	for i := 0; i < len(filter); i++ {
		if textLeft == "" {
			break
		}
		character := filter[i]
		if inEscape {
			if character == '"' {
				inEscape = false
				continue
			}
			if character == ' ' {
				textLeft = textLeft[:i] + "_" + textLeft[i+1:]
			}
		} else {
			if character == '"' {
				inEscape = true
			}
		}
	}

	// ensure parantheses are tokenized
	textLeft = strings.Replace(textLeft, "(", " ( ", -1)
	textLeft = strings.Replace(textLeft, ")", " ) ", -1)

	tokens := strings.Split(textLeft, " ")
	output := ""

	// join tokens back, removing escapes strings
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if strings.Contains(token, ":") {
			token = strings.Replace(token, "\"", "", -1)
			token = "metadata=\"" + token + "\""
		}
		if output != "" {
			output += " "
		}
		output += token
	}
	return output
}
