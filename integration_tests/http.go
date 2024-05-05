package integrationtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/baloo.v3"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"tryffel.net/go/virtualpaper/errors"
)

type httpClient struct {
	client *baloo.Client
}

type httpRequest struct {
	req *baloo.Request
}

type httpResponse struct {
	e *baloo.Expect
}

func (h *httpClient) Get(url string) *httpRequest {
	return &httpRequest{h.client.Get(url)}
}

func (h *httpClient) Put(url string) *httpRequest {
	return &httpRequest{h.client.Put(url)}
}

func (h *httpClient) Post(url string) *httpRequest {
	return &httpRequest{h.client.Post(url)}
}

func (h *httpClient) Delete(url string) *httpRequest {
	return &httpRequest{h.client.Delete(url)}
}

func (h *httpRequest) Json(t *testing.T, data interface{}) *httpRequest {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		t.Errorf("decode json: %v", err)
	}
	return &httpRequest{h.req.SetHeader("Content-Type", "application/json").Body(buf)}
}

func (h *httpRequest) Sort(key, order string) *httpRequest {
	r := h.req.SetQuery("sort", key).SetQuery("order", order)
	return &httpRequest{r}
}

func (h *httpRequest) Page(page, perPage int) *httpRequest {
	r := h.req.SetQuery("page", strconv.Itoa(page)).SetQuery("perPage", strconv.Itoa(perPage))
	return &httpRequest{r}
}

func (h *httpRequest) Expect(t *testing.T) *httpResponse {
	req := h.req.Expect(t).AssertFunc(logRequestFunc(t, "", false))
	return &httpResponse{req}
}

func (h *httpRequest) ExpectName(t *testing.T, name string, logResp bool) *httpResponse {
	req := h.req.Expect(t).AssertFunc(logRequestFunc(t, name, logResp))
	return &httpResponse{req}
}

func (h *httpRequest) SetQueryParam(key, value string) *httpRequest {
	return &httpRequest{h.req.SetQuery(key, value)}
}

func (h *httpResponse) Json(t *testing.T, data interface{}) *httpResponse {
	return &httpResponse{h.e.AssertFunc(readBodyFunc(t, data))}
}

func (h *httpResponse) AssertError(t *testing.T, message string) *httpResponse {
	return &httpResponse{h.e.AssertFunc(readError(t, message))}
}

func (h *httpResponse) AssertStatus(t *testing.T, wantHttpStatus int) *httpResponse {
	return &httpResponse{h.e.AssertFunc(assertStatusFunc(t, wantHttpStatus))}
}

func (h *httpResponse) Done(t *testing.T) {
	err := h.e.Done()
	if err != nil {
		t.Errorf("http: %v", err)
	}
}

func assertStatusFunc(t *testing.T, wantHttpStatus int) func(r *http.Response, w *http.Request) error {
	return func(r *http.Response, w *http.Request) error {
		assert.Equal(t, wantHttpStatus, r.StatusCode, "http status code")
		if r.StatusCode == wantHttpStatus {
			return nil
		}

		b, err := readBody(r)
		if err != nil {
			t.Errorf("read body: %v", err)
		} else if len(b) > 0 {
			var out bytes.Buffer
			err = json.Indent(&out, b, "", "\t")
			out.WriteTo(os.Stdout)
		}
		return fmt.Errorf("http status want %d, got %d", wantHttpStatus, r.StatusCode)
	}
}

func readError(t *testing.T, message string) func(r *http.Response, w *http.Request) error {
	return func(r *http.Response, w *http.Request) error {
		if r.StatusCode == 304 {
			return nil
		}

		type Error struct {
			Error string `json:"Error"`
		}
		data := &Error{}
		b, err := readBody(r)
		err = json.Unmarshal(b, &data)
		if err != nil {
			t.Errorf("parse response json: %v", err)
		}
		assert.Equal(t, data.Error, message)
		return nil
	}
}

func readBodyFunc(t *testing.T, dto interface{}) func(r *http.Response, w *http.Request) error {
	return func(r *http.Response, w *http.Request) error {
		data, err := readBody(r)
		err = json.Unmarshal(data, &dto)
		if err != nil {
			t.Errorf("parse response json: %v", err)
			t.Error(string(data))
		}
		return nil
	}
}

func logRequestFunc(t *testing.T, name string, logResp bool) func(r *http.Response, w *http.Request) error {
	return func(r *http.Response, w *http.Request) error {
		// replace reader with re-readable one
		_, _ = initBody(r)

		var suffix = ""
		if name != "" {
			suffix = " " + name
		}
		if logResp {
			//data, _ := io.ReadAll(r.Body)
			data, _ := readBody(r)
			t.Logf("%s %s %d%s data: %s", w.Method, w.URL.String(), r.StatusCode, suffix, string(data))
		} else {
			t.Logf("%s %s %d%s\n", w.Method, w.URL.String(), r.StatusCode, suffix)
		}
		return nil
	}
}

func queryParams(req *baloo.Request, filter string, page int, pageSize int, sortKey string, sortOrder string) *baloo.Request {
	return req.Params(map[string]string{
		"filter":  filter,
		"page":    strconv.Itoa(page),
		"perPage": strconv.Itoa(pageSize),
		"sort":    sortKey,
		"order":   sortOrder,
	})
}

func assertHttpCode(t *testing.T, wantCode int, logInvalid, fail bool) func(r *http.Response, w *http.Request) error {
	return func(r *http.Response, w *http.Request) error {
		assert.Equal(t, wantCode, r.StatusCode, "http status code")
		if r.StatusCode != wantCode {
			msg := fmt.Sprintf("invalid status code: want %d, got %d", wantCode, r.StatusCode)
			if logInvalid {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read body: %v", err)
				} else {
					msg += fmt.Sprintf(" body: %v", string(body))
				}
			}
			t.Error(msg)
			if fail {
				t.FailNow()
			}
			return errors.New(msg)
		}
		return nil
	}
}

type R struct {
	b      []byte
	reader *bytes.Reader
}

func (r *R) Read(p []byte) (n int, err error) {
	/*
		_, err = r.reader.Seek(0, io.SeekStart)
		if err != nil {
			log.Println(err)
		}

	*/
	return r.reader.Read(p)
}

func (r *R) Close() error {
	r.reader = bytes.NewReader(r.b)
	return nil
}

func readBody(resp *http.Response) ([]byte, error) {
	data, err := io.ReadAll(resp.Body)
	return data, err
}

func initBody(resp *http.Response) ([]byte, error) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	err = resp.Body.Close()
	if err != nil {
		log.Println(err)

	}

	reader := bytes.NewReader(data)
	r := R{data, reader}
	resp.Body = &r
	return data, nil
}
