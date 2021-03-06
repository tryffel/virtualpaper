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
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

var reRegexHasSubMatch = regexp.MustCompile("\\(.+\\)")

func runRules(document *models.Document, rules *[]models.Rule) error {
	for _, rule := range *rules {
		logrus.Debugf("run rule %d against document %s", rule.Id, document.Id)
		match, err := documentMatchesFilter(document, rule.Type, rule.Filter)
		if err != nil {
			logrus.Debugf("automatic rule, filter error: %v", err)
			continue
		}

		if match != "" {
			err = applyRule(document, rule, match)
			if err != nil {
				logrus.Debugf("failed to apply document rule: %v", err)
			}
		}
	}
	return nil
}

func matchMetadata(document *models.Document, values *[]models.MetadataValue) error {
	logrus.Debugf("match metadata keys for doc: %s, total of %d rules", document.Id, len(*values))
	for _, v := range *values {
		match, err := documentMatchesFilter(document, v.MatchType, v.MatchFilter)
		if err != nil {
			logrus.Debugf("automatic metadata rule, filter error: %v", err)
			continue
		}
		if match != "" {
			addMetadataToDocument(document, v.KeyId, v.Id)
		}
	}
	return nil
}

func documentMatchesFilter(document *models.Document, ruleType models.RuleType, filter string) (string, error) {
	if ruleType == models.ExactRule {

		lowerContent := strings.ToLower(document.Content)
		lowerRule := strings.ToLower(filter)
		contains := strings.Contains(lowerContent, lowerRule)
		if contains {
			return lowerRule, nil
		} else {
			return "", nil
		}
	} else if ruleType == models.RegexRule {
		// if regex captures submatch, return first submatch (not the match itself),
		// else return regex match

		re, err := regexp.Compile(filter)
		if err != nil {
			return "", fmt.Errorf("invalid regex: %v", err)
		}

		if reRegexHasSubMatch.MatchString(filter) {
			matches := re.FindStringSubmatch(document.Content)
			if len(matches) == 0 {
				return "", nil
			}
			if len(matches) == 1 {
				return "", nil
			}

			if len(matches) == 2 {
				return matches[1], nil
			} else {
				logrus.Debugf("more than 1 regex matches, pick first. regex: %s doc. %s, matches: %v",
					filter, document.Id, matches)
				return matches[1], nil
			}
		} else {
			match := re.FindString(filter)
			return match, nil
		}
	} else {
		return "", fmt.Errorf("unknown rule type: %s", ruleType)
	}
}

func applyRule(document *models.Document, rule models.Rule, match string) error {
	logMsg := fmt.Sprintf("(automatic rule) doc: %s: ", document.Id)

	if rule.Action.Action.AddMetadata() {
		addMetadataToDocument(document, rule.Action.MetadataKeyId, rule.Action.MetadataValueId)
		logMsg += fmt.Sprintf("add metadata (key %d, value %d)",
			rule.Action.MetadataKeyId, rule.Action.MetadataValueId)
	}
	if rule.Action.Action.Rename() {

		document.Name = match
		logMsg += "rename document"
	}
	if rule.Action.Action.Date() {
		dateMatch := match
		if rule.Action.DateSeparator != "" {
			splits := strings.Split(match, rule.Action.DateSeparator)
			newMatch := ""
			for i, v := range splits {
				if i > 0 {
					newMatch += rule.Action.DateSeparator
				}
				if len(v) == 1 {
					newMatch += "0" + v
				} else {
					newMatch += v
				}
			}
			dateMatch = newMatch
		}

		ts, err := time.Parse(rule.Action.DateFmt, dateMatch)
		if err != nil {
			return fmt.Errorf("date format '%s' does not match string '%s'", rule.Action.DateFmt, match)
		}
		logMsg += "set date"

		document.Date = ts
	}
	if rule.Action.Action.Tag() {
		if document.Tags == nil {
			document.Tags = []models.Tag{}
		}
		tag := models.Tag{
			Id: rule.Action.Tag,
		}
		document.Tags = append(document.Tags, tag)
		logMsg += "add tag"
	}
	if rule.Action.Action.Description() {
		addDescription := rule.Action.Description
		if rule.Type == models.RegexRule {
			if addDescription == "" {
				addDescription = match
			} else {
				addDescription += ": " + match
			}
		}
		if document.Description == "" {
			document.Description = addDescription
		} else {
			document.Description = strings.Join([]string{document.Description, addDescription}, "\n\n")
		}
		logMsg += "set description"
	}
	logrus.Debug(logMsg)
	return nil
}

// add Metadata key-value to document. Make sure document does not already have given
// key-value pair before adding one.
func addMetadataToDocument(doc *models.Document, keyId, valueId int) {
	if len(doc.Metadata) == 0 {
		doc.Metadata = []models.Metadata{{
			KeyId:   keyId,
			ValueId: valueId,
		}}
		return
	}

	for _, v := range doc.Metadata {
		if v.KeyId == keyId && v.ValueId == valueId {
			return
		}
	}

	doc.Metadata = append(doc.Metadata, models.Metadata{KeyId: keyId, ValueId: valueId})
}
