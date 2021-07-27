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

func TestDocumentRule_matchTextByDistance(t *testing.T) {
	type args struct {
		match       string
		text        string
		maxTypos    int
		matchPrefix bool
		matchIs     bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "no typos",
			args: args{
				match:       "a test match",
				text:        "this is a test match indeed",
				maxTypos:    0,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "one typo",
			args: args{
				match:       "a test match",
				text:        "this is a test2match indeed",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "two typos",
			args: args{
				match:       "a test match",
				text:        "this is a1test2match indeed",
				maxTypos:    2,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "two typos, one allowed",
			args: args{
				match:       "a test match",
				text:        "this is a1test2match indeed",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "first nomatch + match",
			args: args{
				match:       "a test match",
				text:        "first a no match a1test matc, second matches a2test m2tch",
				maxTypos:    2,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "skip one character in original text",
			args: args{
				match:       "a test match",
				text:        "empty a testmatch",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "one extra character",
			args: args{
				match:       "a test match",
				text:        "empty a test  match",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "empty text",
			args: args{
				match:       "a test match",
				text:        "",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "empty match",
			args: args{
				match:       "",
				text:        "empty test match",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     false,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "match prefix",
			args: args{
				match:       "empty test",
				text:        "empty 2test match",
				maxTypos:    1,
				matchPrefix: true,
				matchIs:     false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "match failed prefix",
			args: args{
				match:       "empty test",
				text:        "not empty test match",
				maxTypos:    0,
				matchPrefix: true,
				matchIs:     false,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "match is",
			args: args{
				match:       "empty test",
				text:        "enpty test",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     true,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "match too short",
			args: args{
				match:       "empty",
				text:        "enpty test",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     true,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "text too short",
			args: args{
				match:       "empty test",
				text:        "enpty",
				maxTypos:    1,
				matchPrefix: false,
				matchIs:     true,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchTextByDistance(tt.args.match, tt.args.text, tt.args.maxTypos, tt.args.matchPrefix, tt.args.matchIs)
			if (err != nil) != tt.wantErr {
				t.Errorf("matchTextByDistance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("matchTextByDistance() got = %v, want %v", got, tt.want)
			}
		})
	}
}
