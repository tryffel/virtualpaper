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
	Id int
	UserId int
	Name string
	Description string
	Enabled bool
	Order int
	Mode RuleConditionMatchType
	Timestamp
}


type RuleConditionType string

const (
	RuleConditionNameIs RuleConditionType = "name_is"
	RuleConditionNameStarts RuleConditionType = "name_starts"
	RuleConditionNameContains RuleConditionType = "name_contains"

	RuleConditionDescriptionIs RuleConditionType = "description_is"
	RuleConditionDescriptionStarts RuleConditionType = "description_starts"
	RuleConditionDescriptionContains RuleConditionType = "description_contains"

	RuleConditionContentIs RuleConditionType = "content_is"
	RuleConditionContentStarts RuleConditionType = "content_starts"
	RuleConditionContentContains RuleConditionType = "content_contains"

	RuleConditionDateIs RuleConditionType = "date_is"
	RuleConditionDateAfter RuleConditionType = "date_after"
	RuleConditionDateBefore RuleConditionType = "date_before"

	RuleConditionMetadataHasKey RuleConditionType = "metadata_has_key"
	RuleConditionMetadataHasKeyValue RuleConditionType = "metadata_has_key_value"
	RuleConditionMetadataCount RuleConditionType = "metadata_count"
)


type RuleCondition struct {
	Id int
	RuleId int
	Enabled int
	CaseInsensitive bool
	// Inverted inverts the match result
	Inverted bool
	ConditionType RuleConditionType

	// IsRegex defines whether to apply regex pattern
	IsRegex bool
	// Value to compare against, if text field
	Value string

	// Metadata to operate with
	MetadataKey int
	MetadataValue int
}


type RuleActionType string

const (
	RuleActionSetName RuleConditionType = "name_set"
	RuleActionAppendName RuleConditionType = "name_append"
	RuleActionSetDescription RuleConditionType = "description_set"
	RuleActionAppendDescription RuleConditionType = "description_append"
	RuleActionAddMetadata RuleConditionType = "metadata_add"
	RuleActionRemoveMetadata RuleConditionType = "metadata_remove"
	RuleActionSetDate RuleConditionType = "date_set"
)

type RuleAction struct {
	Id int
	RuleId int
	// OnCondition, if vs else
	OnCondition bool

	Action RuleActionType
	Value string
	MetadataKey int
	MetadataValue
}

