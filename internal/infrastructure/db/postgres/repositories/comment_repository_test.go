package repositories_test

import (
	"context"
	"testing"
	"time"

	"ZVideo/internal/domain/comment/entity"
	"ZVideo/internal/infrastructure/db/postgres/repositories"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCommentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&entity.Comment{},
	)
	assert.NoError(t, err)

	return db
}

func TestCommentRepository_Integration(t *testing.T) {
	db := setupCommentTestDB(t)
	repo := repositories.NewCommentRepository(db)
	ctx := context.Background()

	t.Run("Create and Get Comment", func(t *testing.T) {
		comment := &entity.Comment{
			VideoID:   1,
			UserID:    1,
			Content:   "ZZZZZZ!",
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, comment)
		assert.NoError(t, err)
		assert.NotZero(t, comment.ID)

		savedComment, err := repo.GetByID(ctx, comment.ID)
		assert.NoError(t, err)
		assert.NotNil(t, savedComment)
		assert.Equal(t, comment.Content, savedComment.Content)
	})

	t.Run("Update Comment", func(t *testing.T) {
		comment := &entity.Comment{
			VideoID: 1,
			UserID:  1,
			Content: "Old content",
		}
		repo.Create(ctx, comment)

		comment.Content = "New content"
		err := repo.Update(ctx, comment)
		assert.NoError(t, err)

		updatedComment, _ := repo.GetByID(ctx, comment.ID)
		assert.Equal(t, "New content", updatedComment.Content)
	})

	t.Run("Delete Comment", func(t *testing.T) {
		comment := &entity.Comment{VideoID: 1, UserID: 1, Content: "To be deleted"}
		repo.Create(ctx, comment)

		err := repo.Delete(ctx, comment.ID)
		assert.NoError(t, err)

		deleted, err := repo.GetByID(ctx, comment.ID)
		assert.NoError(t, err)
		assert.Nil(t, deleted)
	})

	t.Run("Get Comments by Video and count", func(t *testing.T) {
		videoID := 10
		comment1 := &entity.Comment{VideoID: videoID, UserID: 1, Content: "First"}
		comment2 := &entity.Comment{VideoID: videoID, UserID: 2, Content: "Second"}
		_ = repo.Create(ctx, comment1)
		_ = repo.Create(ctx, comment2)

		comments, err := repo.GetByVideoID(ctx, videoID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, comments, 2)

		count, err := repo.GetCountByVideo(ctx, videoID)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
