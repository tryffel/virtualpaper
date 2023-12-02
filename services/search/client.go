package search

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
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

func NewEngine(db *storage.Database, conf *config.Meilisearch) (*Engine, error) {
	engine := &Engine{
		Url:    conf.Url,
		ApiKey: conf.ApiKey,
		db:     db,
	}
	err := engine.connect()
	return engine, err
}

// connect creates a connection to meilisearch instance and initializes index if neccessary.
func (e *Engine) connect() error {
	logrus.Infof("connect to meilisearch at %s", e.Url)
	e.client = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    e.Url,
		APIKey:  e.ApiKey,
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

var indexNameToUserIdRegex = regexp.MustCompile("virtualpaper-(.+)")

func (e *Engine) ensureIndexExists() error {
	logrus.Debugf("ensure meilisearch indices exist")
	users, err := e.db.UserStore.GetUsers()
	if err != nil {
		return fmt.Errorf("get users: %v", err)
	}

	for _, v := range *users {
		err = e.AddUserIndex(v.Id)
		if err != nil {
			logrus.Errorf("error checking & creating index for user %d: %v", v.Id, err)
		}

	}
	return nil
}

func (e *Engine) UpdateUserPreferences(userId int) error {

	preferences, err := e.db.UserStore.GetUserPreferences(userId)
	if err != nil {
		return fmt.Errorf("get preferences: %v", err)
	}
	index := indexName(userId)

	_, err = e.client.Index(index).UpdateStopWords(&preferences.StopWords)
	if err != nil {
		return fmt.Errorf("update stopwords: %v", err)
	}

	synonyms := buildSynonyms(preferences.Synonyms)
	_, err = e.client.Index(index).UpdateSynonyms(&synonyms)
	if err != nil {
		return fmt.Errorf("update synonyms: %v", err)
	}
	return nil
}

func (e *Engine) ping() error {
	v, err := e.client.GetVersion()
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
		shares, err := e.db.DocumentStore.GetSharedUsers(e.db, v.Id)
		if err != nil {
			return fmt.Errorf("get shares for document: %v", err)
		}

		sharedUsers := make([]int, len(*shares))
		for i, v := range *shares {
			sharedUsers[i] = v.UserId
		}

		tags := make([]string, len(v.Tags))
		for tagI, tag := range v.Tags {
			tags[tagI] = tag.Key
		}

		metadata := make([]string, len(v.Metadata))
		for metadataI, v := range v.Metadata {
			key := normalizeMetadataKey(v.Key)
			value := normalizeMetadataValue(v.Value)
			metadata[metadataI] = key + ":" + value
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
			"metadata":    metadata,
			"date":        v.Date.Unix(),
			"description": v.Description,
			"mimetype":    v.Mimetype,
			"lang":        v.Lang,
			"shares":      sharedUsers,
		}
	}

	_, err := e.client.Index(indexName(userId)).UpdateDocuments(data)
	if err != nil {
		return fmt.Errorf("index documents: %v", err)
	}

	return nil
}

func (e *Engine) DeleteDocument(docId string, userId int) error {

	_, err := e.client.Index(indexName(userId)).DeleteDocument(docId)
	if err != nil {
		return fmt.Errorf("delete documents: %v", err)
	}
	return nil
}

func buildSynonyms(synonyms [][]string) map[string][]string {
	output := map[string][]string{}
	for _, tuple := range synonyms {

		for i, word := range tuple {
			if len(tuple) <= 1 {
				continue
			}

			var words []string
			if i == 0 {
				words = tuple[1:]
			} else if i == len(tuple)-1 {
				words = tuple[:i]
			} else {
				// need to iterate over list. This is needed for not mutating the tuple list.
				words = make([]string, 0, len(tuple)-1)
				for index, v := range tuple {
					if index == i {
						continue
					}
					words = append(words, v)
				}
			}
			output[word] = words
		}
	}

	return output
}

func (e *Engine) GetHealth() (string, bool, error) {
	if e.client.IsHealthy() {
		return "available", true, nil
	}

	resp, err := e.client.Health()
	if err != nil {
		e := errors.ErrInternalError
		e.Err = err
		e.ErrMsg = "cannot query meilisearch health"
		return resp.Status, false, e

	}
	return resp.Status, false, err
}

type EngineStatus struct {
	Ok      bool   `json:"engine_ok"`
	Status  string `json:"status"`
	Version string `json:"version"`
	Name    string `json:"name"`
}

func (e *Engine) GetStatus() (*EngineStatus, error) {
	status := &EngineStatus{}
	status.Name = "Meilisearch"

	version, err := e.client.Version()
	if err != nil {
		return status, err
	}

	status.Version = version.PkgVersion
	if !e.client.IsHealthy() {
		status.Ok = false
		status.Status = "error"
	} else {
		status.Ok = true
		status.Status = "available"
	}
	return status, nil

}

type UserIndexStatus struct {
	UserId       int  `json:"user_id"`
	NumDocuments int  `json:"documents_count"`
	Indexing     bool `json:"indexing"`
}

func (e *Engine) GetUserIndexStatus(userId int) (UserIndexStatus, error) {
	stats, err := e.client.Index(indexName(userId)).GetStats()
	if err != nil {
		return UserIndexStatus{}, err
	}
	stat := UserIndexStatus{
		UserId:       userId,
		NumDocuments: int(stats.NumberOfDocuments),
		Indexing:     stats.IsIndexing,
	}
	return stat, err
}

// GetUserIndicesStatus returns list of indices, total db size in bytes and possible error.
func (e *Engine) GetUserIndicesStatus() (map[int]*UserIndexStatus, int64, error) {
	var err errors.Error
	stats, meiliErr := e.client.GetStats()
	if meiliErr != nil {
		err = errors.ErrInternalError
		err.Err = meiliErr
		err.ErrMsg = "get stats"

		return map[int]*UserIndexStatus{}, 0, err
	}

	indices := make(map[int]*UserIndexStatus, len(stats.Indexes))
	counter := 0
	for i, v := range stats.Indexes {
		user := indexNameToUserIdRegex.FindStringSubmatch(i)
		if len(user) != 2 {
			logrus.Warningf("not virtualpaper index: %s, skipping", user)
		}
		userId, err := strconv.Atoi(user[1])
		if err != nil {
			logrus.Warningf("not virtualpaper index: %s, skipping", user[1])
			continue
		}
		userStats := &UserIndexStatus{
			UserId:       userId,
			Indexing:     v.IsIndexing,
			NumDocuments: int(v.NumberOfDocuments),
		}
		indices[userId] = userStats
		counter += 1
	}
	return indices, stats.DatabaseSize, nil
}

func (e *Engine) DeleteDocuments(userId int) error {
	index := indexName(userId)
	_, err := e.client.Index(index).DeleteAllDocuments()
	if err != nil {
		return fmt.Errorf("delete index: %v", err)
	}
	return nil
}

func (e *Engine) AddUserIndex(userId int) error {
	index := indexName(userId)
	indexExists := false
	logrus.Debugf("ensure meilisearch index %s exists", index)
	var err error
	_, err = e.client.GetIndex(index)
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
		logrus.Warningf("Creating new meilisearch index '%s'", index)
		_, err = e.client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        index,
			PrimaryKey: "document_id",
		})

		fields := &[]string{
			"document_id",
			"user_id",
			"name",
			"file_name",
			"content",
			"hash",
			"created_at",
			"updated_at",
			"tags",
			"metadata",
			"date",
			"description",
			"tags",
			"metadata_key",
			"metadata_value",
			"mimetype",
			"lang",
			"shared_users",
		}
		_, err = e.client.Index(index).UpdateFilterableAttributes(fields)
		if err != nil {
			logrus.Errorf("meilisearch set filterable attributes: %v", err)
		}

		_, err = e.client.Index(index).UpdateSortableAttributes(fields)
		if err != nil {
			logrus.Errorf("meilisearch set sortable attributes: %v", err)
		}
		_, err = e.client.Index(index).UpdateSearchableAttributes(fields)
		if err != nil {
			logrus.Errorf("meilisearch set searchable attributes: %v", err)
		}
	}
	if err != nil {
		return fmt.Errorf("create index: %v", err)
	}
	return nil
}
