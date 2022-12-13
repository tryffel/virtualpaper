package integrationtest

import (
	"bytes"
	"encoding/json"
	"gopkg.in/h2non/baloo.v3"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
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
	return &httpRequest{h.req.SetHeader("Content-Type", "Application/json").Body(buf)}
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
	req := h.req.Expect(t).AssertFunc(logRequestFunc(t, ""))
	return &httpResponse{req}
}

func (h *httpRequest) ExpectName(t *testing.T, name string) *httpResponse {
	req := h.req.Expect(t).AssertFunc(logRequestFunc(t, name))
	return &httpResponse{req}
}

func (h *httpResponse) Json(t *testing.T, data interface{}) *httpResponse {
	return &httpResponse{h.e.AssertFunc(readBodyFunc(t, data))}
}

func readBodyFunc(t *testing.T, dto interface{}) func(r *http.Response, w *http.Request) error {
	return func(r *http.Response, w *http.Request) error {
		data, err := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(data, &dto)
		if err != nil {
			t.Errorf("parse response json: %v", err)
			t.Error(string(data))
		}
		return nil
	}
}

func logRequestFunc(t *testing.T, name string) func(r *http.Response, w *http.Request) error {
	return func(r *http.Response, w *http.Request) error {
		var suffix = ""
		if name != "" {
			suffix = " " + name
		}
		t.Logf("%s %s %d%s\n", w.Method, w.URL.String(), r.StatusCode, suffix)
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
