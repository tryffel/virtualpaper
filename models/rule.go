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

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"tryffel.net/go/virtualpaper/errors"
)

type RuleConditionMatchType int

func (r RuleConditionMatchType) String() string {
	switch r {
	case RuleMatchAll:
		return "match_all"
	case RuleMatchAny:
		return "match_any"
	default:
		return ""
	}
}

func (r *RuleConditionMatchType) FromString(str string) error {
	switch str {
	case "match_all":
		*r = RuleMatchAll
	case "match_any":
		*r = RuleMatchAny
	default:
		e := errors.ErrInvalid
		e.ErrMsg = "invalid match type: " + str
		return errors.ErrInvalid
	}
	return nil
}

const (
	// RuleMatchAll requires all conditions must be matched
	RuleMatchAll RuleConditionMatchType = 1
	//RuleMatchAny allows any condition to match
	RuleMatchAny RuleConditionMatchType = 2
)

type RuleTrigger string

const (
	RuleTriggerCreate RuleTrigger = "document-create"
	RuleTriggerUpdate RuleTrigger = "document-update"
)

var AllRuleTriggers = RuleTriggerArray{RuleTriggerCreate, RuleTriggerUpdate}

type RuleTriggerArray []RuleTrigger

func (t *RuleTriggerArray) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *RuleTriggerArray) ToJsonString() string {
	val, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	return string(val)
}

func (t *RuleTriggerArray) Scan(src interface{}) error {
	array, ok := src.([]byte)
	if !ok {
		return errors.New("invalid type, expected []byte")
	}
	err := json.Unmarshal(array, t)
	return err
}

type Rule struct {
	Id          int                    `db:"id"`
	UserId      int                    `db:"user_id"`
	Name        string                 `db:"name"`
	Description string                 `db:"description"`
	Enabled     bool                   `db:"enabled"`
	Order       int                    `db:"rule_order"`
	Mode        RuleConditionMatchType `db:"mode"`
	Timestamp
	Triggers RuleTriggerArray `db:"triggers"`

	Conditions []*RuleCondition
	Actions    []*RuleAction
}

func (r *Rule) Validate() error {
	for i, v := range r.Conditions {
		err := v.Validate()
		if err != nil {
			if isErr, ok := err.(errors.Error); ok {
				isErr.ErrMsg = fmt.Sprintf("condition %d: %s", i+1, isErr.ErrMsg)
				return isErr
			} else {
				err = fmt.Errorf("condition %d: %v", i, err)
			}
			return err
		}
	}

	if len(r.Actions) == 0 {
		return errors.ErrInvalid
	}
	return nil
}

type RuleConditionType string

func (r RuleConditionType) String() string {
	return string(r)
}

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

var AllConditionTypes = []RuleConditionType{
	RuleConditionNameIs,
	RuleConditionNameStarts,
	RuleConditionNameContains,

	RuleConditionDescriptionIs,
	RuleConditionDescriptionStarts,
	RuleConditionDescriptionContains,

	RuleConditionContentIs,
	RuleConditionContentStarts,
	RuleConditionContentContains,

	RuleConditionDateIs,
	RuleConditionDateAfter,
	RuleConditionDateBefore,

	RuleConditionMetadataHasKey,
	RuleConditionMetadataHasKeyValue,
	RuleConditionMetadataCount,
	RuleConditionMetadataCountLessThan,
	RuleConditionMetadataCountMoreThan,
}

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
	Value   string `db:"value"`
	DateFmt string `db:"date_fmt"`

	// Metadata to operate with
	MetadataKey       IntId `db:"metadata_key"`
	MetadataValue     IntId `db:"metadata_value"`
	MetadataKeyName   Text  `db:"metadata_key_name"`
	MetadataValueName Text  `db:"metadata_value_name"`
}

func (r *RuleCondition) Validate() error {
	err := errors.ErrInvalid

	validType := false
	for _, v := range AllConditionTypes {
		if r.ConditionType == v {
			validType = true
			break
		}
	}

	if !validType {
		err.ErrMsg = fmt.Sprintf("invalid condition type: %s", r.ConditionType)
		return err
	}

	if r.IsRegex {
		_, regexErr := regexp.Compile(r.Value)
		if regexErr != nil {
			err.ErrMsg = "invalid regex"
			err.Err = regexErr
			return err
		}
	}

	condText := r.ConditionType.String()
	if strings.Contains(condText, "name") ||
		strings.Contains(condText, "description") ||
		strings.Contains(condText, "content") {
		if r.HasMetadata() {
			err.ErrMsg = condText + " cannot match metadata"
			return err
		}
		if r.Value == "" {
			err.ErrMsg = "matching value is empty"
			return err
		}
	}

	if r.ConditionType == RuleConditionMetadataHasKey {
		if r.MetadataKey == 0 {
			err.ErrMsg = "must have metadata key defined"
			return err
		}
	}
	if r.ConditionType == RuleConditionMetadataHasKeyValue {
		if r.MetadataKey == 0 || r.MetadataValue == 0 {
			err.ErrMsg = "must have metadata key and value defined"
			return err
		}
	}

	if r.ConditionType == RuleConditionDateIs {
		if r.DateFmt == "" {
			err.ErrMsg = "date format (date_fmt) cannot be empty"
			return err
		}

		if !r.IsRegex {
			err.ErrMsg = "regex must be enabled when parsing date"
		}
	}
	return nil
}

func (r *RuleCondition) HasMetadata() bool {
	return r.MetadataKey > 0 && r.MetadataValue > 0
}

type RuleActionType string

func (r RuleActionType) String() string {
	return string(r)
}

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

	Action            RuleActionType `db:"action"`
	Value             string         `db:"value"`
	MetadataKey       IntId          `db:"metadata_key"`
	MetadataValue     IntId          `db:"metadata_value"`
	MetadataKeyName   Text           `db:"metadata_key_name"`
	MetadataValueName Text           `db:"metadata_value_name"`
}

type MetadataRuleType string

const (
	MetadataMatchExact MetadataRuleType = "exact"
	MetadataMatchRegex MetadataRuleType = "regex"
)
