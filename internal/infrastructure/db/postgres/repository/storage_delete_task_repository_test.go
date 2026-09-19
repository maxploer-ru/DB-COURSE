package repository_test

import (
	"ZVideo/internal/infrastructure/db/postgres/models"
	"ZVideo/internal/infrastructure/db/postgres/repository"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type StorageDeleteTaskRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	tx   *gorm.DB
	repo *repository.StorageDeleteTaskRepository
}

func (s *StorageDeleteTaskRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *StorageDeleteTaskRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewStorageDeleteTaskRepository(s.tx)
}

func (s *StorageDeleteTaskRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *StorageDeleteTaskRepositoryTestSuite) TestClaimBatch_Positive_StateTransition() {
	task := &models.DeleteS3Task{
		Filepath:      "videos/to-delete.mp4",
		NextAttemptAt: time.Now().UTC().Add(-time.Minute),
		CreatedAt:     time.Now().UTC(),
	}
	s.NoError(s.tx.Create(task).Error)

	claimed, err := s.repo.ClaimBatch(context.Background(), 10)

	s.NoError(err)
	s.Len(claimed, 1)
	s.Equal(task.ID, claimed[0].ID)

	var locked models.DeleteS3Task
	s.NoError(s.tx.First(&locked, task.ID).Error)
	s.NotNil(locked.LockedAt)
}

func (s *StorageDeleteTaskRepositoryTestSuite) TestClaimBatch_Negative_EmptyLimit_BoundaryValueAnalysis() {
	claimed, err := s.repo.ClaimBatch(context.Background(), 0)

	s.NoError(err)
	s.Empty(claimed)
}

func (s *StorageDeleteTaskRepositoryTestSuite) TestClaimBatch_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	claimed, err := s.repo.ClaimBatch(ctx, 10)

	s.Error(err)
	s.Nil(claimed)
}

func (s *StorageDeleteTaskRepositoryTestSuite) TestMarkDone_Positive_StateTransition() {
	task := &models.DeleteS3Task{
		Filepath:      "videos/done.mp4",
		NextAttemptAt: time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
	}
	s.NoError(s.tx.Create(task).Error)

	err := s.repo.MarkDone(context.Background(), task.ID)

	s.NoError(err)
	var saved models.DeleteS3Task
	s.NoError(s.tx.First(&saved, task.ID).Error)
	s.True(saved.IsDone)
	s.NotNil(saved.ProcessedAt)
}

func (s *StorageDeleteTaskRepositoryTestSuite) TestMarkDone_Negative_NotFound_EquivalencePartitioning() {
	err := s.repo.MarkDone(context.Background(), 99999)

	s.Error(err)
}

func (s *StorageDeleteTaskRepositoryTestSuite) TestMarkFailed_Positive_StateTransition() {
	task := &models.DeleteS3Task{
		Filepath:      "videos/retry.mp4",
		NextAttemptAt: time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
	}
	s.NoError(s.tx.Create(task).Error)
	nextAttempt := time.Now().UTC().Add(time.Minute)

	err := s.repo.MarkFailed(context.Background(), task.ID, nextAttempt, "storage unavailable")

	s.NoError(err)
	var saved models.DeleteS3Task
	s.NoError(s.tx.First(&saved, task.ID).Error)
	s.Equal(1, saved.Attempts)
	s.Require().NotNil(saved.LastError)
	s.Equal("storage unavailable", *saved.LastError)
}

func (s *StorageDeleteTaskRepositoryTestSuite) TestMarkFailed_Negative_NotFound_EquivalencePartitioning() {
	err := s.repo.MarkFailed(context.Background(), 99999, time.Now().UTC(), "storage unavailable")

	s.Error(err)
}

func TestStorageDeleteTaskRepositorySuite(t *testing.T) {
	suite.Run(t, new(StorageDeleteTaskRepositoryTestSuite))
}
