package aggregates

import (
	"time"
	"tryffel.net/go/virtualpaper/models"
)

// Document
type Document struct {
	// swagger:strfmt uuid
	Id          string `json:"id"`
	Name        string `json:"name"`
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	// swagger:strfmt either null or unix epoch in milliseconds
	DeletedAt   interface{}       `json:"deleted_at"`
	Date        int64             `json:"date"`
	PreviewUrl  string            `json:"preview_url"`
	DownloadUrl string            `json:"download_url"`
	Mimetype    string            `json:"mimetype"`
	Type        string            `json:"type"`
	Size        int64             `json:"size"`
	PrettySize  string            `json:"pretty_size"`
	Status      string            `json:"status"`
	Metadata    []models.Metadata `json:"metadata"`
	Tags        []models.Tag      `json:"tags"`
	Lang        string            `json:"lang"`
}

func DocumentToAggregate(doc *models.Document) *Document {
	resp := &Document{
		Id:          doc.Id,
		Name:        doc.Name,
		Filename:    doc.Filename,
		Content:     doc.Content,
		Description: doc.Description,
		CreatedAt:   doc.CreatedAt.Unix() * 1000,
		UpdatedAt:   doc.UpdatedAt.Unix() * 1000,
		Date:        doc.Date.Unix() * 1000,
		PreviewUrl:  "",
		DownloadUrl: "",
		Mimetype:    doc.Mimetype,
		Type:        doc.GetType(),
		Size:        doc.Size,
		PrettySize:  doc.GetSize(),
		Metadata:    doc.Metadata,
		Tags:        doc.Tags,
		Lang:        doc.Lang.String(),
	}
	if doc.DeletedAt.Valid {
		resp.DeletedAt = doc.DeletedAt.Time.Unix() * 1000
	} else {
		resp.DeletedAt = nil
	}
	return resp
}

type BulkEditDocumentsRequest struct {
	Documents      []string      `json:"documents"`
	AddMetadata    MetadataArray `json:"add_metadata"`
	RemoveMetadata MetadataArray `json:"remove_metadata"`
	Lang           string        `json:"lang"`
	Date           int64         `json:"date"`
}

type Metadata struct {
	KeyId   int `json:"key_id"`
	ValueId int `json:"value_id"`
}

func (m Metadata) ToMetadata() models.Metadata {
	return models.Metadata{
		KeyId:   m.KeyId,
		ValueId: m.ValueId,
	}
}

type MetadataArray []Metadata

func (m MetadataArray) ToMetadataArray() []models.Metadata {
	metadata := make([]models.Metadata, len(m))
	for i, v := range m {
		metadata[i] = v.ToMetadata()
	}
	return metadata
}

func (m MetadataArray) UniqueKeys() []int {
	keyMap := map[int]bool{}
	for _, v := range m {
		keyMap[v.KeyId] = true
	}

	keys := make([]int, len(keyMap))
	index := 0
	for i, _ := range keyMap {
		keys[index] = i
		index += 1
	}
	return keys
}

type DocumentUpdate struct {
	Name        string
	Description string
	Filename    string
	Date        time.Time
	Metadata    []Metadata
	Lang        string
}