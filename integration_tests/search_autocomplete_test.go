package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/services/search"
)

type SearchAutocompleteTestSuite struct {
	ApiTestSuite
}

func TestSearchAutocomplete(t *testing.T) {
	suite.Run(t, new(SearchAutocompleteTestSuite))
}

func (suite *SearchAutocompleteTestSuite) SetupTest() {
	suite.Init()
	initMetadataKeyValues(suite.T(), suite.userHttp)
}

func (suite *SearchAutocompleteTestSuite) TestAutocomplete() {
	year, month, day := time.Now().Date()
	yearStr := fmt.Sprintf("%d", year)
	monthStr := fmt.Sprintf("%d-%d", year, month)
	dayStr := fmt.Sprintf("%d-%d-%d", year, month, day)

	completion := getSearchAutocompletions(suite.T(), suite.userHttp, "", 200)
	assert.NotNil(suite.T(), completion)
	assert.False(suite.T(), completion.ValidQuery)
	assert.Equal(suite.T(), "", completion.Prefix)

	suite.assertCommonAutocompleteSuggestions(completion)

	completion = getSearchAutocompletions(suite.T(), suite.userHttp, "da", 200)
	assert.False(suite.T(), completion.ValidQuery)

	assertAutocompleteInSuggestions(suite.T(), completion, "date:", "key", "")

	completion = getSearchAutocompletions(suite.T(), suite.userHttp, "date:", 200)
	assertAutocompleteInSuggestions(suite.T(), completion, "today", "key", "")
	assertAutocompleteInSuggestions(suite.T(), completion, "yesterday", "key", "")
	assertAutocompleteInSuggestions(suite.T(), completion, "week", "key", "")
	assertAutocompleteInSuggestions(suite.T(), completion, "month", "key", "")
	assertAutocompleteInSuggestions(suite.T(), completion, "year", "key", "")
	assertAutocompleteInSuggestions(suite.T(), completion, yearStr, "key", "")
	assertAutocompleteInSuggestions(suite.T(), completion, monthStr, "key", "")
	assertAutocompleteInSuggestions(suite.T(), completion, dayStr, "key", "")

	completion = getSearchAutocompletions(suite.T(), suite.userHttp, "date:today a", 200)

	assertAutocompleteInSuggestions(suite.T(), completion, "author:", "metadata", "")
	assertAutocompleteInSuggestions(suite.T(), completion, "category:", "metadata", "")

	completion = getSearchAutocompletions(suite.T(), suite.userHttp, "date:today author", 200)
	assertAutocompleteInSuggestions(suite.T(), completion, "author:", "metadata", "")
	assertAutocompleteInSuggestions(suite.T(), completion, "author:darwin", "metadata", "")
	assertAutocompleteInSuggestions(suite.T(), completion, "author:doyle", "metadata", "")
	assert.Equal(suite.T(), "date:today ", completion.Prefix)
}

func (suite *SearchAutocompleteTestSuite) assertCommonAutocompleteSuggestions(results *search.QuerySuggestions) {
	assertAutocompleteInSuggestions(suite.T(), results, "name", "key", "")
	assertAutocompleteInSuggestions(suite.T(), results, "description", "key", "")
	assertAutocompleteInSuggestions(suite.T(), results, "content", "key", "")
	assertAutocompleteInSuggestions(suite.T(), results, "date", "key", "")
}

func assertAutocompleteInSuggestions(t *testing.T, results *search.QuerySuggestions, suggestionName, suggestionType, suggestionHint string) {
	for _, v := range results.Suggestions {
		if v.Value == suggestionName && v.Type == suggestionType && v.Hint == suggestionHint {
			return
		}
	}
	t.Errorf("search suggestion line does not exist: %s", suggestionName)
}

func getSearchAutocompletions(t *testing.T, client *httpClient, query string, wantHttpStatus int) *search.QuerySuggestions {
	dto := &api.SearchSuggestRequest{Filter: query}
	request := client.Post("/api/v1/documents/search/suggest").Json(t, dto).Expect(t)
	if wantHttpStatus == 200 {
		body := &search.QuerySuggestions{}
		request.Json(t, body).e.Status(200).Done()
		return body
	}
	request.e.Status(wantHttpStatus).Done()
	return nil
}
