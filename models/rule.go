/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2021  Tero Vierimaa
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

package models

type RuleConditionMatchType int

const (
	// RuleMatchAll requires all conditions must be matched
	RuleMatchAll RuleConditionMatchType = 1
	//RuleMatchAny allows any condition to match
	RuleMatchAny RuleConditionMatchType = 2
)

type Rule struct {
	Id          int                    `db:"id"`
	UserId      int                    `db:"user_id"`
	Name        string                 `db:"name"`
	Description string                 `db:"description"`
	Enabled     bool                   `db:"enabled"`
	Order       int                    `db:"rule_order"`
	Mode        RuleConditionMatchType `db:"mode"`
	Timestamp

	Conditions []*RuleCondition
	Actions    []*RuleAction
}

type RuleConditionType string

const (
	RuleConditionNameIs       RuleConditionType = "name_is"
	RuleConditionNameStarts   RuleConditionType = "name_starts"
	RuleConditionNameContains RuleConditionType = "name_contains"

	RuleConditionDescriptionIs       RuleConditionType = "description_is"
	RuleConditionDescriptionStarts   RuleConditionType = "description_starts"
	RuleConditionDescriptionContains RuleConditionType = "description_contains"

	RuleConditionContentIs       RuleConditionType = "content_is"
	RuleConditionContentStarts   RuleConditionType = "content_starts"
	RuleConditionContentContains RuleConditionType = "content_contains"

	RuleConditionDateIs     RuleConditionType = "date_is"
	RuleConditionDateAfter  RuleConditionType = "date_after"
	RuleConditionDateBefore RuleConditionType = "date_before"

	RuleConditionMetadataHasKey        RuleConditionType = "metadata_has_key"
	RuleConditionMetadataHasKeyValue   RuleConditionType = "metadata_has_key_value"
	RuleConditionMetadataCount         RuleConditionType = "metadata_count"
	RuleConditionMetadataCountLessThan RuleConditionType = "metadata_count_less_than"
	RuleConditionMetadataCountMoreThan RuleConditionType = "metadata_count_more_than"
)

type RuleCondition struct {
	Id              int  `db:"id"`
	RuleId          int  `db:"rule_id"`
	Enabled         bool `db:"enabled"`
	CaseInsensitive bool `db:"case_insensitive"`
	// Inverted inverts the match result
	Inverted      bool              `db:"inverted_match"`
	ConditionType RuleConditionType `db:"condition_type"`

	// IsRegex defines whether to apply regex pattern
	IsRegex bool `db:"is_regex"`
	// Value to compare against, if text field
	Value string `db:"value"`

	// Metadata to operate with
	MetadataKey   IntId `db:"metadata_key"`
	MetadataValue IntId `db:"metadata_value"`
}

type RuleActionType string

const (
	RuleActionSetName           RuleActionType = "name_set"
	RuleActionAppendName        RuleActionType = "name_append"
	RuleActionSetDescription    RuleActionType = "description_set"
	RuleActionAppendDescription RuleActionType = "description_append"
	RuleActionAddMetadata       RuleActionType = "metadata_add"
	RuleActionRemoveMetadata    RuleActionType = "metadata_remove"
	RuleActionSetDate           RuleActionType = "date_set"
)

type RuleAction struct {
	Id      int  `db:"id"`
	RuleId  int  `db:"rule_id"`
	Enabled bool `db:"enabled"`
	// OnCondition, if vs else
	OnCondition bool `db:"on_condition"`

	Action        RuleActionType `db:"action"`
	Value         string         `db:"value"`
	MetadataKey   IntId          `db:"metadata_key"`
	MetadataValue IntId          `db:"metadata_value"`
}
