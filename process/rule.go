package process

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type DocumentRule struct {
	Rule     *models.Rule
	Document *models.Document
}

func NewDocumentRule(document *models.Document, rule *models.Rule) DocumentRule {
	return DocumentRule{
		Rule:     rule,
		Document: document,
	}
}

func (d *DocumentRule) Match() (bool, error) {
	hasMatch := false

	logrus.Debugf("(documentRule) match document: %s, rule: %d", d.Document.Id, d.Rule.Id)
	for _, condition := range d.Rule.Conditions {

		logrus.Debugf("(documentRule) match condition: %d", condition.Id)
		condText := string(condition.ConditionType)
		fmt.Printf("match %s\n", condition.ConditionType)

		var ok = false
		var err error
		if strings.HasPrefix(condText, "name") {
			ok, err = d.matchText(condition, d.Document.Name)
		} else if strings.HasPrefix(condText, "description") {
			ok, err = d.matchText(condition, d.Document.Description)
		} else if strings.HasPrefix(condText, "content") {
			ok, err = d.matchText(condition, d.Document.Content)
		} else if strings.HasPrefix(condText, "metadata_has_key") {
			ok = d.hasMetadataKey(condition)
		} else if condition.ConditionType == models.RuleConditionMetadataHasKey {
			ok = d.hasMetadataKey(condition)
		} else if condition.ConditionType == models.RuleConditionMetadataHasKeyValue {
			ok = d.hasMetadataKeyValue(condition)
		} else {
			err := errors.ErrInternalError
			err.ErrMsg = "unknown condition type: " + condText
			return false, err
		}
		if err != nil {
			return false, fmt.Errorf("evaluate condition: %v", err)
		}

		if ok {
			if condition.Inverted {
				println("condition (inverted) ok")
				hasMatch = true
			} else {
				println("condition ok")
				hasMatch = true
			}

		} else {
			if d.Rule.Mode == models.RuleMatchAll && !condition.Inverted {
				println("condition failed, skip")
				return false, nil
			} else {
				println("no match")

			}
		}

		if hasMatch && d.Rule.Mode == models.RuleMatchAny {
			println("condition passed any, skip rest")
			// already found a match, skip rest of the conditions
			break
		}
	}
	return hasMatch, nil
}

func (d *DocumentRule) matchText(condition *models.RuleCondition, text string) (bool, error) {
	value := condition.Value
	if condition.CaseInsensitive {
		text = strings.ToLower(text)
		value = strings.ToLower(value)
	}

	switch condition.ConditionType {
	case models.RuleConditionNameIs, models.RuleConditionDescriptionIs, models.RuleConditionContentIs:
		return text == value, nil
	case models.RuleConditionNameStarts, models.RuleConditionDescriptionStarts, models.RuleConditionContentStarts:
		return strings.HasPrefix(text, value), nil
	case models.RuleConditionNameContains, models.RuleConditionDescriptionContains, models.RuleConditionContentContains:
		return strings.Contains(text, value), nil
	default:
		err := errors.ErrInternalError
		err.ErrMsg = fmt.Sprintf("unknown condition type: %s", condition.ConditionType)
		err.SetStack()
		return false, err
	}
}

func (d *DocumentRule) hasMetadataKey(condition *models.RuleCondition) bool {
	for _, v := range d.Document.Metadata {
		if v.KeyId == int(condition.MetadataKey.Int64) {
			return true
		}
	}
	return false
}

func (d *DocumentRule) hasMetadataKeyValue(condition *models.RuleCondition) bool {
	for _, v := range d.Document.Metadata {
		if v.KeyId == int(condition.MetadataKey.Int64) && v.ValueId == int(condition.MetadataValue.Int64) {
			return true
		}
	}
	return false
}

func (d *DocumentRule) hasMetadataCount(condition *models.RuleCondition) (bool, error) {
	limit, err := strconv.Atoi(condition.Value)
	if err != nil || limit < 0 {
		e := errors.ErrInvalid
		e.ErrMsg = "value must be a non-negative number"
		return false, e
	}

	switch condition.ConditionType {
	case models.RuleConditionMetadataCount:
		return len(d.Document.Metadata) == limit, nil
	case models.RuleConditionMetadataCountLessThan:
		return len(d.Document.Metadata) < limit, nil
	case models.RuleConditionMetadataCountMoreThan:
		return len(d.Document.Metadata) > limit, nil
	default:
		return false, fmt.Errorf("not metadata count condition: %v", condition.ConditionType)
	}
}

func (d *DocumentRule) RunActions() error {
	logrus.Debugf("(documentRule) run actions, document: %s, rule: %d", d.Document.Id, d.Rule.Id)

	var err error
	var actionError error

	for _, action := range d.Rule.Actions {
		logrus.Debugf("(documentRule) run action: %d", action.Id)
		switch action.Action {
		case models.RuleActionSetName:
			actionError = d.setName(action)
		case models.RuleActionAppendName:
			actionError = d.appendName(action)
		case models.RuleActionAddMetadata:
			actionError = d.addMetadata(action)
		default:
			e := errors.ErrInternalError
			e.ErrMsg = fmt.Sprintf("unknown action type: %v", action.Action)
			actionError = e
		}

		if actionError != nil {
			err = fmt.Errorf("action (%d): %v", action.Id, actionError)
			actionError = nil
		}
	}
	return err
}

func (d *DocumentRule) setName(action *models.RuleAction) error {
	d.Document.Name = action.Value
	return nil
}

func (d *DocumentRule) appendName(action *models.RuleAction) error {
	d.Document.Name += action.Value
	return nil
}

func (d *DocumentRule) addMetadata(action *models.RuleAction) error {
	if len(d.Document.Metadata) == 0 {
		d.Document.Metadata = []models.Metadata{{
			KeyId:   int(action.MetadataKey.Int64),
			ValueId: int(action.MetadataValue.Int64),
		}}
		return nil
	}

	// check if key-value already exists
	for _, v := range d.Document.Metadata {
		if v.KeyId == int(action.MetadataKey.Int64) && v.ValueId == int(action.MetadataValue.Int64) {
			return nil
		}
	}

	d.Document.Metadata = append(d.Document.Metadata, models.Metadata{
		KeyId:   int(action.MetadataKey.Int64),
		ValueId: int(action.MetadataValue.Int64),
	})
	return nil
}
