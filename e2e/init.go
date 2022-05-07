package e2e

import (
	"encoding/json"
	"fmt"
	"gopkg.in/h2non/baloo.v3"
	"testing"
	"tryffel.net/go/virtualpaper/api"
)

const userName = "user"
const userPassword = "user"

var userToken = ""

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

	addMetadataKey(t, "test-1", "testing")
	addMetadataKey(t, "test-2", "testing another")

	addMetadataKeyValues(t, 1, "value-1")
	addMetadataKeyValues(t, 1, "value-2")
	addMetadataKeyValues(t, 2, "value-10")
	addMetadataKeyValues(t, 2, "value-11")
	metadataAdded = true
}

func jsonToBody(data interface{}) string {
	buf, _ := json.Marshal(data)
	return string(buf)
}

func addMetadataKey(t *testing.T, key, comment string) {
	body := api.MetadataKeyRequest{
		Key:     key,
		Comment: comment,
	}

	t.Log("add metadata key", key)
	test.IsJson().Authorize().client.Post("/api/v1/metadata/keys").BodyString(jsonToBody(body)).
		SetHeader("content-type", "application/json").Expect(t).Status(200).Done()
}

func addMetadataKeyValues(t *testing.T, keyId int, value string) {
	body := api.MetadataValueRequest{
		Value:          value,
		Comment:        "",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}

	t.Log("add metadata key-value", keyId, value)

	test.IsJson().Authorize().client.Post(fmt.Sprintf("/api/v1/metadata/keys/%d/values", keyId)).
		BodyString(jsonToBody(body)).Expect(t).Status(200).Done()
}
