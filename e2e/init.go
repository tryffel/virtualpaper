package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"gopkg.in/h2non/baloo.v3"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

const userName = "user"
const userPassword = "user"

var userToken = ""
var adminToken = ""

const adminUser = "admin"
const adminPassw = "admin"

type httpTest struct {
	client *baloo.Client
}

func (t *httpTest) Authorize() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Authorization", "Bearer "+userToken),
	}
}

func (t *httpTest) AuthorizeAdmin() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Authorization", "Bearer "+adminToken),
	}
}

func (t *httpTest) IsJson() *httpTest {
	return &httpTest{
		client: t.client.SetHeader("Content-Type", "application/json"),
	}
}

var test = &httpTest{client: baloo.New("http://localhost:8000")}

func apiTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
}

var metadataAdded = false

func addMetadata(t *testing.T) {
	if metadataAdded {
		return
	}
	t.Log("add metadata")

	if userToken == "" {
		t.Errorf("no user token found")
	}

	_ = addMetadataKey(t, "test-1", "testing")
	_ = addMetadataKey(t, "test-2", "testing another")

	addMetadataKey(t, "country", "Country")

	category := addMetadataKey(t, "category", "Category")
	author := addMetadataKey(t, "author", "Author")

	addMetadataKey(t, "manyvalues", "testing another")

	// empty matchType should convert to 'exact'
	addMetadataKeyValues(t, category, "economy", false, "", "")
	addMetadataKeyValues(t, category, "scientific", false, "exact", "")
	addMetadataKeyValues(t, category, "energy", true, "regex", "(greenhouse)|(gas emission)")

	addMetadataKeyValues(t, author, "gov.uk", true, "exact", "gov.uk")
	addMetadataKeyValues(t, author, "lorem ipsum", true, "exact", "lorem ipsum")

	metadataAdded = true
}

func jsonToBody(data interface{}) string {
	buf, _ := json.Marshal(data)
	return string(buf)
}

func addMetadataKey(t *testing.T, key, comment string) int {
	body := api.MetadataKeyRequest{
		Key:     key,
		Comment: comment,
	}

	id := 0

	t.Log("add metadata key", key)
	test.IsJson().Authorize().client.Post("/api/v1/metadata/keys").BodyString(jsonToBody(body)).
		SetHeader("content-type", "application/json").Expect(t).Status(200).AssertFunc(func(resp *http.Response, req *http.Request) error {
		dto := &models.MetadataKey{}
		err := json.NewDecoder(resp.Body).Decode(dto)
		if err != nil {
			t.Errorf("parse json: %v", err)
		}

		id = dto.Id
		return nil
	}).Done()

	return id
}

func addMetadataKeyValues(t *testing.T, keyId int, value string, matchDocuments bool, matchType string, matchFilter string) {
	body := api.MetadataValueRequest{
		Value:          value,
		Comment:        "",
		MatchDocuments: matchDocuments,
		MatchType:      matchType,
		MatchFilter:    matchFilter,
	}

	t.Log("add metadata key-value", keyId, value)

	test.IsJson().Authorize().client.Post(fmt.Sprintf("/api/v1/metadata/keys/%d/values", keyId)).
		BodyString(jsonToBody(body)).Expect(t).Status(200).Done()
}
