package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"gopkg.in/h2non/gentleman.v2/plugins/multipart"
	"tryffel.net/go/virtualpaper/api"
)

func TestUploadDocument(t *testing.T) {
	TestLogin(t)
	apiTest(t)
	// works with plain text files only. Multipart form.header holds incorrect mime type, which is always plain/text.
	uploadDocument(t, "text-1.txt", "lorem ipsum", 20)
	uploadDocument(t, "1990-2020-final-emissions-standard-industrial-classification-annex-2.txt",
		"This publication is an extension of the final UK territorial greenhouse", 20)
	uploadDocument(t, "subnational_electricity_and_gas_consumption_summary_report_2020.txt",
		"subnational Electricity and Gas Consumption Statistics", 20)

	/*

		t.Log("upload same document again, should return NOT MODIFIED")
		req = test.Authorize().client.Post("/api/v1/documents")
		req.Request.DelHeader("content-type")
		req.SetHeader("accept", "multipart/form-data").
			Files(files).Expect(t).Status(304).Done()

		// no token
		t.Log("upload document without auth token, should return 400")
		req = test.client.Post("/api/v1/documents")
		req.Request.DelHeader("content-type")
		req.SetHeader("accept", "multipart/form-data").
			Expect(t).Status(400).Done()

		// no file found
		t.Log("upload document, document is missing, should return 400")
		req = test.Authorize().client.Post("/api/v1/documents")
		req.Request.DelHeader("content-type")
		req.SetHeader("accept", "multipart/form-data").
			Expect(t).Status(400).Done()

	*/
}

func uploadDocument(t *testing.T, fileName string, contentPrefix string, timeoutS int) {
	dir := "test_data"
	reader, err := os.Open(path.Join(dir, fileName))
	if err != nil {
		t.Errorf("read input file %s: %v", fileName, err)
		return
	}
	files := []multipart.FormFile{
		{Name: "file", Reader: reader},
	}

	var docId = ""
	docBody := &api.DocumentResponse{}
	docReady := false

	t.Log("upload document, should return OK")
	req := test.Authorize().client.Post("/api/v1/documents")
	req.Request.DelHeader("content-type")
	req.SetHeader("accept", "multipart/form-data").
		Files(files).Expect(t).Status(200).
		AssertFunc(func(r *http.Response, w *http.Request) error {
			err := json.NewDecoder(r.Body).Decode(&docBody)
			if err != nil {
				t.Errorf("parse response json: %v", err)
			}
			docId = docBody.Id
			return nil
		}).Done()

	startTime := time.Now()

	for {
		t.Log("poll document status")
		test.Authorize().client.Get("/api/v1/documents/" + docId).Expect(t).Status(200).AssertFunc(
			func(r *http.Response, w *http.Request) error {
				err := json.NewDecoder(r.Body).Decode(&docBody)
				if err != nil {
					t.Errorf("parse response json: %v", err)
				}

				if docBody.Status == "ready" {
					docReady = true
					t.Log("document ready")
				}
				docId = docBody.Id
				return nil
			}).Done()

		if docReady {
			break
		}
		if time.Now().Sub(startTime).Seconds() > float64(timeoutS) {
			t.Errorf("timeout while indexing document")
			break
		}
		time.Sleep(time.Second)
	}

	if !strings.HasPrefix(docBody.Content, contentPrefix) {
		t.Errorf("document content does not match")
	}
}
