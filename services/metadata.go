package services

import (
	"context"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services/process"
	"tryffel.net/go/virtualpaper/storage"
)

type MetadataService struct {
	db      *storage.Database
	process *process.Manager
}

func NewMetadataService(db *storage.Database, manager *process.Manager) *MetadataService {
	return &MetadataService{
		db:      db,
		process: manager,
	}
}

func (service *MetadataService) GetKeys(ctx context.Context, userId int, ids []int, sort storage.SortKey, pagination storage.Paging) (*[]models.MetadataKeyAnnotated, int, error) {
	return service.db.MetadataStore.GetKeys(userId, ids, sort, pagination)
}

func (service *MetadataService) UserOwnsKey(ctx context.Context, userId, keyId int) (bool, error) {
	return service.db.MetadataStore.UserHasKey(userId, keyId)
}

func (service *MetadataService) GetKeyValues(ctx context.Context, keyId int, sort storage.SortKey, paging storage.Paging) (*[]models.MetadataValue, error) {
	return service.db.MetadataStore.GetValues(keyId, sort, paging)
}

func (service *MetadataService) GetKey(ctx context.Context, keyId int) (*models.MetadataKey, error) {
	return service.db.MetadataStore.GetKey(keyId)
}

func (service *MetadataService) Create(ctx context.Context, userId int, key *models.MetadataKey) error {
	return service.db.MetadataStore.CreateKey(userId, key)
}

func (service *MetadataService) CreateValue(ctx context.Context, value *models.MetadataValue) error {
	return service.db.MetadataStore.CreateValue(value)
}

func (service *MetadataService) UpdateValue(ctx context.Context, value *models.MetadataValue) error {
	// TODO: wrap in transaction
	err := service.db.MetadataStore.UpdateValue(value)
	if err != nil {
		return err
	}
	return service.db.JobStore.IndexDocumentsByMetadata(value.UserId, value.KeyId, value.Id)
}

func (service *MetadataService) UpdateKey(ctx context.Context, key *models.MetadataKey) error {
	// TODO: wrap in transaction
	err := service.db.MetadataStore.UpdateKey(key)
	if err != nil {
		return err
	}
	return service.db.JobStore.IndexDocumentsByMetadata(key.UserId, key.Id, 0)
}

func (service *MetadataService) DeleteKey(ctx context.Context, userId int, keyId int) error {
	// need to add processing when the metadata still exists
	// TODO: wrap in transaction
	err := service.db.JobStore.IndexDocumentsByMetadata(userId, keyId, 0)
	if err != nil {
		return err
	}

	err = service.db.MetadataStore.DeleteKey(userId, keyId)
	if err != nil {
		return err
	}
	service.process.PullDocumentsToProcess()
	return nil
}

func (service *MetadataService) DeleteValue(ctx context.Context, userId, keyId, valueId int) error {
	// need to add processing when the metadata still exists
	err := service.db.JobStore.IndexDocumentsByMetadata(userId, keyId, valueId)
	if err != nil {
		return err
	}

	err = service.db.MetadataStore.DeleteValue(userId, valueId)
	if err != nil {
		return err
	}
	service.process.PullDocumentsToProcess()
	return nil
}

func (service *MetadataService) SearchMetadata(ctx context.Context, userId int, query string) (*models.MetadataSearchResult, error) {
	tx, err := storage.NewTx(service.db, ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit()
	return service.db.MetadataStore.Search(tx, userId, query)
}
