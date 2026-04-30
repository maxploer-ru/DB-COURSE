package repository

import (
	"ZVideo/internal/domain"
	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/mappers"
	"ZVideo/internal/infrastructure/db/mongo/models"
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	db    *mongo.Database
	users *mongo.Collection
	roles *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		db:    db,
		users: db.Collection(mongoinfra.CollectionUsers),
		roles: db.Collection(mongoinfra.CollectionRoles),
	}
}

func (repo *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.Role == nil {
		return domain.ErrInternalServer
	}
	id, err := mongoinfra.NextID(ctx, repo.db, mongoinfra.CollectionUsers)
	if err != nil {
		return err
	}
	user.ID = id

	model := mappers.FromDomainUser(user)
	_, err = repo.users.InsertOne(ctx, model)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			idx := duplicateIndexName(err)
			if strings.Contains(idx, "username") {
				return domain.ErrUserNameAlreadyExists
			}
			if strings.Contains(idx, "email") {
				return domain.ErrUserEmailAlreadyExists
			}
		}
		return fmt.Errorf("create user failed: %w", err)
	}
	return nil
}

func (repo *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	var user models.User
	if err := repo.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id failed: %w", err)
	}

	role, err := repo.getRoleByID(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainUser(&user, role), nil
}

func (repo *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user models.User
	if err := repo.users.FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username failed: %w", err)
	}

	role, err := repo.getRoleByID(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainUser(&user, role), nil
}

func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user models.User
	if err := repo.users.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email failed: %w", err)
	}

	role, err := repo.getRoleByID(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainUser(&user, role), nil
}

func (repo *UserRepository) Update(ctx context.Context, user *domain.User) error {
	if user.Role == nil {
		return domain.ErrInternalServer
	}
	update := bson.M{
		"username":              user.Username,
		"email":                 user.Email,
		"password_hash":         user.PasswordHash,
		"is_active":             user.IsActive,
		"notifications_enabled": user.NotificationsEnabled,
		"role_id":               user.Role.ID,
	}
	res, err := repo.users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *UserRepository) Delete(ctx context.Context, id int) error {
	res, err := repo.users.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"is_active": false}})
	if err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := repo.users.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, fmt.Errorf("get user by email failed: %w", err)
	}
	return count > 0, nil
}

func (repo *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	count, err := repo.users.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, fmt.Errorf("get user by username failed: %w", err)
	}
	return count > 0, nil
}

func (repo *UserRepository) Ban(ctx context.Context, id int) error {
	res, err := repo.users.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"is_active": false}})
	if err != nil {
		return fmt.Errorf("ban user failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *UserRepository) Unban(ctx context.Context, id int) error {
	res, err := repo.users.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"is_active": true}})
	if err != nil {
		return fmt.Errorf("unban user failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *UserRepository) SetNotificationsEnabled(ctx context.Context, id int, enabled bool) error {
	res, err := repo.users.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"notifications_enabled": enabled}})
	if err != nil {
		return fmt.Errorf("update notifications enabled failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *UserRepository) getRoleByID(ctx context.Context, roleID int) (*models.Role, error) {
	var role models.Role
	if err := repo.roles.FindOne(ctx, bson.M{"_id": roleID}, options.FindOne().SetProjection(bson.M{"name": 1, "is_default": 1})).Decode(&role); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get role by id failed: %w", err)
	}
	return &role, nil
}
