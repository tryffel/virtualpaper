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

func (service *RuleService) GetRules(ctx context.Context, userId int, page storage.Paging, query string, enabled string) ([]*models.Rule, int, error) {
	tx, err := storage.NewTx(service.db, ctx)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Close()
	return service.db.RuleStore.GetUserRules(tx, userId, page, query, enabled)
}

func (service *RuleService) Get(ctx context.Context, userId int, id int) (*models.Rule, error) {
	return service.db.RuleStore.GetUserRule(userId, id)
}

func (service *RuleService) Create(ctx context.Context, rule *models.Rule) (*models.Rule, error) {
	tx, err := storage.NewTx(service.db, ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Close()

	err = service.db.RuleStore.AddRule(tx, rule.UserId, rule)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	createdRule, err := service.db.RuleStore.GetUserRule(rule.UserId, rule.Id)
	return createdRule, err
}

func (service *RuleService) Update(ctx context.Context, rule *models.Rule) error {
	tx, err := storage.NewTx(service.db, ctx)
	if err != nil {
		return err
	}
	defer tx.Close()
	err = service.db.RuleStore.UpdateRule(tx, rule.UserId, rule)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (service *RuleService) Delete(ctx context.Context, ruleId int) error {
	return service.db.RuleStore.DeleteRule(ruleId)
}

func (service *RuleService) TestRule(ctx context.Context, userId int, ruleId int, docId string) (*process.RuleTestResult, error) {
	rule, err := service.db.RuleStore.GetUserRule(userId, ruleId)
	if err != nil {
		return nil, err
	}

	doc, err := service.db.DocumentStore.GetDocument(service.db, docId)
	if err != nil {
		return nil, err
	}

	if doc.UserId != userId {
		return nil, errors.ErrRecordNotFound
	}

	metadata, err := service.db.MetadataStore.GetDocumentMetadata(service.db, userId, docId)
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
