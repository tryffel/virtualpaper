package services

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
	"tryffel.net/go/virtualpaper/services/process"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
	"tryffel.net/go/virtualpaper/util/logger"
)

type UploadedFile struct {
	UserId   int
	Filename string
	Mimetype string
	Size     int64
	File     io.ReadCloser
}

type DocumentService struct {
	db      *storage.Database
	search  *search.Engine
	process *process.Manager
}

func NewDocumentService(db *storage.Database, search *search.Engine, manager *process.Manager) *DocumentService {
	return &DocumentService{
		db:      db,
		search:  search,
		process: manager,
	}
}

func (service *DocumentService) SearchDocuments(userId int, query string, sort storage.SortKey, paging storage.Paging) ([]*models.Document, int, error) {
	return service.search.SearchDocuments(userId, query, sort, paging)
}

func (service *DocumentService) GetDocuments(userId int, paging storage.Paging, sort storage.SortKey, limitContent bool) (*[]models.Document, int, error) {
	return service.db.DocumentStore.GetDocuments(userId, paging, sort, limitContent, false)
}

func (service *DocumentService) GetDeletedDocuments(userId int, paging storage.Paging, sort storage.SortKey, limitContent bool) (*[]models.Document, int, error) {
	return service.db.DocumentStore.GetDocuments(userId, paging, sort, limitContent, true)
}

func (service *DocumentService) UserOwnsDocument(documentId string, userId int) (bool, error) {
	return service.db.DocumentStore.UserOwnsDocument(documentId, userId)
}

func (service *DocumentService) UploadFile(ctx context.Context, file *UploadedFile) (*models.Document, error) {
	tempHash, err := config.RandomString(10)
	if err != nil {
		logrus.Errorf("generate temporary hash for document: %v", err)
		return nil, errors.ErrInternalError
	}

	document := &models.Document{
		Id:       "",
		UserId:   file.UserId,
		Name:     file.Filename,
		Content:  "",
		Filename: file.Filename,
		Hash:     tempHash,
		Mimetype: file.Mimetype,
		Size:     file.Size,
		Date:     time.Now(),
	}

	if !process.MimeTypeIsSupported(file.Mimetype, file.Filename) {
		e := errors.ErrInvalid
		e.ErrMsg = fmt.Sprintf("unsupported file type: %v", file.Filename)
		return nil, e
	}

	tempFileName := storage.TempFilePath(tempHash)
	inputFile, err := os.OpenFile(tempFileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		logger.Context(ctx).Errorf("open new file for saving upload: %v", err)
		//respError(resp, fmt.Errorf("open new file for saving upload: %v", err), handler)
		return nil, err
	}
	n, err := inputFile.ReadFrom(file.File)
	if err != nil {
		return nil, fmt.Errorf("write uploaded file to disk: %v", err)
	}

	if n != file.Size {
		logger.Context(ctx).Warnf("did not fully read file: %d, got: %d", file.Size, n)
	}

	err = inputFile.Close()
	if err != nil {
		return nil, fmt.Errorf("close file: %v", err)
	}

	hash, err := process.GetHash(tempFileName)
	if err != nil {
		return nil, fmt.Errorf("get hash for temp file: %v", err)
	}

	existingDoc, err := service.db.DocumentStore.GetByHash(file.UserId, hash)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
		} else {
			return nil, fmt.Errorf("get existing document by hash: %v", err)
		}
	}

	if existingDoc != nil {
		if existingDoc.Id != "" {
			err := os.Remove(tempFileName)
			if err != nil {
				logger.Context(ctx).Errorf("remove duplicated temp file: %v", err)
			}
			return existingDoc, errors.ErrAlreadyExists
		}
	}

	document.Hash = hash
	err = service.db.DocumentStore.Create(document)
	if err != nil {
		return nil, err
	}

	newFile := storage.DocumentPath(document.Id)
	err = storage.CreateDocumentDir(document.Id)
	if err != nil {
		return nil, fmt.Errorf("create directory for doc: %v", err)
	}

	err = storage.MoveFile(tempFileName, newFile)
	if err != nil {
		return nil, fmt.Errorf("rename temp file by document id: %v", err)
	}

	err = service.db.JobStore.ProcessDocumentAllSteps(document.Id)
	if err != nil {
		return nil, fmt.Errorf("add process steps for new document: %v", err)
	}
	err = service.process.AddDocumentForProcessing(document.Id)
	return document, err
}

type DocumentFile struct {
	File     io.ReadCloser
	Size     int64
	Mimetype string
}

func (service *DocumentService) DocumentFile(docId string) (*DocumentFile, error) {
	doc, err := service.db.DocumentStore.GetDocument(docId)
	if err != nil {
		return nil, err
	}

	filePath := storage.DocumentPath(doc.Id)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	size := stat.Size()

	return &DocumentFile{
		File:     file,
		Size:     size,
		Mimetype: doc.Mimetype,
	}, nil
}

func (service *DocumentService) FlushDeletedDocument(ctx context.Context, docId string) error {
	document, err := service.db.DocumentStore.GetDocument(docId)
	if err != nil {
		return err
	}
	if !document.DeletedAt.Valid {
		return errors.ErrRecordNotFound
	}
	document.Update()

	err = process.DeleteDocument(docId)
	if err != nil {
		return fmt.Errorf("delete file: %v", err)
	}

	err = service.search.DeleteDocument(docId, document.UserId)
	if err != nil {
		return fmt.Errorf("delete document from search index: %v", err)
	}

	err = service.db.DocumentStore.DeleteDocument(docId)
	if err != nil {
		return err
	}
	return nil
}

func (service *DocumentService) DeleteDocument(ctx context.Context, docId string, userId int) error {
	doc, err := service.db.DocumentStore.GetDocument(docId)
	if err != nil {
		return err
	}
	if doc.DeletedAt.Valid {
		return errors.ErrInvalid
	}

	logger.Context(ctx).WithField("documentId", docId).Infof("Request deleting document")

	err = service.db.DocumentStore.MarkDocumentDeleted(userId, docId)
	if err != nil {
		return err
	}

	err = service.search.DeleteDocument(docId, userId)
	if err != nil {
		return fmt.Errorf("delete document from search index: %v", err)
	}
	return nil
}

func (service *DocumentService) RestoreDeletedDocument(ctx context.Context, docId string, userId int) (*models.Document, error) {
	document, err := service.db.DocumentStore.GetDocument(docId)
	if err != nil {
		return nil, err
	}
	if !document.DeletedAt.Valid {
		return nil, errors.ErrRecordNotFound
	}
	document.Update()

	err = service.db.DocumentStore.MarkDocumentNonDeleted(userId, docId)
	if err != nil {
		return nil, err
	}

	doc, err := service.db.DocumentStore.GetDocument(docId)
	err = service.search.IndexDocuments(&[]models.Document{*doc}, doc.UserId)
	if err != nil {
		return nil, fmt.Errorf("delete document from search index: %v", err)
	}
	return document, nil
}

func (service *DocumentService) BulkEditDocuments(ctx context.Context, req *aggregates.BulkEditDocumentsRequest, userId int) error {
	if len(req.AddMetadata) > 0 {
		addMetadata := req.AddMetadata.ToMetadataArray()
		keys := req.AddMetadata.UniqueKeys()
		ok, err := service.db.MetadataStore.UserHasKeys(userId, keys)
		if err != nil {
			return fmt.Errorf("check user owns keys: %v", err)
		}
		if !ok {
			return errors.ErrRecordNotFound
		}
		err = service.db.MetadataStore.UpsertDocumentMetadata(userId, req.Documents, addMetadata)
		if err != nil {
			return err
		}
	}
	if len(req.RemoveMetadata) > 0 {
		removeMetadata := req.RemoveMetadata.ToMetadataArray()
		keys := req.RemoveMetadata.UniqueKeys()
		ok, err := service.db.MetadataStore.UserHasKeys(userId, keys)
		if err != nil {
			return fmt.Errorf("check user owns keys: %v", err)
		}
		if !ok {
			return errors.ErrRecordNotFound
		}

		err = service.db.MetadataStore.DeleteDocumentsMetadata(userId, req.Documents, removeMetadata)
		if err != nil {
			return err
		}
	}

	dateIsValid := req.Date != 0
	langIsValid := req.Lang != ""

	var date time.Time
	var lang models.Lang

	if dateIsValid {
		date = time.Unix(req.Date/1000, 0)
	}
	if langIsValid {
		lang = models.Lang(req.Lang)
	}

	if req.Lang != "" || req.Date != 0 {
		err := service.db.DocumentStore.BulkUpdateDocuments(userId, req.Documents, lang, date)
		if err != nil {
			return err
		}
	}

	// need to reindex
	err := service.db.JobStore.AddDocuments(userId, req.Documents, []models.ProcessStep{models.ProcessFts})
	if err != nil {
		if errors.Is(err, errors.ErrAlreadyExists) {
			// already indexing, skip
		} else {
			return err
		}
	}
	service.process.PullDocumentsToProcess()
	return nil
}

func (service *DocumentService) UpdateDocument(ctx context.Context, userId int, docId string, updated *aggregates.DocumentUpdate) (*models.Document, error) {
	doc, err := service.db.DocumentStore.GetDocument(docId)
	if err != nil {
		return nil, err
	}

	if len(updated.Metadata) > 0 {
		uniqueKeys := updated.Metadata.UniqueKeys()
		owns, err := service.db.MetadataStore.UserHasKeys(userId, uniqueKeys)
		if err != nil {
			return nil, err
		}
		if !owns {
			return nil, errors.ErrRecordNotFound
		}
	}

	if !updated.Date.IsZero() {
		doc.Date = updated.Date
	}

	doc.Name = updated.Name
	doc.Description = updated.Description
	doc.Filename = updated.Filename
	metadata := make([]models.Metadata, len(updated.Metadata))
	if updated.Lang != "" {
		doc.Lang = models.Lang(updated.Lang)
	}

	for i, v := range updated.Metadata {
		metadata[i] = models.Metadata{
			KeyId:   v.KeyId,
			ValueId: v.ValueId,
		}
	}
	doc.Update()
	doc.Metadata = metadata

	err = service.db.DocumentStore.Update(userId, doc)
	if err != nil {
		return nil, err
	}

	err = service.db.MetadataStore.UpdateDocumentKeyValues(userId, doc.Id, metadata)
	if err != nil {
		return nil, err
	}
	err = service.db.JobStore.ForceProcessingDocument(doc.Id, []models.ProcessStep{models.ProcessFts})
	if err != nil {
		logger.Context(ctx).Warnf("error marking document for processing (doc %s): %v", doc.Id, err)
	} else {
		err = service.process.AddDocumentForProcessing(doc.Id)
		if err != nil {
			logger.Context(ctx).Warnf("error adding updated document for processing (doc: %s): %v", doc.Id, err)
		}
	}
	return doc, nil
}

func (service *DocumentService) RequestProcessing(ctx context.Context, userId int, docId string) error {
	steps := append(process.RequiredProcessingSteps(models.ProcessRules), models.ProcessRules)
	err := service.db.JobStore.ForceProcessingDocument(docId, steps)
	if err != nil {
		return err
	}

	doc, err := service.db.DocumentStore.GetDocument(docId)
	if err != nil {
		return fmt.Errorf("get document: %v", err)
	} else {
		err = service.process.AddDocumentForProcessing(doc.Id)
		if err != nil {
			return fmt.Errorf("schedule document processing: %v", err)
		}
	}
	return nil
}

func (service *DocumentService) GetHistory(ctx context.Context, userId int, docId string) (*[]models.DocumentHistory, error) {
	return service.db.DocumentStore.GetDocumentHistory(userId, docId)
}

func (service *DocumentService) GetLinkedDocuments(ctx context.Context, userId int, docId string) ([]*models.LinkedDocument, error) {
	return service.db.MetadataStore.GetLinkedDocuments(userId, docId)
}

func (service *DocumentService) UpdateLinkedDocuments(ctx context.Context, userId int, targetDoc string, linkedDocs []string) error {
	if len(targetDoc) > 100 {
		e := errors.ErrInvalid
		e.ErrMsg = "Maximum number of linked documents is 100"
		return e
	}

	ownership, err := service.db.DocumentStore.UserOwnsDocuments(userId, append(linkedDocs, targetDoc))
	if err != nil {
		return err
	}
	if !ownership {
		return errors.ErrRecordNotFound
	}

	err = service.db.MetadataStore.UpdateLinkedDocuments(userId, targetDoc, linkedDocs)
	if err != nil {
		return err
	}

	docIds := make([]string, len(linkedDocs)+1)
	docIds[0] = targetDoc
	for i, _ := range linkedDocs {
		docIds[i+1] = linkedDocs[i]
	}

	err = service.db.DocumentStore.SetModifiedAt(docIds, time.Now())
	if err != nil {
		logger.Context(ctx).Errorf("update document updated_at when linking documents, docId: %s: %v", targetDoc, err)
		return err
	}
	return nil

}
