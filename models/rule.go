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

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type RuleType string

const (
	// RegexRule matches filter with regex
	RegexRule RuleType = "regex"
	// Exact must match exactly.
	ExactRule RuleType = "exact"
)

// RuleAction defines what to do when rule has fired.
// RuleActions are not exclusive, meaning multiple actions can be defined for single rule.
// E.g. add metadata and tag + set description
type RuleAction uint16

func (r RuleAction) AddMetadata() bool {
	return r&RuleActionAddMetadata != 0
}

func (r RuleAction) Rename() bool {
	return r&RuleActionRename != 0
}

func (r RuleAction) Date() bool {
	return r&RuleActionSetDate != 0
}

func (r RuleAction) Tag() bool {
	return r&RuleActionAddTag != 0
}

func (r RuleAction) Description() bool {
	return r&RuleActionSetDescription != 0
}

const (
	RuleActionAddMetadata    RuleAction = 1 << 0
	RuleActionRename         RuleAction = 1 << 1
	RuleActionSetDate        RuleAction = 1 << 2
	RuleActionAddTag         RuleAction = 1 << 3
	RuleActionSetDescription RuleAction = 1 << 4
)

// RuleActionConfig defines action to perform when rule has fired.
// Fields and their content depends on Action, and specific key only applies if
// it's defined in Action.
// RuleActionConfig is serialized as json in database.
type RuleActionConfig struct {
	Action          RuleAction
	MetadataKeyId   int
	MetadataValueId int
	Tag             int
	// DateFmt is format to try to parse time with
	DateFmt string
	// DateSeparator in format. This is to try to ensure and in some cases fix minor errors
	// in invalid formats, e.g. 2020-5-01 -> 2020-05-01.
	DateSeparator string
	Description   string
}

func (r *RuleActionConfig) Scan(src interface{}) error {
	buf := []byte{}
	if b, ok := src.([]byte); ok {
		buf = b
	} else if str, ok := src.(string); ok {
		buf = []byte(str)
	}
	err := json.Unmarshal(buf, r)
	return err
}

func (r *RuleActionConfig) Value() (driver.Value, error) {
	buf, err := json.Marshal(r)
	return string(buf), err
}

// Validate ensures rule configuration is valid. If valid, return nil, else return
// exact reason for why rule is invalid.
func (r *Rule) Validate() error {
	if r.Filter == "" {
		return errors.New("filter cannot be empty")
	}
	if r.Action.MetadataKeyId != 0 && r.Action.MetadataValueId != 0 {
		r.Action.Action |= RuleActionAddMetadata
	}
	if r.Action.Tag != 0 {
		r.Action.Action |= RuleActionAddTag
	}
	if r.Action.DateFmt != "" {
		r.Action.Action |= RuleActionSetDate
	}
	if r.Action.Description != "" {
		r.Action.Action |= RuleActionSetDescription
	}

	if r.Action.Action == 0 {
		return errors.New("no action set")
	}
	return nil
}

// Rule defines single rule, which has filter (either exact or regex) and action to perform when filter
// fires.
type Rule struct {
	Timestamp
	Id      int              `db:"id"`
	Userid  int              `db:"user_id"`
	Type    RuleType         `db:"rule_type" json:"rule_type"`
	Filter  string           `db:"filter"`
	Comment string           `db:"comment"`
	Action  RuleActionConfig `db:"action"`
	Active  bool             `db:"active"`
}
