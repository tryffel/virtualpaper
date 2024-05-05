package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models/aggregates"
)

func AddProperty(t *testing.T, client *httpClient, property api.PropertyRequest, wantHttpStatus int, errorMessage string) *aggregates.Property {
	req := client.Post("/api/v1/properties").Json(t, property)
	body := &aggregates.Property{}
	if wantHttpStatus == 200 {
		req.Expect(t).AssertStatus(t, 200).Json(t, body).Done(t)
		assert.Greaterf(t, body.Id, 0, "id > 0")
		assert.Equal(t, body.Type, property.Type, "type matches")
		assert.Equal(t, body.Name, property.Name, "name matches")
		assert.Equal(t, body.Exclusive, property.Exclusive, "exclusive matches")
		assert.Equal(t, body.Unique, property.Unique, "unique matches")
		assert.Equal(t, body.Readonly, property.Readonly, "readonly matches")
		assert.Equal(t, body.Global, property.Global, "readonly matches")
		assert.True(t, isToday(body.CreatedAt), "timestamp today")
		assert.True(t, isToday(body.UpdatedAt), "timestamp today")
	} else {
		req.Expect(t).AssertError(t, errorMessage).AssertStatus(t, wantHttpStatus).Done(t)
	}
	return body
}

func UpdateProperty(t *testing.T, client *httpClient, id int, property api.PropertyRequest, wantHttpStatus int) *aggregates.Property {
	req := client.Put("/api/v1/properties/"+strconv.Itoa(id)).Json(t, property)
	body := &aggregates.Property{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, body).e.Status(200).Done()
		assert.Greaterf(t, body.Id, 0, "id > 0")
		assert.Equal(t, body.Type, property.Type, "type matches")
		assert.Equal(t, body.Name, property.Name, "name matches")
		assert.Equal(t, body.Exclusive, property.Exclusive, "exclusive matches")
		assert.Equal(t, body.Unique, property.Unique, "unique matches")
		assert.Equal(t, body.Readonly, property.Readonly, "readonly matches")
		assert.Equal(t, body.Global, property.Global, "readonly matches")
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
	}
	return body
}

func GetProperty(t *testing.T, client *httpClient, id int, wantHttpStatus int) *aggregates.Property {
	req := client.Get("/api/v1/properties/" + strconv.Itoa(id))
	dto := &aggregates.Property{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, dto).e.Status(200).Done()
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
	}
	return dto
}

func GetProperties(t *testing.T, client *httpClient, wantHttpStatus int, editFunc func(request *httpRequest) *httpRequest) *[]aggregates.Property {
	req := client.Get("/api/v1/properties")
	if editFunc != nil {
		req = editFunc(req)
	}
	dto := &[]aggregates.Property{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, dto).e.Status(200).Done()
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
	}
	return dto
}

func DeleteProperty(t *testing.T, client *httpClient, id int, wantHttpStatus int) {
	req := client.Delete("/api/v1/properties/" + strconv.Itoa(id))
	req.req.Expect(t).Status(wantHttpStatus).Done()
}
