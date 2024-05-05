package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

type PropertyStore struct {
	*resource
	db *sqlx.DB
	sq squirrel.StatementBuilderType
}

func NewPropertyStore(db *sqlx.DB) *PropertyStore {
	return &PropertyStore{
		resource: &resource{
			name: "property",
			db:   db,
		},
		db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (store *PropertyStore) GetSq() squirrel.StatementBuilderType {
	return store.sq
}

func (store *PropertyStore) UserOwnsProperty(execer SqlExecer, userId int, propertyId int) (bool, error) {
	property, err := store.GetProperty(execer, propertyId)
	if err != nil {
		return false, err
	}
	return property.User == userId, nil
}

func (store *PropertyStore) GetProperty(execer SqlExecer, id int) (*models.Property, error) {
	query := store.sq.Select("*").From("properties").Where("id = ?", id)
	property := &models.Property{}
	err := execer.GetSq(property, query)
	if err != nil {
		return nil, store.parseError(err, "get")
	}
	return property, nil
}

func (store *PropertyStore) GetProperties(execer SqlExecer, userId int, paging Paging, sort SortKey) (*[]models.Property, error) {
	sort.Validate("name")

	query := store.sq.Select("*").
		From("properties").
		Where("user_id = ?", userId).
		Limit(uint64(paging.Limit)).Offset(uint64(paging.Offset)).
		OrderBy(sort.QueryKey() + " " + sort.SortOrder())

	data := &[]models.Property{}

	err := execer.SelectSq(data, query)
	if err != nil {
		return nil, store.parseError(err, "get list")
	}
	return data, nil
}

func (store *PropertyStore) GetTotalProperties(execer SqlExecer, userId int) (int, error) {
	query := store.sq.Select("count(id) as total").
		From("properties").
		Where("user_id = ?", userId)

	rows, err := execer.QuerySq(query)
	if err != nil {
		return 0, store.parseError(err, "get total")
	}

	total := 0
	for rows.Next() {
		err = rows.Scan(&total)
		if err != nil {
			return 0, store.parseError(err, "scan total")
		}
	}
	return total, nil
}

func (store *PropertyStore) AddProperty(execer SqlExecer, property *models.Property) error {
	property.CreatedAt = time.Now()
	property.Update()
	query := store.sq.Insert("properties").
		Columns("user_id", "name", "type", "global", "is_unique", "is_exclusive",
			"counter", "prefix", "mode", "read_only", "date_fmt").
		Values(property.User, property.Name, property.Type, property.Global, property.Unique, property.Exclusive,
			property.Counter, property.Prefix, property.Mode, property.Readonly, property.DateFmt).
		Suffix("RETURNING id")

	var id int
	rows, err := execer.QuerySq(query)
	if err != nil {
		return store.parseError(err, "insert")
	}

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return fmt.Errorf("scan id: %d", err)
		}
	}
	property.Id = id
	property.Update()
	return nil
}

func (store *PropertyStore) UpdatePropertyCounter(execer SqlExecer, property *models.Property) error {
	property.Update()
	query := store.sq.Update("properties").SetMap(map[string]interface{}{
		"user_id":      property.User,
		"name":         property.Name,
		"type":         property.Type,
		"global":       property.Global,
		"is_unique":    property.Unique,
		"is_exclusive": property.Exclusive,
		"counter":      property.Counter,
		"prefix":       property.Prefix,
		"mode":         property.Mode,
		"read_only":    property.Readonly,
		"date_fmt":     property.DateFmt,
		"updated_at":   property.UpdatedAt,
	}).Where("id=?", property.Id)

	_, err := execer.ExecSq(query)
	if err != nil {
		return store.parseError(err, "update")
	}
	return nil
}

func (store *PropertyStore) UpdateProperty(execer SqlExecer, property *models.Property) error {
	err := store.UpdatePropertyCounter(execer, property)
	if err != nil {
		return err
	}

	query := store.sq.Update("document_properties").SetMap(map[string]interface{}{
		"is_unique":    property.Unique,
		"is_exclusive": property.Exclusive,
		"global":       property.Exclusive,
	}).Where("property_id = ?", property.Id)
	_, err = execer.ExecSq(query)
	if err != nil {
		return store.parseError(err, "update document_properties")
	}
	return nil
}

func (store *PropertyStore) AddDocumentProperty(execer SqlExecer, property *models.Property, documentId, value, description string, updateProperty bool) error {
	query := store.sq.Insert("document_properties").
		Columns("document_id", "property_id", "user_id", "value", "description", "is_unique", "is_exclusive", "global").
		Values(documentId, property.Id, property.User, value, description, property.Unique, property.Exclusive, property.Global)

	_, err := execer.ExecSq(query)
	if err != nil {
		return store.parseError(err, "insert document property")
	}

	if updateProperty {
		return store.UpdatePropertyCounter(execer, property)
	}
	return nil
}

func (store *PropertyStore) GetDocumentProperties(execer SqlExecer, documentId string) (*[]models.DocumentProperty, error) {
	data := &[]models.DocumentProperty{}
	query := store.sq.Select("dp.id as id", "dp.document_id as document_id", "dp.property_id as property_id",
		"dp.value as value", "dp.description as description",
		"dp.created_at as created_at", "dp.updated_at as updated_at", "p.name as property_name").
		From("document_properties dp").
		LeftJoin("properties p ON dp.property_id = p.id").
		Where("document_id = ?", documentId).
		OrderBy("p.name ASC")

	err := execer.SelectSq(data, query)
	if err != nil {
		return nil, store.parseError(err, "get document properties")
	}
	return data, err
}

func (store *PropertyStore) DeleteProperty(execer SqlExecer, userId int, properties []int) error {
	query := store.sq.Delete("properties").Where("id IN ?", properties)
	_, err := execer.ExecSq(query)
	if err != nil {
		return store.parseError(err, "delete")
	}
	return nil
}

func (store *PropertyStore) UpdateDocumentProperties(execer SqlExecer, properties *[]models.DocumentProperty) error {
	query := store.sq.Update("document_properties as dp").From("(values ").SetMap(map[string]interface{}{
		"document_id": "news_vals.document_id",
		"property_id": "new_vals.property_id",
		"value":       "new_vals.value",
		"description": "new_vals.description",
		"updated_at":  "new_vals.updated_at",
	})

	for _, v := range *properties {
		v.Update()
		query = query.Suffix("(?, ?, ?, ?, ?, ?)", v.Id, v.Document, v.Property, v.Value, v.Description, v.UpdatedAt)
	}
	query = query.Suffix(") as new_vals(id, document_id, property_id, value, description, updated_at) WHERE new_vals.id = dp.id")

	_, err := execer.ExecSq(query)
	if err != nil {
		return store.parseError(err, "update")
	}
	return nil
}

func (store *PropertyStore) UpdateDocumentProperty(execer SqlExecer, property *models.DocumentProperty) error {
	property.Update()
	query := store.sq.Update("document_properties").SetMap(map[string]interface{}{
		"document_id": property.Document,
		"property_id": property.Property,
		"value":       property.Value,
		"description": property.Description,
		"updated_at":  property.UpdatedAt,
	}).Where("id = ?", property.Id)
	_, err := execer.ExecSq(query)
	if err != nil {
		return store.parseError(err, "update")
	}
	return nil
}

func (store *PropertyStore) DeleteDocumentProperties(execer SqlExecer, userId int, docId string, properties []int) error {
	query := store.sq.Delete("document_properties").
		Where(squirrel.Eq{"id": properties}).
		Where("document_id = ?", docId)
	_, err := execer.ExecSq(query)
	if err != nil {
		return store.parseError(err, "delete")
	}
	return nil
}
