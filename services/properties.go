package services

import (
	"context"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services/process"
	"tryffel.net/go/virtualpaper/storage"
)

type PropertyService struct {
	db      *storage.Database
	process *process.Manager
}

func NewPropertyService(db *storage.Database, manager *process.Manager) *PropertyService {
	return &PropertyService{
		db:      db,
		process: manager,
	}
}

func (s *PropertyService) GetProperties(ctx context.Context, userId int, page storage.Paging, sort storage.SortKey) (*[]models.Property, int, error) {
	props, err := s.db.PropertyStore.GetProperties(s.db, userId, page, sort)
	if err != nil {
		return props, 0, err
	}
	total, err := s.db.PropertyStore.GetTotalProperties(s.db, userId)
	return props, total, err
}

func (s *PropertyService) GetProperty(ctx context.Context, id int) (*models.Property, error) {
	return s.db.PropertyStore.GetProperty(s.db, id)
}

func (s *PropertyService) UserOwnsProperty(ctx context.Context, userId, id int) (bool, error) {
	return s.db.PropertyStore.UserOwnsProperty(s.db, userId, id)
}

func (s *PropertyService) AddProperty(ctx context.Context, user *models.User, property *models.Property) error {
	validationError := errors.ErrInvalid

	if !user.IsAdmin && property.Global {
		validationError.ErrMsg = "user not admin"
		return validationError
	}

	err := s.db.PropertyStore.AddProperty(s.db, property)
	return err
}

func (s *PropertyService) UpdateProperty(ctx context.Context, property *models.Property) error {
	err := s.db.PropertyStore.UpdateProperty(s.db, property)
	return err
}
