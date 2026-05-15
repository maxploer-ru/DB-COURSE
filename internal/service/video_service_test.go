package service_test

import (
	"ZVideo/internal/domain"
	service "ZVideo/internal/service"
	"ZVideo/mocks"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVideoService_CreateVideo(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	videoRepo.On("Create", ctx, mock.MatchedBy(func(v *domain.Video) bool {
		return v.ChannelID == 2 && v.Filepath == "file"
	})).Return(nil)
	subRepo.On("NotifySubscribersAboutNewVideo", ctx, 2).Return(nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	video, err := svc.CreateVideo(ctx, 2, 5, "t", "d", "file")
	require.NoError(t, err)
	require.Equal(t, 2, video.ChannelID)
}

func TestVideoService_GetUploadPresignedURL(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	storageSvc.On("GenerateUploadPresignedURL", ctx, mock.Anything, 15*time.Minute).Return("url", nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	url, key, err := svc.GetUploadPresignedURL(ctx, 2, 5, "file.mp4")
	require.NoError(t, err)
	require.Equal(t, "url", url)
	require.NotEmpty(t, key)
}

func TestVideoService_GetVideo(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	videoRepo.On("GetByID", ctx, 1).Return(&domain.Video{ID: 1}, nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	video, err := svc.GetVideo(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 1, video.ID)
}

func TestVideoService_UpdateVideo(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	video := &domain.Video{ID: 3, ChannelID: 2, Title: "old"}
	videoRepo.On("GetByID", ctx, 3).Return(video, nil)
	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	videoRepo.On("Update", ctx, video).Return(nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	newTitle := "new"
	updated, err := svc.UpdateVideo(ctx, 3, 5, &newTitle, nil)
	require.NoError(t, err)
	require.Equal(t, "new", updated.Title)
}

func TestVideoService_DeleteVideo(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	videoRepo.On("GetByID", ctx, 4).Return(&domain.Video{ID: 4, ChannelID: 2, Filepath: "file"}, nil)
	channelSvc.On("IsOwner", ctx, 2, 5).Return(true, nil)
	storageSvc.On("DeleteObject", ctx, "file").Return(nil)
	videoRepo.On("Delete", ctx, 4).Return(nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	err := svc.DeleteVideo(ctx, 4, 5)
	require.NoError(t, err)
}

func TestVideoService_ListChannelVideos(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	channelSvc.On("Exists", ctx, 2).Return(true, nil)
	videoRepo.On("ListByChannel", ctx, 2, 10, 0, domain.VideoSortNewest).Return([]*domain.Video{{ID: 1}}, nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	videos, err := svc.ListChannelVideos(ctx, 2, 10, 0, domain.VideoSortNewest)
	require.NoError(t, err)
	require.Len(t, videos, 1)
}

func TestVideoService_ListMyVideos(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	channelSvc.On("GetChannelByUserID", ctx, 5).Return(&domain.Channel{ID: 2, UserID: 5}, nil)
	videoRepo.On("ListByChannel", ctx, 2, 10, 0, domain.VideoSortNewest).Return([]*domain.Video{{ID: 1}}, nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	videos, err := svc.ListMyVideos(ctx, 5, 10, 0, domain.VideoSortNewest)
	require.NoError(t, err)
	require.Len(t, videos, 1)
}

func TestVideoService_ListAllVideos(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	videoRepo.On("List", ctx, 10, 0, domain.VideoSortNewest).Return([]*domain.Video{{ID: 1}}, nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	videos, err := svc.ListAllVideos(ctx, 10, 0, domain.VideoSortNewest)
	require.NoError(t, err)
	require.Len(t, videos, 1)
}

func TestVideoService_GetStreamingPresignedURL(t *testing.T) {
	ctx := context.Background()
	videoRepo := mocks.NewVideoRepository(t)
	subRepo := mocks.NewSubscriptionRepository(t)
	channelSvc := mocks.NewChannelService(t)
	storageSvc := mocks.NewStorageService(t)

	videoRepo.On("GetByID", ctx, 10).Return(&domain.Video{ID: 10, Filepath: "file"}, nil)
	storageSvc.On("GenerateAccessPresignedURL", ctx, "file", time.Hour).Return("url", nil)

	svc := service.NewVideoService(videoRepo, subRepo, channelSvc, storageSvc)
	url, err := svc.GetStreamingPresignedURL(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, "url", url)
}
