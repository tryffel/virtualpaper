package services

import (
	"context"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

type MetadataService struct {
	db *storage.Database
}

func NewMetadataService(db *storage.Database) *MetadataService {
	return &MetadataService{
		db: db,
	}
}

func (service *MetadataService) GetKeys(ctx context.Context, userId int, ids []int, sort storage.SortKey, pagination storage.Paging) (*[]models.MetadataKeyAnnotated, int, error) {
	return service.db.MetadataStore.GetKeys(userId, ids, sort, pagination)
}

func (service *MetadataService) UserOwnsKey(ctx context.Context, userId, keyId int) (bool, error) {
	return service.db.MetadataStore.UserHasKey(userId, keyId)
}

func (service *MetadataService) GetKeyValue(ctx context.Context, keyId int, sort storage.SortKey, paging storage.Paging) (*[]models.MetadataValue, error) {
	return service.db.MetadataStore.GetValues(keyId, sort, paging)
}
