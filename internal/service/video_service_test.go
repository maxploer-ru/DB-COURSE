package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type VideoServiceTestSuite struct {
	suite.Suite
	mockVideoRepo  *mocks.VideoRepository
	mockSubSvc     *mocks.SubscriptionService
	mockChanSvc    *mocks.ChannelService
	mockStorageSvc *mocks.StorageService
	service        service.VideoService
	mother         mother.VideoMother
}

func (s *VideoServiceTestSuite) SetupTest() {
	s.mockVideoRepo = mocks.NewVideoRepository(s.T())
	s.mockSubSvc = mocks.NewSubscriptionService(s.T())
	s.mockChanSvc = mocks.NewChannelService(s.T())
	s.mockStorageSvc = mocks.NewStorageService(s.T())
	s.service = service.NewVideoService(s.mockVideoRepo, s.mockSubSvc, s.mockChanSvc, s.mockStorageSvc)
	s.mother = mother.VideoMother{}
}

func (s *VideoServiceTestSuite) TestInitUpload_Positive() {
	ctx := context.Background()
	channelID := 1
	userID := 1
	filename := "test.mp4"

	s.mockChanSvc.On("IsOwner", ctx, channelID, userID).Return(true, nil)
	s.mockVideoRepo.On("Create", ctx, mock.AnythingOfType("*domain.Video")).Return(nil)
	s.mockStorageSvc.On("GenerateUploadPresignedURL", ctx, mock.AnythingOfType("string"), 15*time.Minute).Return("http://upload.url", nil)

	vid, url, err := s.service.InitUpload(ctx, channelID, userID, "Title", "Desc", filename)

	s.NoError(err)
	s.NotNil(vid)
	s.Equal("http://upload.url", url)
	s.Equal(domain.VideoStatusPending, vid.Status)
}

func (s *VideoServiceTestSuite) TestInitUpload_Negative() {
	ctx := context.Background()

	s.mockChanSvc.On("IsOwner", ctx, 1, 999).Return(false, nil)

	vid, url, err := s.service.InitUpload(ctx, 1, 999, "Title", "Desc", "test.mp4")

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(vid)
	s.Empty(url)
}

func (s *VideoServiceTestSuite) TestConfirmUpload_Positive() {
	ctx := context.Background()
	vid := s.mother.PendingVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)
	s.mockChanSvc.On("IsOwner", ctx, vid.ChannelID, 1).Return(true, nil)
	s.mockVideoRepo.On("Update", ctx, vid).Return(nil)
	s.mockSubSvc.On("NotifyAboutNewVideo", ctx, vid.ChannelID).Return(nil)

	err := s.service.ConfirmUpload(ctx, vid.ID, 1)

	s.NoError(err)
	s.Equal(domain.VideoStatusReady, vid.Status)
}

func (s *VideoServiceTestSuite) TestConfirmUpload_Negative() {
	ctx := context.Background()

	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	err := s.service.ConfirmUpload(ctx, 999, 1)

	s.ErrorIs(err, domain.ErrVideoNotFound)
}

func (s *VideoServiceTestSuite) TestGetVideo_Positive() {
	ctx := context.Background()
	expected := s.mother.ReadyVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, expected.ID).Return(expected, nil)

	vid, err := s.service.GetVideo(ctx, expected.ID)

	s.NoError(err)
	s.Equal(expected.ID, vid.ID)
}

func (s *VideoServiceTestSuite) TestGetVideo_Negative() {
	ctx := context.Background()

	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	vid, err := s.service.GetVideo(ctx, 999)

	s.ErrorIs(err, domain.ErrVideoNotFound)
	s.Nil(vid)
}

func (s *VideoServiceTestSuite) TestUpdateVideo_Positive() {
	ctx := context.Background()
	vid := s.mother.ReadyVideoForChannel(1)
	newTitle := "New Title"

	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)
	s.mockChanSvc.On("IsOwner", ctx, vid.ChannelID, 1).Return(true, nil)
	s.mockVideoRepo.On("Update", ctx, vid).Return(nil)

	updatedVid, err := s.service.UpdateVideo(ctx, vid.ID, 1, &newTitle, nil)

	s.NoError(err)
	s.Equal(newTitle, updatedVid.Title)
}

func (s *VideoServiceTestSuite) TestUpdateVideo_Negative() {
	ctx := context.Background()
	vid := s.mother.ReadyVideoForChannel(1)
	newTitle := "New Title"

	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)
	s.mockChanSvc.On("IsOwner", ctx, vid.ChannelID, 999).Return(false, nil)

	updatedVid, err := s.service.UpdateVideo(ctx, vid.ID, 999, &newTitle, nil)

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(updatedVid)
}

func (s *VideoServiceTestSuite) TestDeleteVideo_Positive() {
	ctx := context.Background()
	vid := s.mother.ReadyVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)
	s.mockVideoRepo.On("Delete", ctx, vid.ID).Return(nil)

	err := s.service.DeleteVideo(ctx, vid.ID, 1, domain.RoleAdmin)

	s.NoError(err)
}

func (s *VideoServiceTestSuite) TestDeleteVideo_Negative() {
	ctx := context.Background()

	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	err := s.service.DeleteVideo(ctx, 999, 1, domain.RoleUser)

	s.ErrorIs(err, domain.ErrVideoNotFound)
}

func (s *VideoServiceTestSuite) TestListChannelVideos_Positive() {
	ctx := context.Background()
	expected := []*domain.Video{s.mother.ReadyVideoForChannel(1)}

	s.mockChanSvc.On("Exists", ctx, 1).Return(true, nil)
	s.mockVideoRepo.On("ListByChannel", ctx, 1, 10, 0).Return(expected, nil)

	videos, err := s.service.ListChannelVideos(ctx, 1, 10, 0)

	s.NoError(err)
	s.Len(videos, 1)
}

func (s *VideoServiceTestSuite) TestListChannelVideos_Negative() {
	ctx := context.Background()

	s.mockChanSvc.On("Exists", ctx, 999).Return(false, nil)

	videos, err := s.service.ListChannelVideos(ctx, 999, 10, 0)

	s.ErrorIs(err, domain.ErrChannelNotFound)
	s.Nil(videos)
}

func (s *VideoServiceTestSuite) TestListMyVideos_Positive() {
	ctx := context.Background()
	chMother := mother.ChannelMother{}
	ch := chMother.ChannelForUser(1)
	expected := []*domain.Video{s.mother.ReadyVideoForChannel(ch.ID)}

	s.mockChanSvc.On("GetChannelByUserID", ctx, 1).Return(ch, nil)
	s.mockVideoRepo.On("ListByChannel", ctx, ch.ID, 10, 0).Return(expected, nil)

	videos, err := s.service.ListMyVideos(ctx, 1, 10, 0)

	s.NoError(err)
	s.Len(videos, 1)
}

func (s *VideoServiceTestSuite) TestListMyVideos_Negative() {
	ctx := context.Background()

	s.mockChanSvc.On("GetChannelByUserID", ctx, 999).Return(nil, errors.New("db error"))

	videos, err := s.service.ListMyVideos(ctx, 999, 10, 0)

	s.Error(err)
	s.Nil(videos)
}

func (s *VideoServiceTestSuite) TestListAllVideos_Positive() {
	ctx := context.Background()
	expected := []*domain.Video{s.mother.ReadyVideoForChannel(1)}

	s.mockVideoRepo.On("List", ctx, 10, 0).Return(expected, nil)

	videos, err := s.service.ListAllVideos(ctx, 10, 0)

	s.NoError(err)
	s.Len(videos, 1)
}

func (s *VideoServiceTestSuite) TestListAllVideos_Negative() {
	ctx := context.Background()

	s.mockVideoRepo.On("List", ctx, 10, 0).Return(nil, errors.New("db error"))

	videos, err := s.service.ListAllVideos(ctx, 10, 0)

	s.Error(err)
	s.Nil(videos)
}

func (s *VideoServiceTestSuite) TestGetStreamingPresignedURL_Positive() {
	ctx := context.Background()
	vid := s.mother.ReadyVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)
	s.mockStorageSvc.On("GenerateAccessPresignedURL", ctx, vid.Filepath, 1*time.Hour).Return("http://stream.url", nil)

	url, err := s.service.GetStreamingPresignedURL(ctx, vid.ID)

	s.NoError(err)
	s.Equal("http://stream.url", url)
}

func (s *VideoServiceTestSuite) TestGetStreamingPresignedURL_Negative() {
	ctx := context.Background()
	vid := s.mother.PendingVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)

	url, err := s.service.GetStreamingPresignedURL(ctx, vid.ID)

	s.Error(err)
	s.Empty(url)
}

func TestVideoServiceSuite(t *testing.T) {
	suite.Run(t, new(VideoServiceTestSuite))
}
