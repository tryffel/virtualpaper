package integrationtest

import (
	"fmt"
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

func getDocumentWithVisit(t *testing.T, client *httpClient, id string, wantHttpStatus int) *api.DocumentResponse {
	doc := &api.DocumentResponse{}
	req := client.Get(fmt.Sprintf("/api/v1/documents/%s", id)).SetQueryParam("visit", "1").Expect(t)
	if wantHttpStatus == 200 {
		req.Json(t, doc).e.Status(200).Done()
		return doc
	} else {
		req.e.Status(wantHttpStatus).Done()
		return nil
	}
}

func getDocuments(t *testing.T, client *httpClient, wantHttpStatus int) *[]api.DocumentResponse {
	doc := &[]api.DocumentResponse{}
	req := client.Get("/api/v1/documents/").Expect(t)
	if wantHttpStatus == 200 {
		req.Json(t, doc).e.Status(200).Done()
		return doc
	} else {
		req.e.Status(wantHttpStatus).Done()
		return nil
	}
}

func getDeletedDocuments(t *testing.T, client *httpClient, wantHttpStatus int) *[]api.DocumentResponse {
	doc := &[]api.DocumentResponse{}
	req := client.Get("/api/v1/documents/deleted").Expect(t)
	if wantHttpStatus == 200 {
		req.Json(t, doc).e.Status(200).Done()
		return doc
	} else {
		req.e.Status(wantHttpStatus).Done()
		return nil
	}
}
