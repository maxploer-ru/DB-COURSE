package worker

import (
	"ZVideo/internal/domain"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type storageTaskRepositoryStub struct {
	tasks      []*domain.StorageDeleteTask
	claimErr   error
	markDoneID int
	markFailed struct {
		id          int
		nextAttempt time.Time
		lastError   string
	}
}

func (s *storageTaskRepositoryStub) ClaimBatch(ctx context.Context, limit int) ([]*domain.StorageDeleteTask, error) {
	if s.claimErr != nil {
		return nil, s.claimErr
	}
	return s.tasks, nil
}

func (s *storageTaskRepositoryStub) MarkDone(ctx context.Context, id int) error {
	s.markDoneID = id
	return nil
}

func (s *storageTaskRepositoryStub) MarkFailed(ctx context.Context, id int, nextAttemptAt time.Time, lastError string) error {
	s.markFailed.id = id
	s.markFailed.nextAttempt = nextAttemptAt
	s.markFailed.lastError = lastError
	return nil
}

type objectDeleterStub struct {
	err error
}

func (s objectDeleterStub) DeleteObject(context.Context, string) error {
	return s.err
}

func TestStorageDeleteWorkerProcessOnce_Positive_StateTransition(t *testing.T) {
	repo := &storageTaskRepositoryStub{
		tasks: []*domain.StorageDeleteTask{{ID: 7, Filepath: "videos/7.mp4"}},
	}
	worker := NewStorageDeleteWorker(repo, objectDeleterStub{}, domain.GetLogger(context.Background()), 10, time.Hour)

	err := worker.ProcessOnce(context.Background())

	require.NoError(t, err)
	require.Equal(t, 7, repo.markDoneID)
}

func TestStorageDeleteWorkerProcessOnce_Negative_DeleteError_StateTransition(t *testing.T) {
	repo := &storageTaskRepositoryStub{
		tasks: []*domain.StorageDeleteTask{{ID: 8, Filepath: "videos/8.mp4", Attempts: 1}},
	}
	worker := NewStorageDeleteWorker(repo, objectDeleterStub{err: errors.New("storage unavailable")}, domain.GetLogger(context.Background()), 10, time.Hour)

	err := worker.ProcessOnce(context.Background())

	require.NoError(t, err)
	require.Equal(t, 8, repo.markFailed.id)
	require.Equal(t, "storage unavailable", repo.markFailed.lastError)
	require.False(t, repo.markFailed.nextAttempt.IsZero())
}

func TestStorageDeleteWorkerRun_Positive_CancelledContext_StateTransition(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := &storageTaskRepositoryStub{claimErr: context.Canceled}
	worker := NewStorageDeleteWorker(repo, objectDeleterStub{}, domain.GetLogger(context.Background()), 10, time.Hour)

	err := worker.Run(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestStorageDeleteWorkerProcessOnce_Negative_ClaimError_Exception(t *testing.T) {
	claimErr := errors.New("database unavailable")
	repo := &storageTaskRepositoryStub{claimErr: claimErr}
	worker := NewStorageDeleteWorker(repo, objectDeleterStub{}, domain.GetLogger(context.Background()), 10, time.Hour)

	err := worker.ProcessOnce(context.Background())

	require.ErrorIs(t, err, claimErr)
}
