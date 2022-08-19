package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"tryffel.net/go/virtualpaper/models"
)

func TestMetadata(t *testing.T) {
	apiTest(t)

	DoLogin(t)

	addMetadata(t)
	testGetMetadataKeys(t)
}

func testGetMetadataKeys(t *testing.T) {
	apiTest(t)
	TestLogin(t)

	keys := &[]models.MetadataKey{}
	test.Authorize().client.Get(`/api/v1/metadata/keys`).
		Params(map[string]string{
			"filter":    "{}",
			"page":      "1",
			"page_size": "10",
			"sort":      "[\"match_filter\", \"ASC\"]",
		}).
		Expect(t).Status(200).
		AssertFunc(func(r *http.Response, w *http.Request) error {

			err := json.NewDecoder(r.Body).Decode(&keys)
			if err != nil {
				t.Errorf("parse response json: %v", err)
			}

			lastId := 0
			if len(*keys) == 0 {
				t.Errorf("no keys returned")
				return nil
			}

			for _, v := range *keys {
				if v.Id <= lastId {
					t.Errorf("ids not ascending")
					return nil
				}
				lastId = v.Id
			}
			return nil
		}).Done()

	keys = &[]models.MetadataKey{}
	test.Authorize().client.Get(`/api/v1/metadata/keys`).
		Params(map[string]string{
			"filter":    "{}",
			"page":      "1",
			"page_size": "1",
			"sort":      "[\"key\", \"DESC\"]",
		}).
		Expect(t).Status(200).
		AssertFunc(func(r *http.Response, w *http.Request) error {

			err := json.NewDecoder(r.Body).Decode(&keys)
			if err != nil {
				t.Errorf("parse response json: %v", err)
			}

			lastId := 1000
			if len(*keys) == 0 {
				t.Errorf("no keys returned")
				return nil
			}

			for _, v := range *keys {
				if v.Id >= lastId {
					t.Errorf("ids not ascending")
					return nil
				}
				lastId = v.Id
			}
			return nil
		}).Done()
}
