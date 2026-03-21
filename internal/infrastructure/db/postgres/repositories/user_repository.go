package repositories

import (
	"ZVideo/internal/domain/entity"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) Create(ctx context.Context, user *entity.User) error {
	model := mappers.FromDomainUser(user)

	if err := repo.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt

	return nil
}

func (repo *UserRepository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	var model models.User

	err := repo.db.WithContext(ctx).
		Preload("Role").
		First(&model, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mappers.ToDomainUser(&model), nil
}

func (repo *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var model models.User

	err := repo.db.WithContext(ctx).
		Preload("Role").
		Where("username = ?", username).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mappers.ToDomainUser(&model), nil
}

func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model models.User

	err := repo.db.WithContext(ctx).
		Preload("Role").
		Where("email = ?", email).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
	}

	return mappers.ToDomainUser(&model), nil
}

func (repo *UserRepository) Update(ctx context.Context, user *entity.User) error {
	model := mappers.FromDomainUser(user)

	if err := repo.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}

	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt

	return nil
}

func (repo *UserRepository) Delete(ctx context.Context, id int) error {
	return repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Set("is_active", false).Error
}
