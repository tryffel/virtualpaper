package integrationtest

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/baloo.v3"
	"gopkg.in/h2non/gentleman.v2/plugins/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type UploadDocumentSuite struct {
	ApiTestSuite
}

func TestUploadDocument(t *testing.T) {
	suite.Run(t, new(UploadDocumentSuite))
}

func (suite *UploadDocumentSuite) SetupTest() {
	suite.Init()
	clearDbMetadataTables(suite.T(), suite.db)
	clearDbDocumentTables(suite.T(), suite.db)
}

func (suite *UploadDocumentSuite) TestUploadFail() {
	suite.publicHttp.Post("/api/v1/documents").Json(suite.T(), "").Expect(suite.T()).e.Status(401).Done()
	docId := uploadDocument(suite.T(), suite.userClient, "pdf-1.pdf", "Lorem ipsum", 20)
	assert.NotEqual(suite.T(), "", docId, "document id not empty")

	doc := getDocument(suite.T(), suite.adminHttp, docId, 404)
	assert.Nil(suite.T(), doc, "another user can't get the document")
	doc = getDocument(suite.T(), suite.userHttp, docId, 200)

	assert.NotNil(suite.T(), doc, "get document")
	assert.Equal(suite.T(), doc.Name, "pdf-1.pdf", "document name")

	uploadDocumentWithStatus(suite.T(), suite.userClient, "pdf-1-illegal.png", "Lorem ipsum", 20, 400)
}

func (suite *UploadDocumentSuite) TestEditDocFail() {
	_ = insertTestDocuments(suite.T(), suite.db)
	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	doc.Name = "empty"
	updateDocument(suite.T(), suite.userHttp, doc, 200)

	doc.Name = ""
	updateDocument(suite.T(), suite.userHttp, doc, 400)

	// too long name
	doc.Name = testDocumentX86Intel.Content
	updateDocument(suite.T(), suite.userHttp, doc, 400)

	doc.Name = "valid"
	updateDocument(suite.T(), suite.userHttp, doc, 200)

	// unsafe name succeeds but is being sanitized
	doc.Filename = "unsafe/filename//.md"
	updateDocument(suite.T(), suite.userHttp, doc, 200)

	unsafeDocName := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), ".md", unsafeDocName.Filename)

	doc.Filename = "file.txt"
	updateDocument(suite.T(), suite.userHttp, doc, 200)

	doc.Date = -1
	updateDocument(suite.T(), suite.userHttp, doc, 400)
}

func (suite *UploadDocumentSuite) TestUploadTxt() {
	uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 60)
}

func (suite *UploadDocumentSuite) TestUploadImage1() {
	uploadDocument(suite.T(), suite.userClient, "jpg-1.jpg", "Lorem ipsum", 60)
	uploadDocument(suite.T(), suite.userClient, "png-1.png", "Lorem ipsum", 60)
}

func (suite *UploadDocumentSuite) TestUploadImage2() {
	// different mimetype, but image is same, server returns existing image instead
	uploadDocument(suite.T(), suite.userClient, "jpg-1.jpeg", "Lorem ipsum", 60)
}

func (suite *UploadDocumentSuite) TestUploadPdf() {
	uploadDocument(suite.T(), suite.userClient, "pdf-1.pdf", "Lorem ipsum", 20)
}

func (suite *UploadDocumentSuite) TestUploadSchedulesProcessing() {
	docId := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 20)
	time.Sleep(time.Second * 2)

	logs := getDocumentProcessingSteps(suite.T(), suite.userHttp, docId, 200)
	assert.Equal(suite.T(), 5, len(*logs))
	assert.True(suite.T(), strings.HasPrefix((*logs)[0].Message, "hash"), "1st step is hash")
	assert.Equal(suite.T(), (*logs)[0].Status, models.JobFinished)
	assert.True(suite.T(), strings.HasPrefix((*logs)[1].Message, "generate thumbnail"), "2nd step is thumbnail")
	assert.Equal(suite.T(), (*logs)[1].Status, models.JobFinished)
	assert.True(suite.T(), strings.HasPrefix((*logs)[2].Message, "extract content"), "3rd step is extract")
	assert.Equal(suite.T(), (*logs)[2].Status, models.JobFinished)
	assert.True(suite.T(), strings.HasPrefix((*logs)[3].Message, "process user rules"), "4th step is rules")
	assert.Equal(suite.T(), (*logs)[3].Status, models.JobFinished)
	assert.True(suite.T(), strings.HasPrefix((*logs)[4].Message, "index for search engine"), "5th step is indexing")
	assert.Equal(suite.T(), (*logs)[4].Status, models.JobFinished)
}

func (suite *UploadDocumentSuite) TestEditSchedulesProcessing() {
	docId := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 20)
	doc := getDocument(suite.T(), suite.userHttp, docId, 200)
	logs := getDocumentProcessingSteps(suite.T(), suite.userHttp, docId, 200)
	assert.Equal(suite.T(), 5, len(*logs))

	editedDoc := updateDocument(suite.T(), suite.userHttp, doc, 200)
	time.Sleep(time.Millisecond * 500)
	logs = getDocumentProcessingSteps(suite.T(), suite.userHttp, docId, 200)
	assert.Equal(suite.T(), 6, len(*logs))

	assert.Equal(suite.T(), doc.Id, editedDoc.Id, "id matches")
	assert.NotEqual(suite.T(), doc.CreatedAt, editedDoc.UpdatedAt, "updated_at is updated")

	doc.Name = "new name"
	doc.Description = "descriptive text"
	// Fri, 18 Nov 2022 15:54:00 GMT
	doc.Date = time.Unix(1668786840, 0).UnixNano() / 1000000

	editedDoc = updateDocument(suite.T(), suite.userHttp, doc, 200)

	// give server some time to finish processing
	time.Sleep(time.Second * 2)

	assert.Equal(suite.T(), "new name", editedDoc.Name, "name")
	assert.Equal(suite.T(), "descriptive text", editedDoc.Description, "description")
	assert.Equal(suite.T(), int64(1668786840000), editedDoc.Date, "date")

	history := getDocumentHistory(suite.T(), suite.userHttp, doc.Id, 200)
	assert.Equal(suite.T(), 4, len(*history))

	logs = getDocumentProcessingSteps(suite.T(), suite.userHttp, docId, 200)

	assert.Equal(suite.T(), 7, len(*logs))
	assert.Equal(suite.T(), (*logs)[5].Status, models.JobFinished)
}

func (suite *UploadDocumentSuite) TestUploadDuplicate() {
	// one user cannot upload duplicate
	uploadDocumentWithStatus(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 60, 200)
	uploadDocumentWithStatus(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 60, 400)

	// another user can upload the same document
	uploadDocumentWithStatus(suite.T(), suite.adminClient, "text-1.txt", "Lorem ipsum", 60, 200)
	uploadDocumentWithStatus(suite.T(), suite.adminClient, "text-1.txt", "Lorem ipsum", 60, 400)
}

func uploadDocument(t *testing.T, client *baloo.Client, fileName string, contentPrefix string, timeoutS int) string {
	return uploadDocumentWithStatus(t, client, fileName, contentPrefix, timeoutS, 200)
}

func uploadDocumentWithStatus(t *testing.T, client *baloo.Client, fileName string, contentPrefix string, timeoutS int, httpStatus int) string {
	dir := "testdata"
	reader, err := os.Open(path.Join(dir, fileName))
	if err != nil {
		t.Errorf("read input file %s: %v", fileName, err)
		return ""
	}
	files := []multipart.FormFile{
		{Name: fileName, Reader: reader},
	}
	form := multipart.FormData{
		Data:  multipart.DataFields{},
		Files: files,
	}
	form.Data["name"] = []string{fileName}

	var docId = ""
	docBody := &api.DocumentResponse{}
	docReady := false

	t.Log("upload document, should return OK")
	req := client.Post("/api/v1/documents")
	req.Request.DelHeader("content-type")
	req.SetHeader("accept", "multipart/form-data").
		Form(form).
		Expect(t).
		AssertFunc(assertHttpCode(t, httpStatus, true, true)).
		AssertFunc(func(r *http.Response, w *http.Request) error {
			err := json.NewDecoder(r.Body).Decode(&docBody)
			if err != nil {
				t.Errorf("parse response json: %v", err)
			}
			docId = docBody.Id
			return nil
		}).Done()

	startTime := time.Now()

	if httpStatus == 200 {
		t.Log("start polling document status")
		for {
			client.Get("/api/v1/documents/" + docId).Expect(t).Status(200).AssertFunc(
				func(r *http.Response, w *http.Request) error {
					err := json.NewDecoder(r.Body).Decode(&docBody)
					if err != nil {
						t.Errorf("parse response json: %v", err)
					}

					if docBody.Status == "ready" {
						docReady = true
						t.Log("document ready")
					} else {
						t.Log("polling document, status", docBody.Status)
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
			time.Sleep(time.Second * 2)
		}

		if !strings.HasPrefix(docBody.Content, contentPrefix) {
			t.Errorf("document content does not match")
		}
	}
	return docId
}

func updateDocument(t *testing.T, client *httpClient, doc *api.DocumentResponse, wantHttpStatus int) *api.DocumentResponse {
	dto := &api.DocumentUpdateRequest{
		Name:        doc.Name,
		Description: doc.Description,
		Filename:    doc.Filename,
		Date:        doc.Date,
		Metadata:    make([]api.MetadataRequest, len(doc.Metadata)),
	}

	for i, v := range doc.Metadata {
		dto.Metadata[i] = api.MetadataRequest{
			KeyId:   v.KeyId,
			ValueId: v.ValueId,
		}
	}

	req := client.Put("/api/v1/documents/"+doc.Id).Json(t, doc).Expect(t)
	if wantHttpStatus == 200 {
		out := &api.DocumentResponse{}
		req.Json(t, out).e.Status(200).Done()
		return out
	}
	req.e.Status(wantHttpStatus).Done()
	return nil
}

func updateDocumentMetadata(t *testing.T, client *httpClient, docId string, metadata api.MetadataUpdateRequest, wantHttpStatus int) {
	client.Post(fmt.Sprintf("/api/v1/documents/%s/metadata", docId)).
		Json(t, metadata).req.Expect(t).Status(wantHttpStatus).Done()
}

func getDocumentHistory(t *testing.T, client *httpClient, docId string, wantHttpStatus int) *[]models.DocumentHistory {
	dto := &[]models.DocumentHistory{}
	req := client.Get("/api/v1/documents/" + docId + "/history").Expect(t)
	if wantHttpStatus == 200 {
		req.Json(t, dto).e.Status(200).Done()
		return dto
	} else {
		req.e.Status(wantHttpStatus).Done()
		return nil
	}
}

func getDocumentProcessingSteps(t *testing.T, client *httpClient, docId string, wantHttpStatus int) *[]models.Job {
	dto := &[]models.Job{}
	req := client.Get("/api/v1/documents/" + docId + "/jobs").Expect(t)
	if wantHttpStatus == 200 {
		req.Json(t, dto).e.Status(200).Done()
		return dto
	} else {
		req.e.Status(wantHttpStatus).Done()
		return nil
	}
}
