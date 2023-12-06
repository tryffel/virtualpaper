package services

import (
	"context"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
)

type UserService struct {
	db     *storage.Database
	search *search.Engine
}

func NewUserServices(db *storage.Database, search *search.Engine) *UserService {
	return &UserService{
		db:     db,
		search: search,
	}
}

func (service *UserService) GetPreferences(ctx context.Context, userId int) (*models.UserPreferences, error) {
	preferences, err := service.db.UserStore.GetUserPreferences(userId)
	if err != nil {
		return nil, err
	}
	user, err := service.db.UserStore.GetUser(userId)
	if err != nil {
		return nil, err
	}

	preferences.CreatedAt = user.CreatedAt
	preferences.UpdatedAt = user.UpdatedAt
	preferences.Email = user.Email
	return preferences, nil
}

func (service *UserService) UpdatePreferences(ctx context.Context, preferences *models.UserPreferences) error {
	user, err := service.db.UserStore.GetUser(preferences.UserId)
	if err != nil {
		return err
	}
	attributeChanged := false
	if preferences.Email != "" {
		user.Email = preferences.Email
		attributeChanged = true
	}

	if attributeChanged {
		user.Update()
		err = service.db.UserStore.Update(user)
		if err != nil {
			return err
		}
	}
	if !attributeChanged {
		return errors.ErrAlreadyExists
	}
	return nil
}

func (service *UserService) GetUsers(ctx context.Context) (*[]aggregates.User, error) {
	raw, err := service.db.UserStore.GetUsers()
	if err != nil {
		return nil, err
	}
	users := make([]aggregates.User, len(*raw))
	for i, v := range *raw {
		users[i] = aggregates.User{
			Id:    v.Id,
			Name:  v.Name,
			Email: v.Email,
		}
	}
	return &users, nil
}
