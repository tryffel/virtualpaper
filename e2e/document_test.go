package e2e

import (
	"gopkg.in/h2non/gentleman.v2/plugins/multipart"
	"os"
	"path"
	"testing"
)

func TestUploadDocument(t *testing.T) {
	TestLogin(t)
	dir := "test_data"
	doc := "text-1.txt"
	reader, err := os.Open(path.Join(dir, doc))
	if err != nil {
		t.Errorf("read input file %s: %v", doc, err)
		return
	}
	files := []multipart.FormFile{
		{Name: "file", Reader: reader},
	}

	t.Log("upload document, should return OK")
	req := test.Authorize().client.Post("/api/v1/documents")
	req.Request.DelHeader("content-type")
	req.SetHeader("accept", "multipart/form-data").
		Files(files).Expect(t).Status(200).Done()

	// send same file again, should return 304
	reader, err = os.Open(path.Join(dir, doc))
	if err != nil {
		t.Errorf("read input file %s: %v", doc, err)
		return
	}
	files = []multipart.FormFile{
		{Name: "file", Reader: reader},
	}

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
}
