package repositories_test

import (
	"context"
	"testing"
	"time"

	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/infrastructure/db/postgres/repositories"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVideoTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&entity.Video{},
	)
	assert.NoError(t, err)

	return db
}

func TestVideoRepository_Integration(t *testing.T) {
	db := setupVideoTestDB(t)
	repo := repositories.NewVideoRepository(db)
	ctx := context.Background()

	t.Run("Create and Get Video", func(t *testing.T) {
		video := &entity.Video{
			ChannelID:   1,
			Title:       "Test Video",
			Description: "Desc",
			Filepath:    "/path/to/video.mp4",
			CreatedAt:   time.Now(),
		}

		err := repo.Create(ctx, video)
		assert.NoError(t, err)
		assert.NotZero(t, video.ID)

		savedVideo, err := repo.GetByID(ctx, video.ID)
		assert.NoError(t, err)
		assert.NotNil(t, savedVideo)
		assert.Equal(t, video.Title, savedVideo.Title)
	})

	t.Run("Update Video", func(t *testing.T) {
		video := &entity.Video{
			ChannelID:   1,
			Title:       "Old Title",
			Description: "Desc",
			Filepath:    "/path",
		}
		repo.Create(ctx, video)

		video.Title = "New Title"
		err := repo.Update(ctx, video)
		assert.NoError(t, err)

		updatedVideo, _ := repo.GetByID(ctx, video.ID)
		assert.Equal(t, "New Title", updatedVideo.Title)
	})

	t.Run("Delete Video", func(t *testing.T) {
		video := &entity.Video{ChannelID: 1, Title: "Delete Me"}
		repo.Create(ctx, video)

		err := repo.Delete(ctx, video.ID)
		assert.NoError(t, err)

		deleted, err := repo.GetByID(ctx, video.ID)
		assert.NoError(t, err)
		assert.Nil(t, deleted)
	})
}
