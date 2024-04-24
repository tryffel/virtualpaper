package models

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"strconv"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/errors"
)

type Property struct {
	Id        int          `json:"id" db:"id"`
	User      int          `json:"user_id" db:"user_id"`
	Name      string       `json:"name" db:"name"`
	Type      PropertyType `json:"type" db:"type"`
	Global    bool         `json:"global" db:"global"`
	Unique    bool         `json:"unique" db:"is_unique"`
	Exclusive bool         `json:"exclusive" db:"is_exclusive"`
	Counter   int          `json:"counter" db:"counter"`
	// probably won't be needed
	Offset   int    `json:"offset" db:"counter_offset"`
	Prefix   string `json:"prefix" db:"prefix"`
	Mode     string `json:"mode" db:"mode"`
	Readonly bool   `json:"read_only" db:"read_only"`
	DateFmt  string `json:"date_fmt" db:"date_fmt"`
	Timestamp
}

var GeneratedPropertyTypes = []PropertyType{IdProperty, TextProperty, CounterProperty, DateProperty}

var AllPropertyTypes = []PropertyType{LinkProperty, IdProperty, JsonProperty,
	TextProperty, CounterProperty,
	IntProperty, FloatProperty, BooleanProperty,
	DateProperty, UserProperty}

func (p *Property) IsGenerated() bool {
	for _, v := range GeneratedPropertyTypes {
		if p.Type == v {
			return true
		}
	}
	return false
}

func (p *Property) Validate() error {
	// validate configuration
	validateError := errors.ErrInvalid
	if p.Type == IdProperty {
		if p.Mode != "uuid" {
			validateError.ErrMsg = fmt.Sprintf("invalid mode: '%s'", p.Mode)
			return validateError
		}
	}

	return nil
}

func (p *Property) ValidateValue(val string) error {
	validateError := errors.ErrInvalid
	if p.Type == JsonProperty {
		ok := json.Valid([]byte(val))
		if !ok {
			validateError.ErrMsg = "invalid json"
			return validateError
		}
	}
	if p.Type == IntProperty {
		_, err := strconv.Atoi(val)
		if err != nil {
			validateError.ErrMsg = "invalid number"
			return validateError
		}
	}
	if p.Type == FloatProperty {
		_, err := strconv.ParseFloat(val, 64)
		if err != nil {
			validateError.ErrMsg = "invalid number"
			return validateError
		}
	}
	if p.Type == BooleanProperty {
		lower := strings.ToLower(val)
		if lower != "true" && lower != "false" {
			validateError.ErrMsg = "must be boolean (true|false)"
			return validateError
		}
	}
	return nil
}

func (p *Property) Generate() (string, error) {

	switch p.Type {
	case IdProperty:
		return p.generateId()
	case TextProperty:
		return p.Prefix, nil
	case CounterProperty:
		p.Counter += 1
		return p.Prefix + strconv.Itoa(p.Counter), nil
	case DateProperty:
		now := time.Now()
		return now.Format(p.DateFmt), nil
	}
	err := errors.ErrInternalError
	err.ErrMsg = fmt.Sprintf("unknown proprerty type '%s'", p.Type)
	return "", err
}

func (p *Property) generateId() (string, error) {
	switch p.Mode {
	case "uuid":
		value, err := uuid.GenerateUUID()
		return value, err
	}
	return "", fmt.Errorf("cannot generate id for unknown type '%s'", p.Type)
}

func (p *Property) FilterAttributes() []string {
	return []string{"id", "name", "created_at", "type", "global", "unique", "counter", "prefix", "mode", "readonly", "date_fmt"}
}

func (p *Property) SortNoCase() []string {
	return []string{"name", "type", "prefix", "mode", "date_fmt"}
}

func (p *Property) SortAttributes() []string {
	return p.FilterAttributes()
}

type PropertyType string

const (
	LinkProperty    PropertyType = "url"
	IdProperty      PropertyType = "id"
	JsonProperty    PropertyType = "json"
	TextProperty    PropertyType = "text"
	CounterProperty PropertyType = "counter"
	IntProperty     PropertyType = "int"
	FloatProperty   PropertyType = "float"
	BooleanProperty PropertyType = "boolean"
	DateProperty    PropertyType = "date"
	UserProperty    PropertyType = "user"
)

type DocumentProperty struct {
	Id           int    `json:"id" db:"id"`
	Document     string `json:"document_id" db:"document_id"`
	Property     int    `json:"property_id" db:"property_id"`
	PropertyName string `json:"property_name" db:"property_name"`
	Value        string `json:"value"`
	Description  string `json:"description" db:"description"`
	Timestamp
}
