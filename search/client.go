package search

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

// Engine is as search engine that uses Meilisearch to provide full-text-search
// across documents.
type Engine struct {
	client *meilisearch.Client
	db     *storage.Database
	Url    string
	ApiKey string
}

func NewEngine(db *storage.Database) (*Engine, error) {
	engine := &Engine{
		Url:    config.C.Meilisearch.Url,
		ApiKey: config.C.Meilisearch.ApiKey,
		db:     db,
	}
	err := engine.connect()
	return engine, err
}

// connect creates a connection to meilisearch instance and initializes index if neccessary.
func (e *Engine) connect() error {
	logrus.Infof("connect to meilisearch at %s", e.Url)
	e.client = meilisearch.NewClient(meilisearch.Config{
		Host:   e.Url,
		APIKey: e.ApiKey,
	})

	e.client = meilisearch.NewClientWithCustomHTTPClient(meilisearch.Config{
		Host:   e.Url,
		APIKey: e.ApiKey,
	}, http.Client{
		Timeout: 10 * time.Second,
	})

	err := e.ping()
	if err != nil {
		return fmt.Errorf("cannot connect to meilisearch: %v", err)
	}
	return e.ensureIndexExists()
}

func indexName(userid int) string {
	return "virtualpaper-" + strconv.Itoa(userid)
}

func (e *Engine) ensureIndexExists() error {
	logrus.Debugf("ensure meilisearch indices exist")
	users, err := e.db.UserStore.GetUsers()
	if err != nil {
		return fmt.Errorf("get users: %v", err)
	}

	indices := make([]string, len(*users))
	for i, v := range *users {
		indices[i] = indexName(v.Id)
		indexExists := false
		logrus.Debugf("ensure meilisearch index %s exists", indices[i])
		var err error
		_, err = e.client.Indexes().Get(indices[i])
		if err != nil {
			if e, ok := err.(*meilisearch.Error); ok {
				if e.StatusCode == 404 {
					err = nil
					indexExists = false
				}
			}
		} else {
			indexExists = true
		}
		if err != nil {
			return fmt.Errorf("get indexes: %v", err)
		}

		if !indexExists {
			logrus.Warningf("Creating new meilisearch index '%s'", indices[i])
			_, err = e.client.Indexes().Create(meilisearch.CreateIndexRequest{
				UID:        indices[i],
				PrimaryKey: "document_id",
			})

			_, err = e.client.Settings(indices[i]).UpdateAttributesForFaceting([]string{
				"tags",
				"metadata_key",
				"metadata_value",
			})
			if err != nil {
				logrus.Errorf("meilisearch set faceted search attributes: %v", err)

			}
		}
		if err != nil {
			return fmt.Errorf("create index: %v", err)
		}
	}
	return nil
}

func (e *Engine) ping() error {
	v, err := e.client.Version().Get()
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("cannot connect to meilisearch server: %v", err)
		}
		return fmt.Errorf("get version: %v", err)
	}

	logrus.Infof("meilisearch version: %s", v.PkgVersion)
	return nil
}

// IndexDocuments sends documents to meilisearch for indexing
func (e *Engine) IndexDocuments(docs *[]models.Document, userId int) error {
	data := make([]map[string]interface{}, len(*docs))
	for i, v := range *docs {

		tags := make([]string, len(v.Tags))
		for tagI, tag := range v.Tags {
			tags[tagI] = tag.Key
		}

		data[i] = map[string]interface{}{
			"document_id": v.Id,
			"user_id":     v.UserId,
			"name":        v.Name,
			"file_name":   v.Filename,
			"content":     v.Content,
			"hash":        v.Hash,
			"created_at":  v.CreatedAt.Unix(),
			"updated_at":  v.UpdatedAt.Unix(),
			"tags":        tags,
			"metadata":    v.Metadata,
			"date":        v.Date.Unix(),
			"description": v.Description,
		}
	}

	_, err := e.client.Documents(indexName(userId)).AddOrReplace(data)
	if err != nil {
		return fmt.Errorf("index documents: %v", err)
	}

	return nil
}

func documentToDto(doc *models.Document) *map[string]interface{} {
	return &map[string]interface{}{
		"document_id": doc.Id,
		"user_id":     doc.UserId,
		"name":        doc.Name,
		"file_name":   doc.Filename,
		"content":     doc.Content,
		"hash":        doc.Hash,
		"created_at":  doc.CreatedAt.String(),
	}
}
