/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package process

import (
	"testing"
	"tryffel.net/go/virtualpaper/models"
)

func Test_documentMatchesFilter(t *testing.T) {

	doc := &models.Document{
		Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo " +
			"consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat " +
			"nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui " +
			"officia deserunt mollit anim id est laborum.",
	}

	type args struct {
		document *models.Document
		rule     models.Rule
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple exact match",
			args: args{
				document: doc,
				rule: models.Rule{
					Type:   "exact",
					Filter: "ut Enim ad minim veniam",
				},
			},
			want:    "ut enim ad minim veniam",
			wantErr: false,
		},
		{
			name: "simple exact no match",
			args: args{
				document: doc,
				rule: models.Rule{
					Type:   "exact",
					Filter: "at Enim",
				},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "regex match",
			args: args{
				document: doc,
				rule: models.Rule{
					Type:   "regex",
					Filter: "tempor",
				},
			},
			want:    "tempor",
			wantErr: false,
		},
		{
			name: "regex sub match",
			args: args{
				document: doc,
				rule: models.Rule{
					Type:   "regex",
					Filter: "tempor (incididunt)",
				},
			},
			want:    "incididunt",
			wantErr: false,
		},
		{
			name: "invalid regex",
			args: args{
				document: doc,
				rule: models.Rule{
					Type:   "regex",
					Filter: "tempor (incididunt) )))",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "sub-regex with escape",
			args: args{
				document: doc,
				rule: models.Rule{
					Type:   "regex",
					Filter: "tempor \\(incididunt\\)",
				},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := documentMatchesFilter(tt.args.document, tt.args.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("documentMatchesFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("documentMatchesFilter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applyRule(t *testing.T) {
	type args struct {
		document *models.Document
		rule     models.Rule
		match    string
		validate func(document *models.Document)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "add metadata",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action:          models.RuleActionAddMetadata,
						MetadataKeyId:   1,
						MetadataValueId: 10,
					},
				},
				match: "anything",
				validate: func(doc *models.Document) {
					if len(doc.Metadata) == 0 {
						t.Errorf("no metadata on document")
						return
					}

					if doc.Metadata[0].KeyId != 1 {
						t.Errorf("wrong metadata key id")
					}
					if doc.Metadata[0].ValueId != 10 {
						t.Errorf("wrong metadata value id")
					}
				},
			},
			wantErr: false,
		},
		{
			name: "rename",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action: models.RuleActionRename,
					},
				},
				match: "some match inside document",
				validate: func(doc *models.Document) {
					if doc.Name != "some match inside document" {
						t.Errorf("invalid document name")
					}
				},
			},
			wantErr: false,
		},
		{
			name: "add tag",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action: models.RuleActionAddTag,
						Tag:    10,
					},
				},
				validate: func(doc *models.Document) {
					if len(doc.Tags) == 0 {
						t.Errorf("no tags in document")
						return
					}

					if doc.Tags[0].Id != 10 {
						t.Errorf("wrong tag id")
					}
				},
			},
			wantErr: false,
		},
		{
			name: "set description, empty on start",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action: models.RuleActionSetDescription,
					},
				},
				match: "some description found",
				validate: func(doc *models.Document) {
					if doc.Description != "some description found" {
						t.Errorf("invalid description")
					}
				},
			},
			wantErr: false,
		},
		{
			name: "set description, not-empty on start",
			args: args{
				document: &models.Document{
					Description: "initial description",
				},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action: models.RuleActionSetDescription,
					},
				},
				match: "some description found",
				validate: func(doc *models.Document) {
					if doc.Description != "initial description\n\nsome description found" {
						t.Errorf("invalid description")
					}
				},
			},
			wantErr: false,
		},
		{
			name: "valid date",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action:  models.RuleActionSetDate,
						DateFmt: "2006-01-02",
					},
				},
				match: "2020-11-14",
				validate: func(doc *models.Document) {
					y, m, d := doc.Date.Date()

					if y != 2020 || m != 11 || d != 14 {
						t.Errorf("invalid date")
					}
				},
			},
			wantErr: false,
		},
		{
			name: "another valid date",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action:        models.RuleActionSetDate,
						DateFmt:       "2006.01.02",
						DateSeparator: ".",
					},
				},
				match: "2020.11.14",
				validate: func(doc *models.Document) {
					y, m, d := doc.Date.Date()
					if y != 2020 || m != 11 || d != 14 {
						t.Errorf("invalid date")
					}
				},
			},
			wantErr: false,
		},
		{
			name: "invalid date format",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action:  models.RuleActionSetDate,
						DateFmt: "2006-01-ab",
					},
				},
				match: "2020.11.14",
				validate: func(doc *models.Document) {
				},
			},
			wantErr: true,
		},
		{
			name: "invalid date match",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action:  models.RuleActionSetDate,
						DateFmt: "2006-01-02",
					},
				},
				match: "2020.11.ab",
				validate: func(doc *models.Document) {
				},
			},
			wantErr: true,
		},
		{
			name: "fix leading zeros on date",
			args: args{
				document: &models.Document{},
				rule: models.Rule{
					Action: models.RuleActionConfig{
						Action:        models.RuleActionSetDate,
						DateFmt:       "2006-01-02",
						DateSeparator: "-",
					},
				},
				match: "2020-11-5",
				validate: func(doc *models.Document) {
					y, m, d := doc.Date.Date()
					if y != 2020 || m != 11 || d != 5 {
						t.Errorf("invalid date")
					}
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyRule(tt.args.document, tt.args.rule, tt.args.match); (err != nil) != tt.wantErr {
				t.Errorf("applyRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
