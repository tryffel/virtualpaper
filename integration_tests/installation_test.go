package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type AdminTestSuite struct {
	ApiTestSuite
}

func TestAdminSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}

func (suite *AdminTestSuite) TestGetServerInstallation() {
	suite.publicHttp.Get("/api/v1/admin/systeminfo").Expect(suite.T()).e.Status(401).Done()
	suite.userHttp.Get("/api/v1/admin/systeminfo").Expect(suite.T()).e.Status(200).Done()

	data := &api.SystemInfo{}
	suite.adminHttp.Get("/api/v1/admin/systeminfo").Expect(suite.T()).Json(suite.T(), data).e.Status(200).Done()

	assert.Equal(suite.T(), "Virtualpaper", data.Name)
	assert.Equal(suite.T(), true, data.PopplerInstalled)
	assert.Equal(suite.T(), true, data.PandocInstalled)
	assert.Equal(suite.T(), true, data.SearchEngineStatus.Ok)
}

func (suite *AdminTestSuite) TestGetUsers() {
	suite.userHttp.Get("/api/v1/admin/users").Expect(suite.T()).e.Status(401).Done()

	data := &[]models.UserInfo{}
	suite.adminHttp.Get("/api/v1/admin/users").Expect(suite.T()).Json(suite.T(), data).e.Status(200).Done()
	assert.Equal(suite.T(), len(*data), 2, "system has two users")

	assert.Equal(suite.T(), "admin", (*data)[0].UserName)
	assert.Equal(suite.T(), "", (*data)[0].Email)
	assert.Equal(suite.T(), true, (*data)[0].IsAdmin)
	assert.Equal(suite.T(), true, (*data)[0].IsActive)

	assert.Equal(suite.T(), "user", (*data)[1].UserName)
	assert.Equal(suite.T(), "", (*data)[1].Email)
	assert.Equal(suite.T(), false, (*data)[1].IsAdmin)
	assert.Equal(suite.T(), true, (*data)[1].IsActive)
}

func (suite *AdminTestSuite) TestGetFiletypes() {
	dto := &api.MimeTypesSupportedResponse{}
	suite.userHttp.Get("/api/v1/filetypes").Expect(suite.T()).Json(suite.T(), dto).e.Status(200).Done()
	assert.Equal(suite.T(), 12, len(dto.Names))

	assert.Contains(suite.T(), dto.Names, ".csv")
	assert.Contains(suite.T(), dto.Names, ".epub")
	assert.Contains(suite.T(), dto.Names, ".html")
	assert.Contains(suite.T(), dto.Names, ".jpg")
	assert.Contains(suite.T(), dto.Names, ".jpeg")
	assert.Contains(suite.T(), dto.Names, ".png")
	assert.Contains(suite.T(), dto.Names, ".md")
	assert.Contains(suite.T(), dto.Names, ".txt")

	assert.Contains(suite.T(), dto.Mimetypes, "image/png")
	assert.Contains(suite.T(), dto.Mimetypes, "image/jpeg")
	assert.Contains(suite.T(), dto.Mimetypes, "image/jpg")
	assert.Contains(suite.T(), dto.Mimetypes, "text/plain")
	assert.Contains(suite.T(), dto.Mimetypes, "application/pdf")
}
