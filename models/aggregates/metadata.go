package aggregates

import "tryffel.net/go/virtualpaper/models"

type Property struct {
	Id        int                 `json:"id"`
	Name      string              `json:"name"`
	Type      models.PropertyType `json:"property_type"`
	Global    bool                `json:"global"`
	Unique    bool                `json:"unique"`
	Exclusive bool                `json:"exclusive"`
	Counter   int                 `json:"counter"`
	Prefix    string              `json:"prefix" db:"prefix"`
	Mode      string              `json:"mode" db:"mode"`
	Readonly  bool                `json:"read_only" db:"read_only"`
	DateFmt   string              `json:"date_fmt" db:"date_fmt"`
	models.Timestamp
}

func MapProperty(p *models.Property) *Property {
	return &Property{
		Id:        p.Id,
		Name:      p.Name,
		Type:      p.Type,
		Global:    p.Global,
		Unique:    p.Unique,
		Exclusive: p.Exclusive,
		Counter:   p.Counter,
		Prefix:    p.Prefix,
		Mode:      p.Mode,
		Readonly:  p.Readonly,
		DateFmt:   p.DateFmt,
		Timestamp: models.Timestamp{
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
	}
}

type DocumentProperty struct {
	Id           int    `json:"id"`
	Property     int    `json:"property_id"`
	PropertyName string `json:"property_name"`
	Value        string `json:"value"`
	Description  string `json:"description"`
	models.Timestamp
}

func mapDocumentProperty(dp models.DocumentProperty) DocumentProperty {
	return DocumentProperty{
		Id:           dp.Id,
		Property:     dp.Property,
		PropertyName: dp.PropertyName,
		Value:        dp.Value,
		Description:  dp.Description,
		Timestamp:    dp.Timestamp,
	}
}

func mapDocumentPropertyArray(dp *[]models.DocumentProperty) *[]DocumentProperty {
	props := make([]DocumentProperty, len(*dp))

	for i, v := range *dp {
		props[i] = mapDocumentProperty(v)
	}
	return &props
}
