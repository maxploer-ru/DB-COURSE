package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CommentServiceTestSuite struct {
	suite.Suite
	mockCommentRepo *mocks.CommentRepository
	mockVideoRepo   *mocks.VideoRepository
	mockCountCache  *mocks.VideoStatsCache
	mockChanSvc     *mocks.ChannelService
	service         service.CommentService
	comMother       mother.CommentMother
	vidMother       mother.VideoMother
}

func (s *CommentServiceTestSuite) SetupTest() {
	s.mockCommentRepo = mocks.NewCommentRepository(s.T())
	s.mockVideoRepo = mocks.NewVideoRepository(s.T())
	s.mockCountCache = mocks.NewVideoStatsCache(s.T())
	s.mockChanSvc = mocks.NewChannelService(s.T())
	s.service = service.NewCommentService(s.mockCommentRepo, s.mockVideoRepo, s.mockCountCache, s.mockChanSvc)
	s.comMother = mother.CommentMother{}
	s.vidMother = mother.VideoMother{}
}

func (s *CommentServiceTestSuite) TestCreate_Positive() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, video.ID).Return(video, nil)
	s.mockCommentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Comment")).Return(nil)
	s.mockCountCache.On("IncrComments", ctx, video.ID).Return(nil)

	comment, err := s.service.Create(ctx, 1, video.ID, "Nice video!")

	s.NoError(err)
	s.NotNil(comment)
	s.Equal("Nice video!", comment.Content)
}

func (s *CommentServiceTestSuite) TestCreate_Negative_VideoNotFound() {
	ctx := context.Background()

	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	comment, err := s.service.Create(ctx, 1, 999, "Nice video!")

	s.ErrorIs(err, domain.ErrVideoNotFound)
	s.Nil(comment)
}

func (s *CommentServiceTestSuite) TestCreate_Negative_EmptyContent() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, video.ID).Return(video, nil)

	comment, err := s.service.Create(ctx, 1, video.ID, "")

	s.Error(err)
	s.Nil(comment)
}

func (s *CommentServiceTestSuite) TestGetByID_Positive() {
	ctx := context.Background()
	expected := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, expected.ID).Return(expected, nil)

	comment, err := s.service.GetByID(ctx, expected.ID)

	s.NoError(err)
	s.Equal(expected.ID, comment.ID)
}

func (s *CommentServiceTestSuite) TestGetByID_Negative() {
	ctx := context.Background()

	s.mockCommentRepo.On("GetByID", ctx, 999).Return(nil, nil)

	comment, err := s.service.GetByID(ctx, 999)

	s.ErrorIs(err, domain.ErrCommentNotFound)
	s.Nil(comment)
}

func (s *CommentServiceTestSuite) TestListByVideo_Positive() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(1)
	expectedList := []*domain.Comment{s.comMother.CommentForVideo(1, video.ID)}

	s.mockVideoRepo.On("GetByID", ctx, video.ID).Return(video, nil)
	s.mockCommentRepo.On("ListByVideo", ctx, video.ID, 10, 0).Return(expectedList, nil)

	list, err := s.service.ListByVideo(ctx, video.ID, 10, 0)

	s.NoError(err)
	s.Len(list, 1)
}

func (s *CommentServiceTestSuite) TestListByVideo_Negative() {
	ctx := context.Background()

	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	list, err := s.service.ListByVideo(ctx, 999, 10, 0)

	s.ErrorIs(err, domain.ErrVideoNotFound)
	s.Nil(list)
}

func (s *CommentServiceTestSuite) TestUpdate_Positive() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockCommentRepo.On("Update", ctx, existing).Return(nil)

	updated, err := s.service.Update(ctx, existing.ID, existing.UserID, "Updated text")

	s.NoError(err)
	s.Equal("Updated text", updated.Content)
}

func (s *CommentServiceTestSuite) TestUpdate_Negative_Forbidden() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	updated, err := s.service.Update(ctx, existing.ID, 999, "Updated text")

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(updated)
}

func (s *CommentServiceTestSuite) TestDelete_Positive_Author() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockCommentRepo.On("Delete", ctx, existing.ID).Return(nil)
	s.mockCountCache.On("DecrComments", ctx, existing.VideoID).Return(nil)

	err := s.service.Delete(ctx, existing.ID, existing.UserID, domain.RoleUser)

	s.NoError(err)
}

func (s *CommentServiceTestSuite) TestDelete_Positive_Moderator() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockCommentRepo.On("Delete", ctx, existing.ID).Return(nil)
	s.mockCountCache.On("DecrComments", ctx, existing.VideoID).Return(nil)

	err := s.service.Delete(ctx, existing.ID, 999, domain.RoleModerator)

	s.NoError(err)
}

func (s *CommentServiceTestSuite) TestDelete_Negative_Forbidden() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)
	video := s.vidMother.ReadyVideoForChannel(1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockVideoRepo.On("GetByID", ctx, existing.VideoID).Return(video, nil)
	s.mockChanSvc.On("IsOwner", ctx, video.ChannelID, 999).Return(false, nil)

	err := s.service.Delete(ctx, existing.ID, 999, domain.RoleUser)

	s.ErrorIs(err, domain.ErrForbidden)
}

func (s *CommentServiceTestSuite) TestGetCount_Positive_CacheHit() {
	ctx := context.Background()

	s.mockCountCache.On("GetCommentsCount", ctx, 1).Return(int64(42), true, nil)

	count, err := s.service.GetCount(ctx, 1)

	s.NoError(err)
	s.Equal(int64(42), count)
}

func (s *CommentServiceTestSuite) TestGetCount_Positive_DBFallback() {
	ctx := context.Background()

	s.mockCountCache.On("GetCommentsCount", ctx, 1).Return(int64(0), false, errors.New("cache miss"))
	s.mockCommentRepo.On("CountByVideo", ctx, 1).Return(int64(15), nil)
	s.mockCountCache.On("SetCommentsCount", ctx, 1, int64(15)).Return(nil)

	count, err := s.service.GetCount(ctx, 1)

	s.NoError(err)
	s.Equal(int64(15), count)
}

func TestCommentServiceSuite(t *testing.T) {
	suite.Run(t, new(CommentServiceTestSuite))
}
