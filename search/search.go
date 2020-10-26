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
	"time"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

// DocumentFilter defines filter for searching/filtering documents
type DocumentFilter struct {
	Query  string    `json:"q"`
	Tag    string    `json:"tag"`
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

// SearchDocument searches documents for given user. Query can be anything. If field="", search in any field,
// else search only specified field
func (e *Engine) SearchDocuments(userId int, query *DocumentFilter, paging storage.Paging) ([]*models.Document, int, error) {

	request := meilisearch.SearchRequest{
		Query:                 query.Query,
		Offset:                int64(paging.Offset),
		Limit:                 int64(paging.Limit),
		AttributesToRetrieve:  []string{"document_id", "name", "content", "description"},
		AttributesToCrop:      []string{"content"},
		CropLength:            1000,
		AttributesToHighlight: []string{"content", "name"},
		Filters:               "",
		Matches:               false,
		FacetsDistribution:    nil,
		FacetFilters:          nil,
		PlaceholderSearch:     false,
	}

	if query.Query == "" {
		request.PlaceholderSearch = true
	}

	facets := []interface{}{}

	if query.Tag != "" {
		facets = append(facets, fmt.Sprintf(`tags:%s`, query.Tag))
	}
	if !query.After.IsZero() {
		request.Filters += fmt.Sprintf("date > %d", query.After.Add(-time.Hour*24).Unix())
	}
	if !query.Before.IsZero() {
		if request.Filters != "" {
			request.Filters += " AND "
		}
		request.Filters += fmt.Sprintf("date < %d", query.Before.Add(time.Hour*24).Unix())
	}

	if len(facets) != 0 {
		request.FacetFilters = facets
	}

	logrus.Debugf("Meilisearch query: %v", request)

	docs := make([]*models.Document, 0)

	res, err := e.client.Search(indexName(userId)).Search(request)
	if err != nil {
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
			doc.Id = getInt("document_id", isMap)
			doc.Name = getString("name", isMap)
			doc.Content = getString("content", isMap)
			doc.Description = getString("description", isMap)
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
	return docs, int(res.NbHits), nil
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
