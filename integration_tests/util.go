package integrationtest

import (
	"testing"
	"tryffel.net/go/virtualpaper/api"
)

func getDocument(t *testing.T, client *httpClient, id string, wantHttpStatus int) *api.DocumentResponse {
	doc := &api.DocumentResponse{}
	req := client.Get("/api/v1/documents/" + id).Expect(t)
	if wantHttpStatus == 200 {
		req.Json(t, doc).e.Status(200).Done()
		return doc
	} else {
		req.e.Status(wantHttpStatus).Done()
		return nil
	}
}
