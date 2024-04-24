package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
)

type PropertySuite struct {
	ApiTestSuite
}

func (suite *PropertySuite) SetupTest() {
	suite.Init()
	clearDbMetadataTables(suite.T(), suite.db)
}

func TestProperties(t *testing.T) {
	suite.Run(t, new(PropertySuite))
}

func (suite *PropertySuite) TestValidations() {
}

func (suite *PropertySuite) TestGet() {
	property := api.PropertyRequest{
		Name:      "test",
		Type:      "text",
		Global:    false,
		Unique:    true,
		Exclusive: true,
		Counter:   0,
		Offset:    0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}

	created := AddProperty(suite.T(), suite.userHttp, property, 200, "")
	assert.NotZero(suite.T(), created.Id, "id not zero")

	final := GetProperty(suite.T(), suite.userHttp, created.Id, 200)

	// these won't match exactly, better discard them
	final.CreatedAt = created.CreatedAt
	final.UpdatedAt = created.UpdatedAt

	assert.Equal(suite.T(), created, final)
	assert.Equal(suite.T(), property.Name, final.Name)
	assert.Equal(suite.T(), property.Type, final.Type)
	assert.Equal(suite.T(), property.Global, final.Global)
	assert.Equal(suite.T(), property.Unique, final.Unique)
	assert.Equal(suite.T(), property.Exclusive, final.Exclusive)
	assert.Equal(suite.T(), property.Counter, final.Counter)
	assert.Equal(suite.T(), property.Offset, final.Offset)
	assert.Equal(suite.T(), property.Prefix, final.Prefix)
	assert.Equal(suite.T(), property.Mode, final.Mode)
	assert.Equal(suite.T(), property.Readonly, final.Readonly)
	assert.Equal(suite.T(), property.DateFmt, final.DateFmt)
}

func (suite *PropertySuite) TestUpdate() {
	property := api.PropertyRequest{
		Name:      "test",
		Type:      "text",
		Global:    false,
		Unique:    true,
		Exclusive: true,
		Counter:   0,
		Offset:    0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}

	created := AddProperty(suite.T(), suite.userHttp, property, 200, "")
	assert.NotZero(suite.T(), created.Id, "id not zero")

	property.Name = "changed"
	property.Type = "counter"
	property.Readonly = false
	property.Unique = false
	property.Counter = 1
	property.Offset = 2
	property.Prefix = "prefix-"
	property.Mode = "uuid"
	property.DateFmt = "2006.01.02"

	UpdateProperty(suite.T(), suite.userHttp, created.Id, property, 200)

	final := GetProperty(suite.T(), suite.userHttp, created.Id, 200)
	assert.Equal(suite.T(), property.Name, final.Name)
	assert.Equal(suite.T(), property.Type, final.Type)
	assert.Equal(suite.T(), property.Global, final.Global)
	assert.Equal(suite.T(), property.Unique, final.Unique)
	assert.Equal(suite.T(), property.Exclusive, final.Exclusive)
	assert.Equal(suite.T(), property.Counter, final.Counter)
	assert.Equal(suite.T(), property.Offset, final.Offset)
	assert.Equal(suite.T(), property.Prefix, final.Prefix)
	assert.Equal(suite.T(), property.Mode, final.Mode)
	assert.Equal(suite.T(), property.Readonly, final.Readonly)
	assert.Equal(suite.T(), property.DateFmt, final.DateFmt)
}

func (suite *PropertySuite) TestList() {
	property := api.PropertyRequest{
		Name:      "test",
		Type:      "text",
		Global:    false,
		Unique:    true,
		Exclusive: true,
		Counter:   0,
		Offset:    0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}

	AddProperty(suite.T(), suite.userHttp, property, 200, "")
	property.Name = "second"
	AddProperty(suite.T(), suite.userHttp, property, 200, "")

	property.Name = "archive id"
	AddProperty(suite.T(), suite.userHttp, property, 200, "")

	property.Name = "link"
	property.Type = "url"
	property.Exclusive = false
	property.Readonly = false
	AddProperty(suite.T(), suite.userHttp, property, 200, "")

	properties := GetProperties(suite.T(), suite.userHttp, 200, nil)

	assert.Len(suite.T(), *properties, 4)
	assert.Equal(suite.T(), (*properties)[0].Name, "archive id")
	assert.Equal(suite.T(), (*properties)[1].Name, "link")
	assert.Equal(suite.T(), (*properties)[2].Name, "second")
	assert.Equal(suite.T(), (*properties)[3].Name, "test")

	assert.Equal(suite.T(), (*properties)[1].Type, models.PropertyType("url"))
	assert.Equal(suite.T(), (*properties)[1].Exclusive, false)
	assert.Equal(suite.T(), (*properties)[1].Readonly, false)
	assert.Equal(suite.T(), (*properties)[1].Unique, true)
}

func (suite *PropertySuite) TestAccess() {
	property := api.PropertyRequest{
		Name:      "test",
		Type:      "text",
		Global:    false,
		Unique:    true,
		Exclusive: true,
		Counter:   0,
		Offset:    0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}

	userProperty := AddProperty(suite.T(), suite.userHttp, property, 200, "")
	property.Name = "second"
	adminProperty := AddProperty(suite.T(), suite.adminHttp, property, 200, "")

	GetProperty(suite.T(), suite.userHttp, adminProperty.Id, 404)
	GetProperty(suite.T(), suite.adminHttp, adminProperty.Id, 200)

	userProps := GetProperties(suite.T(), suite.userHttp, 200, nil)
	assert.Len(suite.T(), *userProps, 1)
	assert.Equal(suite.T(), (*userProps)[0].Id, userProperty.Id)

	adminProps := GetProperties(suite.T(), suite.adminHttp, 200, nil)
	assert.Len(suite.T(), *adminProps, 1)
	assert.Equal(suite.T(), (*adminProps)[0].Id, adminProperty.Id)
}

func (suite *PropertySuite) TestGlobalRequiresAdmin() {
	property := api.PropertyRequest{
		Name:      "test",
		Type:      "text",
		Global:    true,
		Unique:    true,
		Exclusive: true,
		Counter:   0,
		Offset:    0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}
	AddProperty(suite.T(), suite.userHttp, property, 400, "user not admin")
	AddProperty(suite.T(), suite.adminHttp, property, 200, "")
}

func (suite *PropertySuite) TestGlobalDeniesDuplicate() {
	property := api.PropertyRequest{
		Name:      "test",
		Type:      "text",
		Global:    true,
		Unique:    true,
		Exclusive: true,
		Counter:   0,
		Offset:    0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}
	AddProperty(suite.T(), suite.adminHttp, property, 200, "")
	AddProperty(suite.T(), suite.adminHttp, property, 304, "")

	property.Global = false
	AddProperty(suite.T(), suite.userHttp, property, 200, "")

	property.Name = "changed"
	AddProperty(suite.T(), suite.userHttp, property, 200, "")
	AddProperty(suite.T(), suite.adminHttp, property, 200, "")
}

func (suite *PropertySuite) TestGenerated() {
	property := api.PropertyRequest{
		Name:      "test",
		Type:      "text",
		Global:    true,
		Unique:    true,
		Exclusive: true,
		Counter:   0,
		Offset:    0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}
	AddProperty(suite.T(), suite.adminHttp, property, 200, "")
	AddProperty(suite.T(), suite.adminHttp, property, 304, "")

	property.Global = false
	AddProperty(suite.T(), suite.userHttp, property, 200, "")

	property.Name = "changed"
	AddProperty(suite.T(), suite.userHttp, property, 200, "")
	AddProperty(suite.T(), suite.adminHttp, property, 200, "")
}

func AddProperty(t *testing.T, client *httpClient, property api.PropertyRequest, wantHttpStatus int, errorMessage string) *aggregates.Property {
	req := client.Post("/api/v1/properties").Json(t, property)
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
		assert.True(t, isToday(body.CreatedAt), "timestamp today")
		assert.True(t, isToday(body.UpdatedAt), "timestamp today")
	} else {
		req.Expect(t).AssertError(t, errorMessage).e.Status(wantHttpStatus).Done()
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
	//req := client.Get("/api/v1/properties?page=1&page_size=100&sort=name&sort_order=ASC")
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
