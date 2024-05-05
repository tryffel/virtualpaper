package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"regexp"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
)

type DocumentPropertySuite struct {
	ApiTestSuite
	properties map[string]*aggregates.Property
}

func (suite *DocumentPropertySuite) SetupTest() {
	suite.Init()
	clearDbMetadataTables(suite.T(), suite.db)
	clearDbDocumentTables(suite.T(), suite.db)
	insertTestDocuments(suite.T(), suite.db)

	suite.properties = make(map[string]*aggregates.Property)
	suite.properties["archive-id"] = AddProperty(suite.T(), suite.userHttp, api.PropertyRequest{
		Name:      "archive-id",
		Type:      "text",
		Global:    false,
		Unique:    true,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  false,
		DateFmt:   "",
	}, 200, "")
	suite.properties["source-url"] = AddProperty(suite.T(), suite.userHttp, api.PropertyRequest{
		Name:      "source-url",
		Type:      "url",
		Global:    false,
		Unique:    false,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  false,
		DateFmt:   "",
	}, 200, "")
	suite.properties["free-form"] = AddProperty(suite.T(), suite.userHttp, api.PropertyRequest{
		Name:      "free-form",
		Type:      "text",
		Global:    false,
		Unique:    false,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  false,
		DateFmt:   "",
	}, 200, "")
}

func TestDocumentProperties(t *testing.T) {
	suite.Run(t, new(DocumentPropertySuite))
}

func (suite *DocumentPropertySuite) TestCrud() {
	intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), intel.Properties, 0)
	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     suite.properties["archive-id"].Id,
			PropertyName: "",
			Value:        "id",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 200)
	updated := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), updated.Properties, 1)
	assert.Equal(suite.T(), intel.Properties[0].Property, updated.Properties[0].Property)
	assert.Equal(suite.T(), intel.Properties[0].Value, updated.Properties[0].Value)
	assert.Equal(suite.T(), intel.Properties[0].Description, updated.Properties[0].Description)

	updated.Properties = append(updated.Properties,
		aggregates.DocumentProperty{
			Id:           0,
			Property:     suite.properties["free-form"].Id,
			PropertyName: "",
			Value:        "some value",
			Description:  "",
			Timestamp:    models.Timestamp{},
		},
	)
	updateDocument(suite.T(), suite.userHttp, updated, 200)
	updated = getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	originalPropertyId := updated.Properties[1].Id

	assert.Len(suite.T(), updated.Properties, 2)
	assert.Equal(suite.T(), originalPropertyId, updated.Properties[1].Id, "property id hasn't changed")

	updated.Properties = updated.Properties[1:]
	updateDocument(suite.T(), suite.userHttp, updated, 200)
	updated = getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), updated.Properties, 1)
	assert.Equal(suite.T(), originalPropertyId, updated.Properties[0].Id, "property has been deleted")
}

func (suite *DocumentPropertySuite) TestReadOnly() {
	prop := suite.properties["archive-id"]
	prop.Readonly = true
	UpdateProperty(suite.T(), suite.userHttp, prop.Id, api.PropertyRequest{
		Name:      prop.Name,
		Type:      prop.Type,
		Global:    false,
		Unique:    false,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}, 200)

	intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), intel.Properties, 0)
	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "original",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 200)
	updated := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)

	updated.Properties[0].Value = "changed"
	updateDocument(suite.T(), suite.userHttp, updated, 400)
}

func (suite *DocumentPropertySuite) TestValidate() {
	//TODO: implement
}

func (suite *DocumentPropertySuite) TestDate() {
	prop := AddProperty(suite.T(), suite.userHttp, api.PropertyRequest{
		Name:      "date",
		Type:      models.DateProperty,
		Global:    false,
		Unique:    false,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   time.DateOnly,
	}, 200, "")

	intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), intel.Properties, 0)
	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 200)
	updated := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)

	assert.Len(suite.T(), updated.Properties, 1)
	assert.Equal(suite.T(), time.Now().Format(time.DateOnly), updated.Properties[0].Value)

	UpdateProperty(suite.T(), suite.userHttp, prop.Id, api.PropertyRequest{
		Name:      prop.Name,
		Type:      models.DateProperty,
		Global:    false,
		Unique:    false,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   time.DateTime,
	}, 200)

	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 200)
	updated = getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)

	assert.Len(suite.T(), updated.Properties, 1)
	assert.Equal(suite.T(), time.Now().Format(time.DateTime), updated.Properties[0].Value)
}

func (suite *DocumentPropertySuite) TestCounter() {
	prop := suite.properties["archive-id"]
	UpdateProperty(suite.T(), suite.userHttp, prop.Id, api.PropertyRequest{
		Name:      prop.Name,
		Type:      models.CounterProperty,
		Global:    false,
		Unique:    false,
		Exclusive: false,
		Counter:   5000,
		Prefix:    "aid-",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}, 200)
	intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	metaMorphosis := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	props := []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	intel.Properties = props
	metaMorphosis.Properties = props
	updateDocument(suite.T(), suite.userHttp, intel, 200)
	updateDocument(suite.T(), suite.userHttp, metaMorphosis, 200)

	intel = getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	metaMorphosis = getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)

	assert.Len(suite.T(), intel.Properties, 1)
	assert.Len(suite.T(), metaMorphosis.Properties, 1)
	assert.Equal(suite.T(), "aid-5001", intel.Properties[0].Value)
	assert.Equal(suite.T(), "aid-5002", metaMorphosis.Properties[0].Value)
	assert.Equal(suite.T(), "autogenerated", metaMorphosis.Properties[0].Description)
	assert.Equal(suite.T(), "autogenerated", intel.Properties[0].Description)

	prop = GetProperty(suite.T(), suite.userHttp, prop.Id, 200)
	assert.Equal(suite.T(), prop.Counter, 5002)
}

func (suite *DocumentPropertySuite) TestUuid() {
	regex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	prop := AddProperty(suite.T(), suite.userHttp, api.PropertyRequest{
		Name:      "uuid",
		Type:      models.IdProperty,
		Global:    false,
		Unique:    false,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "uuid",
		Readonly:  false,
		DateFmt:   "",
	}, 200, "")

	intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), intel.Properties, 0)
	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 200)
	updated := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)

	assert.Len(suite.T(), updated.Properties, 1)
	assert.Regexp(suite.T(), regex, updated.Properties[0].Value)

	metamorphosis := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	metamorphosis.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, metamorphosis, 200)
	updatedMetadatamorphosis := getDocument(suite.T(), suite.userHttp, metamorphosis.Id, 200)

	assert.NotEqual(suite.T(), updatedMetadatamorphosis.Properties[0].Value, updated.Properties[0].Value)
	assert.Regexp(suite.T(), regex, updatedMetadatamorphosis.Properties[0].Value)
}

func (suite *DocumentPropertySuite) TestUnique() {
	prop := suite.properties["archive-id"]
	prop.Readonly = true
	UpdateProperty(suite.T(), suite.userHttp, prop.Id, api.PropertyRequest{
		Name:      prop.Name,
		Type:      prop.Type,
		Global:    false,
		Unique:    true,
		Exclusive: false,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}, 200)

	intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	metaMorphosis := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "original",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 200)
	metaMorphosis.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "original",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, metaMorphosis, 400)
	metaMorphosis.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "new value",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, metaMorphosis, 200)
}

func (suite *DocumentPropertySuite) TestExclusive() {
	prop := suite.properties["archive-id"]
	prop.Readonly = true
	UpdateProperty(suite.T(), suite.userHttp, prop.Id, api.PropertyRequest{
		Name:      prop.Name,
		Type:      prop.Type,
		Global:    false,
		Unique:    false,
		Exclusive: true,
		Counter:   0,
		Prefix:    "",
		Mode:      "",
		Readonly:  true,
		DateFmt:   "",
	}, 200)

	intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "value",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "other value",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 400)
	intel.Properties = []aggregates.DocumentProperty{
		aggregates.DocumentProperty{
			Id:           0,
			Property:     prop.Id,
			PropertyName: "",
			Value:        "value",
			Description:  "description",
			Timestamp:    models.Timestamp{},
		},
	}
	updateDocument(suite.T(), suite.userHttp, intel, 200)
}
