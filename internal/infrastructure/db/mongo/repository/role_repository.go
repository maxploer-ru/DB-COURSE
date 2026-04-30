package repository

import (
	"ZVideo/internal/domain"
	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/mappers"
	"ZVideo/internal/infrastructure/db/mongo/models"
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RoleRepository struct {
	db    *mongo.Database
	roles *mongo.Collection
}

func NewRoleRepository(db *mongo.Database) *RoleRepository {
	return &RoleRepository{
		db:    db,
		roles: db.Collection(mongoinfra.CollectionRoles),
	}
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionRoles)
	if err != nil {
		return err
	}
	role.ID = id

	model := mappers.FromDomainRole(role)
	if _, err := r.roles.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create role failed: %w", err)
	}
	return nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id int) (*domain.Role, error) {
	var role models.Role
	if err := r.roles.FindOne(ctx, bson.M{"_id": id}).Decode(&role); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get role failed: %w", err)
	}
	return mappers.ToDomainRole(&role), nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role models.Role
	if err := r.roles.FindOne(ctx, bson.M{"name": name}).Decode(&role); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get role failed: %w", err)
	}
	return mappers.ToDomainRole(&role), nil
}

func (r *RoleRepository) GetDefaultRole(ctx context.Context) (*domain.Role, error) {
	var role models.Role
	if err := r.roles.FindOne(ctx, bson.M{"is_default": true}).Decode(&role); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get default role failed: %w", err)
	}
	return mappers.ToDomainRole(&role), nil
}

func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	update := bson.M{"$set": bson.M{"name": role.Name, "is_default": role.IsDefault}}
	res, err := r.roles.UpdateOne(ctx, bson.M{"_id": role.ID}, update)
	if err != nil {
		return fmt.Errorf("update role failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id int) error {
	res, err := r.roles.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete role failed: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
