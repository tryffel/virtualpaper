package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

func AddMetadataKey(t *testing.T, client *httpClient, key string, comment string, wantHttpStatus int) *models.MetadataKey {
	dto := &api.MetadataKeyRequest{
		Key:     key,
		Comment: comment,
	}

	req := client.Post("/api/v1/metadata/keys").Json(t, dto)
	body := &models.MetadataKey{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, body).e.Status(200).Done()
		assert.Greaterf(t, body.Id, 0, "id > 0")
		assert.Equal(t, body.Key, key, "key")
		assert.Equal(t, body.Comment, comment, "comment")
		//assert.Equal(t, body.NumDocuments, 0, "num documents")
		assert.True(t, isToday(body.CreatedAt), 0, "timestamp today")
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
	}
	return body
}

func GetMetadataKeys(t *testing.T, client *httpClient, wantHttpStatus int, editFunc func(request *httpRequest) *httpRequest) *[]models.MetadataKey {
	req := client.Get("/api/v1/metadata/keys")
	if editFunc != nil {
		req = editFunc(req)
	}
	dto := &[]models.MetadataKey{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, dto).e.Status(200).Done()
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
	}
	return dto
}

func GetMetadataKey(t *testing.T, client *httpClient, keyId int, wantHttpStatus int) *models.MetadataKey {
	req := client.Get("/api/v1/metadata/keys/" + strconv.Itoa(keyId))
	dto := &models.MetadataKey{}
	e := req.Expect(t)
	if wantHttpStatus == 200 {
		e.Json(t, dto).e.Status(200).Done()
	} else {
		e.e.Status(wantHttpStatus).Done()
	}
	return dto
}

func GetMetadataKeyValues(t *testing.T, client *httpClient, keyId int, sort *api.SortKey, wantHttpStatus int) *[]models.MetadataValue {
	url := fmt.Sprintf("/api/v1/metadata/keys/%d/values", keyId)

	req := client.Get(url)
	if sort != nil {
		order := ""
		if sort.Order {
			order = "ASC"
		} else {
			order = "DESC"
		}
		req = req.Sort(sort.Key, order)
	}

	dto := &[]models.MetadataValue{}
	e := req.Expect(t)
	if wantHttpStatus == 200 {
		e.Json(t, dto).e.Status(200).Done()
	} else {
		e.e.Status(wantHttpStatus).Done()
	}
	return dto
}

func UpdateMetadataKey(t *testing.T, client *httpClient, wantHttpStatus int, key *models.MetadataKey) {
	req := client.Put("/api/v1/metadata/keys/" + strconv.Itoa(key.Id))
	req.Json(t, key)
	req.req.Expect(t).Status(wantHttpStatus).Done()
}

func DeleteMetadataKey(t *testing.T, client *httpClient, wantHttpStatus int, keyId int) {
	req := client.Delete("/api/v1/metadata/keys/" + strconv.Itoa(keyId))
	req.Expect(t).e.Status(wantHttpStatus).Done()
}

func AddMetadataValue(t *testing.T, client *httpClient, keyId int, value *models.MetadataValue, wantHttpStatus int) *models.MetadataValue {
	dto := &api.MetadataValueRequest{
		Value:          value.Value,
		Comment:        value.Comment,
		MatchDocuments: value.MatchDocuments,
		MatchType:      string(value.MatchType),
		MatchFilter:    value.MatchFilter,
	}

	req := client.Post("/api/v1/metadata/keys/"+strconv.Itoa(keyId)+"/values").Json(t, dto)
	body := &models.MetadataValue{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, body).e.Status(200).Done()
		assert.True(t, body.Id > 0, "id must be > 0")
		body.KeyId = keyId
		return body
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
	}
	return nil
}
