package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

// check given metadata is subset of document's metadata
func assertDocumentMetadataContainsValues(t *testing.T, document *api.DocumentResponse, wantValues []*models.MetadataValue) {
	getKeyValueId := func(value *models.MetadataValue) string {
		return fmt.Sprintf("%d-%s:%d-%s-user-%d", value.KeyId, value.Key, value.Id, value.Value, value.UserId)
	}
	values := map[string]bool{}
	for _, wantValue := range wantValues {
		values[getKeyValueId(wantValue)] = false
		for _, gotValue := range document.Metadata {
			if wantValue.Id == gotValue.ValueId && wantValue.KeyId == gotValue.KeyId {
				values[getKeyValueId(wantValue)] = true
			}
		}
	}
	missingValues := map[string]bool{}
	for i, v := range values {
		if !v {
			missingValues[i] = v
		}
	}
	if len(missingValues) > 0 {
		t.Error("document is missing metadata: ", missingValues)
	}
}

// check document's metadata is subset of given metadata
func assertDocumentMetadataInValues(t *testing.T, document *api.DocumentResponse, wantValues []*models.MetadataValue) {
	getKeyValueId := func(value models.Metadata) string {
		return fmt.Sprintf("%d-%s:%d-%s", value.KeyId, value.Key, value.ValueId, value.Value)
	}
	values := map[string]bool{}
	for _, wantValue := range document.Metadata {
		values[getKeyValueId(wantValue)] = false
		for _, gotValue := range wantValues {
			if wantValue.ValueId == gotValue.Id && wantValue.KeyId == gotValue.KeyId {
				values[getKeyValueId(wantValue)] = true
			}
		}
	}
	missingValues := map[string]bool{}
	for i, v := range values {
		if !v {
			missingValues[i] = v
		}
	}
	if len(missingValues) > 0 {
		t.Error("document has extra metadata: ", missingValues)
	}
}

func assertDocumentMetadataMatches(t *testing.T, document *api.DocumentResponse, wantValues []*models.MetadataValue) {
	assertDocumentMetadataContainsValues(t, document, wantValues)
	assertDocumentMetadataInValues(t, document, wantValues)
}

func assertDocumentInArray(t *testing.T, id string, docs *[]api.DocumentResponse) {
	for _, v := range *docs {
		if v.Id == id {
			return
		}
	}
	t.Errorf("document not found from list of documents")
}

func assertDocumentNotInArray(t *testing.T, id string, docs *[]api.DocumentResponse) {
	for _, v := range *docs {
		if v.Id == id {
			t.Errorf("document found from list of documents")
			return
		}
	}
}

func assertDateMatches(t *testing.T, epoch1 int64, epoch2 int64) {
	t1 := time.Unix(epoch1, 0)
	t2 := time.Unix(epoch2, 0)

	m1 := models.MidnightForDate(t1)
	m2 := models.MidnightForDate(t2)

	assert.True(t, m1.Equal(m2))
}
