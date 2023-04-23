package process

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

// run rule in test mode, logging all results
func (d *DocumentRule) MatchTest() *RuleTestResult {
	logBuf := &bytes.Buffer{}

	logger := logrus.New()
	logger.SetOutput(logBuf)
	logger.SetFormatter(&formatter{})
	format, ok := logger.Formatter.(*logrus.TextFormatter)
	if ok {
		format.FullTimestamp = false
		format.DisableTimestamp = true

	}

	hasMatch := false

	result := &RuleTestResult{
		StartedAt:       int(time.Now().UnixNano() / 1000000),
		RuleId:          d.Rule.Id,
		Conditions:      make([]RuleTestConditionResult, len(d.Rule.Conditions)),
		Actions:         make([]RuleTestAction, len(d.Rule.Actions)),
		ConditionOutput: [][]string{},
		ActionOutput:    [][]string{},
	}

	for i, v := range d.Rule.Conditions {
		result.Conditions[i].ConditionId = v.Id
		result.Conditions[i].ConditionType = v.ConditionType.String()
		result.Conditions[i].Skipped = true
	}
	for i, v := range d.Rule.Actions {
		result.Actions[i].Skipped = !v.Enabled
		result.Actions[i].ActionId = v.Id
		result.Actions[i].ActionType = v.Action.String()
	}

	var conditionResult []string
	logConditionOut := func(format string, args ...interface{}) {
		out := fmt.Sprintf(format, args...)
		conditionResult = append(conditionResult, out)
	}

	logger.Infof("Try to match document: %s with rule: '%s' (id: %d)", d.Document.Id, d.Rule.Name, d.Rule.Id)
	for i, condition := range d.Rule.Conditions {
		result.Conditions[i].ConditionId = condition.Id
		doBreak := false
		if !condition.Enabled {
			logger.Warnf("condition: %d (id:%d), %s is disabled, skipping condition", condition.Id, i+1, condition.ConditionType)
			logConditionOut("condition disabled")
			continue
		} else {
			result.Conditions[i].Skipped = false
			logger.Infof("evaluate condition %d (id:%d), type: '%s'", condition.Id, i+1, condition.ConditionType)
			condText := string(condition.ConditionType)
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
			} else if strings.HasPrefix(condText, "date") {
				ok, err = d.extractDates(condition, time.Now(), logger)

				if ok {
					y, m, d := d.date.Date()
					logger.Infof("found date %d-%d-%d", y, m, d)
					logConditionOut("found date %d-%d-%d", y, m, d)
				}

			} else if strings.HasPrefix(condText, "metadata_count") {
				ok, err = d.hasMetadataCount(condition)
			} else if condition.ConditionType == models.RuleConditionMetadataHasKey {
				ok = d.hasMetadataKey(condition)
			} else if condition.ConditionType == models.RuleConditionMetadataHasKeyValue {
				ok = d.hasMetadataKeyValue(condition)
			} else {
				err := errors.ErrInternalError
				err.ErrMsg = "unknown condition type: " + condText
				result.Error = err.Error()
				break
			}
			if err != nil {
				e := errors.ErrInternalError
				e.ErrMsg = fmt.Errorf("evaluate condition: %v", err).Error()
				result.Error = e.Error()
				break
			}

			if condition.Inverted {
				logConditionOut("invert condition matched: %t -> %t", ok, !ok)
				ok = !ok
			}

			if ok {
				hasMatch = true
				logger.Infof("condition %d (id %d) matched", i+1, condition.Id)
				logConditionOut("condition matched")
				result.Conditions[i].Matched = true
				if d.Rule.Mode == models.RuleMatchAny {
					// already found a match, skip rest of the conditions
					logger.Infof("document matches and mode is set to 'match any', skip rest conditions")
					logConditionOut("rule mode is set to 'match any', skip rest conditions")
					doBreak = true
				}

			} else if d.Rule.Mode == models.RuleMatchAll {
				logger.Infof("condition %d didn't match, skip rest", condition.Id)
				logConditionOut("condition didn't match")
				logConditionOut("rule mode is set to 'match all', stopping execution")
				hasMatch = false
				doBreak = true
			} else {
				logger.Infof("condition %d didn't match, continuing", condition.Id)
				logConditionOut("condition didn't match")
			}
		}

		result.ConditionOutput = append(result.ConditionOutput, conditionResult)
		conditionResult = []string{}
		if doBreak {
			break
		}
	}

	if hasMatch {
		var actionResult []string
		logActionOut := func(format string, args ...interface{}) {
			out := fmt.Sprintf(format, args...)
			actionResult = append(actionResult, out)
		}

		for _, action := range d.Rule.Actions {
			err := d.runAction(action, logActionOut)
			if err != nil {
				break
			}
			result.ActionOutput = append(result.ActionOutput, actionResult)
			actionResult = []string{}
		}
	}

	result.StoppedAt = int(time.Now().UnixNano() / 1000000)
	result.TookMs = result.StoppedAt - result.StartedAt
	result.Match = hasMatch

	result.Log = logBuf.String()
	return result
}
