package integrationtest

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/baloo.v3"
	"gopkg.in/h2non/gentleman.v2/plugins/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
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
	clearDbMetadataTables(suite.T())
	clearDbDocumentTables(suite.T())
}

func (suite *UploadDocumentSuite) TestUploadFail() {
	suite.publicHttp.Post("/api/v1/documents").Json(suite.T(), "").Expect(suite.T()).e.Status(401).Done()
	docId := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 20)
	assert.NotEqual(suite.T(), "", docId, "document id not empty")

	doc := getDocument(suite.T(), suite.adminHttp, docId, 404)
	assert.Nil(suite.T(), doc, "another user can't get the document")
	doc = getDocument(suite.T(), suite.userHttp, docId, 200)

	assert.NotNil(suite.T(), doc, "get document")
	assert.Equal(suite.T(), doc.Name, "file", "document name")
}

func (suite *UploadDocumentSuite) TestUploadTxt() {
	uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 20)
}

func (suite *UploadDocumentSuite) TestUploadImage1() {
	uploadDocument(suite.T(), suite.userClient, "jpg-1.jpg", "Lorem ipsum", 20)
	uploadDocument(suite.T(), suite.userClient, "png-1.png", "Lorem ipsum", 20)
}

func (suite *UploadDocumentSuite) TestUploadImage2() {
	// different mimetype, but image is same, server returns existing image instead
	uploadDocument(suite.T(), suite.userClient, "jpg-1.jpeg", "Lorem ipsum", 20)
}

func (suite *UploadDocumentSuite) TestUploadPdf() {
	uploadDocument(suite.T(), suite.userClient, "pdf-1.pdf", "Lorem ipsum", 20)
}

func (suite *UploadDocumentSuite) TestEdit() {
	docId := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 20)
	doc := getDocument(suite.T(), suite.userHttp, docId, 200)

	oldDocDate := strconv.Itoa(int(doc.Date / 1000))

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

	assert.Equal(suite.T(), doc.Id, docId)

	history := getDocumentHistory(suite.T(), suite.userHttp, doc.Id, 200)
	assert.Equal(suite.T(), 1, len(*history))
	assert.Equal(suite.T(), docId, (*history)[0].DocumentId)
	assert.Equal(suite.T(), "create", (*history)[0].Action)
	assert.Equal(suite.T(), "", (*history)[0].OldValue)
	assert.Equal(suite.T(), "file", (*history)[0].NewValue)

	_ = updateDocument(suite.T(), suite.adminHttp, doc, 404)

	editedDoc := updateDocument(suite.T(), suite.userHttp, doc, 200)

	assert.Equal(suite.T(), doc.Id, editedDoc.Id, "id matches")
	assert.NotEqual(suite.T(), doc.CreatedAt, editedDoc.UpdatedAt, "updated_at is updated")

	doc.Name = "new name"
	doc.Description = "descriptive text"
	// Fri, 18 Nov 2022 15:54:00 GMT
	doc.Date = time.Unix(1668786840, 0).UnixNano() / 1000000

	editedDoc = updateDocument(suite.T(), suite.userHttp, doc, 200)

	newDocDate := strconv.Itoa(int(doc.Date / 1000))

	assert.Equal(suite.T(), "new name", editedDoc.Name, "name")
	assert.Equal(suite.T(), "descriptive text", editedDoc.Description, "description")
	assert.Equal(suite.T(), int64(1668786840000), editedDoc.Date, "date")

	history = getDocumentHistory(suite.T(), suite.userHttp, doc.Id, 200)
	assert.Equal(suite.T(), 4, len(*history))

	assert.Equal(suite.T(), "create", (*history)[0].Action)
	assert.Equal(suite.T(), docId, (*history)[1].DocumentId)
	assert.Equal(suite.T(), "rename", (*history)[1].Action)
	assert.Equal(suite.T(), "file", (*history)[1].OldValue)
	assert.Equal(suite.T(), "new name", (*history)[1].NewValue)

	assert.Equal(suite.T(), docId, (*history)[2].DocumentId)
	assert.Equal(suite.T(), "description", (*history)[2].Action)
	assert.Equal(suite.T(), "", (*history)[2].OldValue)
	assert.Equal(suite.T(), "descriptive text", (*history)[2].NewValue)

	assert.Equal(suite.T(), docId, (*history)[3].DocumentId)
	assert.Equal(suite.T(), "date", (*history)[3].Action)
	assert.Equal(suite.T(), oldDocDate, (*history)[3].OldValue)
	assert.Equal(suite.T(), newDocDate, (*history)[3].NewValue)

	doc.Description = "yet another description"
	editedDoc = updateDocument(suite.T(), suite.userHttp, doc, 200)
	history = getDocumentHistory(suite.T(), suite.userHttp, doc.Id, 200)
	assert.Equal(suite.T(), 5, len(*history))
	assert.Equal(suite.T(), docId, (*history)[3].DocumentId)
	assert.Equal(suite.T(), "date", (*history)[3].Action)
	assert.Equal(suite.T(), oldDocDate, (*history)[3].OldValue)
	assert.Equal(suite.T(), newDocDate, (*history)[3].NewValue)
	assert.Equal(suite.T(), "description", (*history)[4].Action)
	assert.Equal(suite.T(), "descriptive text", (*history)[4].OldValue)
	assert.Equal(suite.T(), "yet another description", (*history)[4].NewValue)

	logs = getDocumentProcessingSteps(suite.T(), suite.userHttp, docId, 200)

	// sometimes additoinal indexing step is created. Amount is thus 6 or 7 steps
	assert.GreaterOrEqual(suite.T(), len(*logs), 6, "6-7 steps")
	assert.LessOrEqual(suite.T(), len(*logs), 7, "6-7 steps")
	assert.True(suite.T(), strings.HasPrefix((*logs)[5].Message, "index for search engine"))
	assert.Equal(suite.T(), (*logs)[5].Status, models.JobFinished)
}

func uploadDocument(t *testing.T, client *baloo.Client, fileName string, contentPrefix string, timeoutS int) string {
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
		Expect(t).Status(200).
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
		client.Get("/api/v1/documents/" + docId).Expect(t).Status(200).AssertFunc(
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
