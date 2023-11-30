package services

import (
	"context"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services/process"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
	"tryffel.net/go/virtualpaper/util/logger"
)

type RuleService struct {
	db      *storage.Database
	search  *search.Engine
	process *process.Manager
}

func NewRuleService(db *storage.Database, search *search.Engine, manager *process.Manager) *RuleService {
	return &RuleService{
		db:      db,
		search:  search,
		process: manager,
	}
}

func (service *RuleService) UserOwnsRule(ctx context.Context, userId, ruleId int) (bool, error) {
	return service.db.RuleStore.UserOwnsRule(userId, ruleId)
}

func (service *RuleService) GetRules(ctx context.Context, userId int, page storage.Paging) ([]*models.Rule, int, error) {
	return service.db.RuleStore.GetUserRules(userId, page)
}

func (service *RuleService) Get(ctx context.Context, userId int, id int) (*models.Rule, error) {
	return service.db.RuleStore.GetUserRule(userId, id)
}

func (service *RuleService) Create(ctx context.Context, rule *models.Rule) (*models.Rule, error) {
	err := service.db.RuleStore.AddRule(rule.UserId, rule)
	if err != nil {
		return nil, err
	}

	createdRule, err := service.db.RuleStore.GetUserRule(rule.UserId, rule.Id)
	if err != nil {
		return nil, err
	}
	return createdRule, nil
}

func (service *RuleService) Update(ctx context.Context, rule *models.Rule) error {
	return service.db.RuleStore.UpdateRule(rule.UserId, rule)
}

func (service *RuleService) Delete(ctx context.Context, ruleId int) error {
	return service.db.RuleStore.DeleteRule(ruleId)
}

func (service *RuleService) TestRule(ctx context.Context, userId int, ruleId int, docId string) (*process.RuleTestResult, error) {
	rule, err := service.db.RuleStore.GetUserRule(userId, ruleId)
	if err != nil {
		return nil, err
	}

	doc, err := service.db.DocumentStore.GetDocument(docId)
	if err != nil {
		return nil, err
	}

	if doc.UserId != userId {
		return nil, errors.ErrRecordNotFound
	}

	metadata, err := service.db.MetadataStore.GetDocumentMetadata(userId, docId)
	if err != nil {
		return nil, err
	}
	doc.Metadata = *metadata
	logger.Context(ctx).WithField("user", userId).WithField("documentId", doc.Id).WithField("rule", ruleId).Info("Test rule")
	processRule := process.NewDocumentRule(doc, rule)
	status := processRule.MatchTest()

	logger.Context(ctx).WithField("documentId", doc.Id).WithField("rule", ruleId).Infof("Test rule finished: %v", status.Match)
	return status, nil
}

func (service *RuleService) Reorder(ctx context.Context, userId int, ruleIds []int) error {
	return service.db.RuleStore.ReorderRules(userId, ruleIds)
}
