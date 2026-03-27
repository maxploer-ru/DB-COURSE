package mocks

import (
	"ZVideo/internal/domain/video/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type ViewingRepositoryMock struct {
	mock.Mock
}

func (m *ViewingRepositoryMock) Create(ctx context.Context, video *entity.Viewing) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}
