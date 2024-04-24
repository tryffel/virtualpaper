package api

import (
	"github.com/labstack/echo/v4"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
)

type PropertyRequest struct {
	Name      string              `json:"name" valid:"stringlength(1|100)"`
	Type      models.PropertyType `json:"property_type" valid:"property_type~Invalid type"`
	Global    bool                `json:"global" valid:"-"`
	Unique    bool                `json:"unique" valid:"-"`
	Exclusive bool                `json:"exclusive" valid:"-"`
	Counter   int                 `json:"counter" valid:"-"`
	Offset    int                 `json:"offset" valid:"-"`
	Prefix    string              `json:"prefix" valid:"maxstringlength(200),optional"`
	Mode      string              `json:"mode" valid:"-"`
	Readonly  bool                `json:"read_only" valid:"-"`
	DateFmt   string              `json:"date_format" valid:"-"`
}

func (p *PropertyRequest) ToProperty() *models.Property {
	return &models.Property{
		Id:        0,
		User:      0,
		Name:      p.Name,
		Type:      p.Type,
		Global:    p.Global,
		Unique:    p.Unique,
		Exclusive: p.Exclusive,
		Counter:   p.Counter,
		Offset:    p.Offset,
		Prefix:    p.Prefix,
		Mode:      p.Mode,
		Readonly:  p.Readonly,
		DateFmt:   p.DateFmt,
		Timestamp: models.Timestamp{},
	}
}

type DocumentPropertyRequest struct {
	PropertyId int `json:"property_id" valid:"-"`
	//DocumentId  string `json:"documentId" valid:"uuid"`
	Value       string `json:"value" valid:"-"`
	Description string `json:"description" valid:"-"`
}

func (a *Api) GetProperties(c echo.Context) error {
	// swagger:route GET /api/v1/properties Properties GetProperties
	// Get properties
	//
	// responses:
	//   200: Property
	ctx := c.(UserContext)
	paging := getPagination(c)
	sort := getSort(c)
	if sort.Key == "" {
		sort.Key = "name"
	}
	opOk := false
	defer logCrudProperty(ctx.UserId, "get list", &opOk, "")

	properties, err := a.propertyService.GetProperties(c.Request().Context(), ctx.UserId, paging.toPagination(), sort.ToKey())
	if err != nil {
		return err
	}

	props := make([]aggregates.Property, len(*properties))
	for i, v := range *properties {
		props[i] = *aggregates.MapProperty(&v)

	}
	// TODO: need to get correct number of properties
	respResourceList(c.Response(), properties, len(*properties))
	opOk = true
	return nil
}

func (a *Api) GetProperty(c echo.Context) error {
	// swagger:route GET /api/v1/property/:id Properties GetProperty
	// Get property
	//
	// responses:
	//   200: Property
	ctx := c.(UserContext)
	opOk := false
	defer logCrudProperty(ctx.UserId, "get", &opOk, "")
	id, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	property, err := a.propertyService.GetProperty(c.Request().Context(), id)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(200, aggregates.MapProperty(property))
}

func (a *Api) AddProperty(c echo.Context) error {
	// swagger:route POST /api/v1/property Properties AddProperty
	// Add property
	//
	// responses:
	//   200: Property
	ctx := c.(UserContext)
	opOk := false
	defer logCrudProperty(ctx.UserId, "create", &opOk, "")
	data := &PropertyRequest{}
	err := unMarshalBody(c.Request(), data)
	if err != nil {
		return err
	}

	property := data.ToProperty()
	property.User = ctx.UserId
	err = a.propertyService.AddProperty(c.Request().Context(), ctx.User, property)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(200, aggregates.MapProperty(property))
}

func (a *Api) UpdateProperty(c echo.Context) error {
	// swagger:route PUT /api/v1/property/:id Properties UpdateProperty
	// Update property
	//
	// responses:
	//   200: Property
	ctx := c.(UserContext)
	opOk := false
	id, err := bindPathIdInt(c)
	if err != nil {
		return err
	}
	defer logCrudProperty(ctx.UserId, "update", &opOk, "")
	data := &PropertyRequest{}
	err = unMarshalBody(c.Request(), data)
	if err != nil {
		return err
	}
	property := data.ToProperty()
	property.User = ctx.UserId
	property.Id = id
	err = a.propertyService.UpdateProperty(c.Request().Context(), property)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(200, aggregates.MapProperty(property))
}
