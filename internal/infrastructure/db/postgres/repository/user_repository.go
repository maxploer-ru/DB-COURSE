package repository

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/mappers"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"context"
	"errors"
	"fmt"
	"log"

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

func (repo *UserRepository) Create(ctx context.Context, user *domain.User) error {
	model := mappers.FromDomainUser(user)

	if err := repo.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	user.ID = model.ID
	return nil
}

func (repo *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	var model models.User

	err := repo.db.WithContext(ctx).
		Preload("Role").
		First(&model, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id failed: %w", err)
	}

	return mappers.ToDomainUser(&model), nil
}

func (repo *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var model models.User

	err := repo.db.WithContext(ctx).
		Preload("Role").
		Where("username = ?", username).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username failed: %w", err)
	}

	return mappers.ToDomainUser(&model), nil
}

func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model models.User

	err := repo.db.WithContext(ctx).
		Preload("Role").
		Where("email = ?", email).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email failed: %w", err)
	}
	log.Print(model)
	return mappers.ToDomainUser(&model), nil
}

func (repo *UserRepository) Update(ctx context.Context, user *domain.User) error {
	model := mappers.FromDomainUser(user)

	if err := repo.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}

	user.ID = model.ID
	return nil
}

func (repo *UserRepository) Delete(ctx context.Context, id int) error {
	return repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Set("is_active", false).Error
}

func (repo *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64

	err := repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("get user by email failed: %w", err)
	}

	return count > 0, nil
}

func (repo *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64

	err := repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("username = ?", username).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("get user by username failed: %w", err)
	}

	return count > 0, nil
}

func (repo *UserRepository) Ban(ctx context.Context, id int) error {
	result := repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("ban user failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *UserRepository) Unban(ctx context.Context, id int) error {
	result := repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("is_active", true)
	if result.Error != nil {
		return fmt.Errorf("unban user failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *UserRepository) SetNotificationsEnabled(ctx context.Context, id int, enabled bool) error {
	result := repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("notifications_enabled", enabled)
	if result.Error != nil {
		return fmt.Errorf("update notifications enabled failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
