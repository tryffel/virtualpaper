package process

import (
	"database/sql"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

func TestDocumentRule_matchText(t *testing.T) {

	doc := &models.Document{
		Timestamp:   models.Timestamp{},
		Id:          "1234",
		UserId:      1,
		Name:        "a Test Document.5",
		Description: "This is a test document",
		Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo " +
			"consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat " +
			"nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui " +
			"officia deserunt mollit anim id est laborum.",
		Filename:  "",
		Hash:      "",
		Mimetype:  "",
		Size:      0,
		Date:      time.Time{},
		Metadata:  nil,
		Tags:      nil,
		DeletedAt: sql.NullTime{},
	}

	rule := &models.Rule{
		Id:          0,
		UserId:      2,
		Name:        "test rule",
		Description: "",
		Enabled:     true,
		Order:       0,
		Mode:        models.RuleMatchAll,
		Timestamp:   models.Timestamp{},
		Conditions: []*models.RuleCondition{
			{
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionNameIs,
				IsRegex:         false,
				Value:           "a Test Document.5",
			},
			{
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionNameStarts,
				IsRegex:         false,
				Value:           "a test",
			},
			{
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionNameContains,
				IsRegex:         false,
				Value:           "document",
			},
			{
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionDescriptionIs,
				IsRegex:         false,
				Value:           "This is a test document",
			},
			{
				CaseInsensitive: true,
				Inverted:        true,
				ConditionType:   models.RuleConditionDescriptionIs,
				IsRegex:         false,
				Value:           "aThis is a test document",
			},
			{
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionDescriptionStarts,
				IsRegex:         false,
				Value:           "this is",
			},
			{
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionDescriptionContains,
				IsRegex:         false,
				Value:           "test",
			},
			{
				CaseInsensitive: true,
				Inverted:        true,
				ConditionType:   models.RuleConditionDescriptionStarts,
				IsRegex:         false,
				Value:           "this is not",
			},
		},
		Actions: []*models.RuleAction{
			{
				OnCondition: true,
				Action:      models.RuleActionSetName,
				Value:       "test",
			},
		},
	}
	dc := NewDocumentRule(doc, rule)
	got, err := dc.Match()
	if err != nil {
		t.Errorf("matchText() error = %v", err)
		return
	}
	if !got {
		t.Errorf("matchText() got = %v", got)
	}
}

func TestDocumentRule_RunActions(t *testing.T) {

	doc := &models.Document{
		Timestamp:   models.Timestamp{},
		Id:          "1234",
		UserId:      1,
		Name:        "a Test Document.5",
		Description: "This is a test document",
		Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo " +
			"consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat " +
			"nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui " +
			"officia deserunt mollit anim id est laborum.",
		Filename:  "",
		Hash:      "",
		Mimetype:  "",
		Size:      0,
		Date:      time.Time{},
		Metadata:  nil,
		Tags:      nil,
		DeletedAt: sql.NullTime{},
	}

	rule := &models.Rule{
		Id:          0,
		UserId:      2,
		Name:        "test rule",
		Description: "",
		Enabled:     true,
		Order:       0,
		Mode:        models.RuleMatchAll,
		Timestamp:   models.Timestamp{},
		Actions: []*models.RuleAction{
			{
				OnCondition: true,
				Action:      models.RuleActionSetName,
				Value:       "test",
			},
			{
				OnCondition: true,
				Action:      models.RuleActionAppendName,
				Value:       ", suffix",
			},
			{
				OnCondition:   true,
				Action:        models.RuleActionAddMetadata,
				MetadataKey:   1,
				MetadataValue: 2,
			},
		},
	}

	wantName := "test, suffix"
	dc := NewDocumentRule(doc, rule)
	err := dc.RunActions()
	if err != nil {
		t.Errorf("matchText() error = %v", err)
		return
	}

	if wantName != doc.Name {
		t.Errorf("runActions(), wantName = %s, got name: %s", wantName, doc.Name)
	}

	if !doc.HasMetadataKeyValue(1, 2) {
		t.Errorf("runActions(), missing metadata key value")
	}

}
