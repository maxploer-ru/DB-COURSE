package repositories_test

import (
	"context"
	"testing"
	"time"

	"ZVideo/internal/domain/channel/entity"
	"ZVideo/internal/infrastructure/db/postgres/repositories"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&entity.Channel{},
	)
	assert.NoError(t, err)

	return db
}

func TestChannelRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewChannelRepository(db)
	ctx := context.Background()

	t.Run("Create and Get Channel", func(t *testing.T) {
		channel := &entity.Channel{
			UserID:      1,
			Name:        "Z Channel",
			Description: "Test Description",
			CreatedAt:   time.Now(),
		}

		err := repo.Create(ctx, channel)
		assert.NoError(t, err)
		assert.NotZero(t, channel.ID)

		savedChannel, err := repo.GetByID(ctx, channel.ID)
		assert.NoError(t, err)
		assert.NotNil(t, savedChannel)
		assert.Equal(t, channel.Name, savedChannel.Name)
	})

	t.Run("Update Channel", func(t *testing.T) {
		channel := &entity.Channel{
			UserID:      1,
			Name:        "Z Z Z Z",
			Description: "Old",
		}
		repo.Create(ctx, channel)

		channel.Description = "New"
		err := repo.Update(ctx, channel)
		assert.NoError(t, err)

		updatedChannel, _ := repo.GetByID(ctx, channel.ID)
		assert.Equal(t, "New", updatedChannel.Description)
	})

	t.Run("Delete Channel", func(t *testing.T) {
		channel := &entity.Channel{UserID: 1, Name: "Z Z Z Z"}
		repo.Create(ctx, channel)

		err := repo.Delete(ctx, channel.ID)
		assert.NoError(t, err)

		deleted, err := repo.GetByID(ctx, channel.ID)
		assert.NoError(t, err)
		assert.Nil(t, deleted)
	})
}
