package process

import (
	"database/sql"
	"reflect"
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
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionNameIs,
				IsRegex:         false,
				Value:           "a Test Document.5",
			},
			{
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionNameStarts,
				IsRegex:         false,
				Value:           "a test",
			},
			{
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionNameContains,
				IsRegex:         false,
				Value:           "document",
			},
			{
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionDescriptionIs,
				IsRegex:         false,
				Value:           "This is a test document",
			},
			{
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        true,
				ConditionType:   models.RuleConditionDescriptionIs,
				IsRegex:         false,
				Value:           "aThis is a test document",
			},
			{
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionDescriptionStarts,
				IsRegex:         false,
				Value:           "this is",
			},
			{
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        false,
				ConditionType:   models.RuleConditionDescriptionContains,
				IsRegex:         false,
				Value:           "test",
			},
			{
				Enabled:         true,
				CaseInsensitive: true,
				Inverted:        true,
				ConditionType:   models.RuleConditionDescriptionStarts,
				IsRegex:         false,
				Value:           "this is not",
			},
		},
		Actions: []*models.RuleAction{
			{
				Enabled:     true,
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
		Filename: "",
		Hash:     "",
		Mimetype: "",
		Size:     0,
		Date:     time.Time{},
		Metadata: []models.Metadata{
			{KeyId: 10, ValueId: 15},
		},
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
				Enabled:     true,
				OnCondition: true,
				Action:      models.RuleActionSetName,
				Value:       "test",
			},
			{
				Enabled:     true,
				OnCondition: true,
				Action:      models.RuleActionAppendName,
				Value:       ", suffix",
			},
			{
				Enabled:       true,
				OnCondition:   true,
				Action:        models.RuleActionAddMetadata,
				MetadataKey:   1,
				MetadataValue: 2,
			},
			{
				Enabled:       true,
				OnCondition:   true,
				Action:        models.RuleActionAddMetadata,
				MetadataKey:   2,
				MetadataValue: 3,
			},
			{
				Enabled:       true,
				OnCondition:   true,
				Action:        models.RuleActionRemoveMetadata,
				MetadataKey:   10,
				MetadataValue: 15,
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

	if doc.HasMetadataKeyValue(10, 15) {
		t.Errorf("runActions(), metadata not removed")
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

func BenchmarkDocumentRule_matchTextByDistance_shorttext(b *testing.B) {
	// 66 words to search from.
	match := "a short match"
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo " +
		"consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat " +
		"nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui " +
		"officia deserunt mollit anim id est laborum."

	for i := 0; i < b.N; i++ {
		_, _ = matchTextByDistance(match, text, 1, false, false)
	}
}

func BenchmarkDocumentRule_matchTextByDistance_mediumtext(b *testing.B) {
	// 6600 words to search from.
	match := "a short match"
	text := `Lorem ipsum dolor sit amet, consectetur adipiscing elit,
		sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
		Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo
		consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat
		nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui 
		officia deserunt mollit anim id est laborum.`

	longText := ""

	b.StopTimer()
	for i := 0; i < 100; i++ {
		longText += text
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = matchTextByDistance(match, longText, 1, false, false)
	}
}

func BenchmarkDocumentRule_matchTextByDistance_longtext(b *testing.B) {
	// 660000 words to search from.
	match := "a short match"
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo " +
		"consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat " +
		"nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui " +
		"officia deserunt mollit anim id est laborum."

	longText := ""

	b.StopTimer()
	for i := 0; i < 10000; i++ {
		longText += text
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = matchTextByDistance(match, longText, 1, false, false)
	}
}

func Test_removeMetadata(t *testing.T) {
	type args struct {
		doc     *models.Document
		keyId   int
		valueId int
	}
	tests := []struct {
		name    string
		args    args
		wantDoc *models.Document
	}{
		{
			name: "no matching metadata",
			args: args{
				doc:     &models.Document{Metadata: []models.Metadata{{KeyId: 1, ValueId: 2}}},
				keyId:   1,
				valueId: 5,
			},
			wantDoc: &models.Document{Metadata: []models.Metadata{{KeyId: 1, ValueId: 2}}},
		},
		{
			name: "delete one key-value",
			args: args{
				doc:     &models.Document{Metadata: []models.Metadata{{KeyId: 1, ValueId: 2}, {KeyId: 1, ValueId: 3}}},
				keyId:   1,
				valueId: 2,
			},
			wantDoc: &models.Document{Metadata: []models.Metadata{{KeyId: 1, ValueId: 3}}},
		},
		{
			name: "delete by key",
			args: args{
				doc: &models.Document{Metadata: []models.Metadata{
					{KeyId: 1, ValueId: 2},
					{KeyId: 1, ValueId: 3},
					{KeyId: 2, ValueId: 2}}},
				keyId:   1,
				valueId: 0,
			},
			wantDoc: &models.Document{Metadata: []models.Metadata{{KeyId: 2, ValueId: 2}}},
		},
		{
			name: "delete all metadata",
			args: args{
				doc: &models.Document{Metadata: []models.Metadata{
					{KeyId: 1, ValueId: 2},
					{KeyId: 1, ValueId: 3},
					{KeyId: 1, ValueId: 4}}},
				keyId:   1,
				valueId: 0,
			},
			wantDoc: &models.Document{Metadata: []models.Metadata{}},
		},
		{
			name: "no matching metadata",
			args: args{
				doc: &models.Document{Metadata: []models.Metadata{
					{KeyId: 1, ValueId: 2},
					{KeyId: 1, ValueId: 3},
					{KeyId: 1, ValueId: 4}}},
				keyId:   2,
				valueId: 0,
			},
			wantDoc: &models.Document{Metadata: []models.Metadata{
				{KeyId: 1, ValueId: 2},
				{KeyId: 1, ValueId: 3},
				{KeyId: 1, ValueId: 4},
			},
			},
		},
		{
			name: "empty document",
			args: args{
				doc:     &models.Document{},
				keyId:   1,
				valueId: 0,
			},
			wantDoc: &models.Document{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeMetadata(tt.args.doc, tt.args.keyId, tt.args.valueId)
			if !reflect.DeepEqual(tt.args.doc, tt.wantDoc) {
				t.Errorf("doc metadata differs: want %d, got %d",
					len(tt.wantDoc.Metadata), len(tt.args.doc.Metadata))
			}
		})
	}
}

func TestDocumentRule_extractDates(t *testing.T) {
	now := time.Unix(1627620345, 0)

	type fields struct {
		Rule     *models.Rule
		Document *models.Document
		date     time.Time
	}
	type args struct {
		condition *models.RuleCondition
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     bool
		wantErr  bool
		wantDate string
	}{
		{
			name: "three dates",
			fields: fields{
				Document: &models.Document{
					Content: `2021-07-30  2020-07-30 2020-07-30 2020-07-30 Lorem ipsum dolor sit amet, 
	consectetur adipiscing elit, 2021-07-31 2021-07-30`},
				date: time.Time{},
			},
			args: args{
				condition: &models.RuleCondition{
					ConditionType: models.RuleConditionDateIs,
					Value:         "(\\d{4}-\\d{1,2}-\\d{1,2})",
					DateFmt:       "2006-01-02",
				},
			},
			want:     true,
			wantErr:  false,
			wantDate: "2021-07-31",
		},
		{
			name: "invalid regex",
			fields: fields{
				Document: &models.Document{
					Content: `2021-07-30 Lorem ipsum dolor sit amet, consectetur adipiscing elit, 2021-07-31 
	2021-07-30`},
				date: time.Time{},
			},
			args: args{
				condition: &models.RuleCondition{
					ConditionType: models.RuleConditionDateIs,
					Value:         "(\\d{4}-\\d{1,2}-\\d{1,2}))))",
					DateFmt:       "2006-01-02",
				},
			},
			want:     false,
			wantErr:  true,
			wantDate: "2021-07-31",
		},
		{
			name: "past date",
			fields: fields{
				Document: &models.Document{
					Content: `2020-07-30 Lorem ipsum dolor sit amet, consectetur adipiscing elit, 2020-07-31 
	2020-07-30`},
				date: time.Time{},
			},
			args: args{
				condition: &models.RuleCondition{
					ConditionType: models.RuleConditionDateIs,
					Value:         "(\\d{4}-\\d{1,2}-\\d{1,2})",
					DateFmt:       "2006-01-02",
				},
			},
			want:     true,
			wantErr:  false,
			wantDate: "2020-07-30",
		},
		{
			name: "upcoming date overrides past date",
			fields: fields{
				Document: &models.Document{
					Content: `2020-07-30 Lorem ipsum dolor sit amet, consectetur adipiscing elit, 2021-07-31 
	2020-07-31`},
				date: time.Time{},
			},
			args: args{
				condition: &models.RuleCondition{
					ConditionType: models.RuleConditionDateIs,
					Value:         "(\\d{4}-\\d{1,2}-\\d{1,2})",
					DateFmt:       "2006-01-02",
				},
			},
			want:     true,
			wantErr:  false,
			wantDate: "2021-07-31",
		},
		{
			name: "no date found",
			fields: fields{
				Document: &models.Document{
					Content: ""},
			},
			args: args{
				condition: &models.RuleCondition{
					ConditionType: models.RuleConditionDateIs,
					Value:         "(\\d{4}-\\d{1,2}-\\d{1,2})",
					DateFmt:       "2006-01-02",
				},
			},
			want:     false,
			wantErr:  false,
			wantDate: "0001-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DocumentRule{
				Rule:     tt.fields.Rule,
				Document: tt.fields.Document,
				date:     tt.fields.date,
			}
			got, err := d.extractDates(tt.args.condition, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractDates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractDates() got = %v, want %v", got, tt.want)
			}

			if tt.want {
				wantDate, err := time.Parse("2006-01-02", tt.wantDate)
				if err != nil {
					t.Errorf("invalid date format: %v", err)
				}
				if !d.date.Equal(wantDate) {
					t.Errorf("date does not match, want: %s, got: %s", wantDate.String(), d.date.String())
				}
			}
		})
	}
}
