package services

import (
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
)

type DocumentService struct {
	db     *storage.DocumentStore
	search *search.Engine
}

func NewDocumentService(db *storage.DocumentStore, search *search.Engine) *DocumentService {
	return &DocumentService{
		db:     db,
		search: search,
	}
}

func (service *DocumentService) SearchDocuments(userId int, query string, sort storage.SortKey, paging storage.Paging) ([]*models.Document, int, error) {
	return service.search.SearchDocuments(userId, query, sort, paging)
}

func (service *DocumentService) GetDocuments(userId int, paging storage.Paging, sort storage.SortKey, limitContent bool) (*[]models.Document, int, error) {
	return service.db.GetDocuments(userId, paging, sort, limitContent, false)
}

func (service *DocumentService) GetDeletedDocuments(userId int, paging storage.Paging, sort storage.SortKey, limitContent bool) (*[]models.Document, int, error) {
	return service.db.GetDocuments(userId, paging, sort, limitContent, true)
}

func (service *DocumentService) UserOwnsDocument(documentId string, userId int) (bool, error) {
	return service.db.UserOwnsDocument(documentId, userId)
}
