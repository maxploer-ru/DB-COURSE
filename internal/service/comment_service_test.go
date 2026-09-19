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

func (s *CommentServiceTestSuite) TestCreateComment_Positive_EquivalencePartitioning() {
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

func (s *CommentServiceTestSuite) TestCreateComment_Negative_VideoNotFound_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	comment, err := s.service.Create(ctx, 1, 999, "Nice video!")

	s.ErrorIs(err, domain.ErrVideoNotFound)
	s.Nil(comment)
}

func (s *CommentServiceTestSuite) TestCreateComment_Negative_EmptyContent_BoundaryValueAnalysis() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(1)

	s.mockVideoRepo.On("GetByID", ctx, video.ID).Return(video, nil)

	comment, err := s.service.Create(ctx, 1, video.ID, "")

	s.Error(err)
	s.Nil(comment)
}

func (s *CommentServiceTestSuite) TestGetCommentByID_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	expected := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, expected.ID).Return(expected, nil)

	comment, err := s.service.GetByID(ctx, expected.ID)

	s.NoError(err)
	s.Equal(expected.ID, comment.ID)
}

func (s *CommentServiceTestSuite) TestGetCommentByID_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockCommentRepo.On("GetByID", ctx, 999).Return(nil, nil)

	comment, err := s.service.GetByID(ctx, 999)

	s.ErrorIs(err, domain.ErrCommentNotFound)
	s.Nil(comment)
}

func (s *CommentServiceTestSuite) TestListCommentsByVideo_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	video := s.vidMother.ReadyVideoForChannel(1)
	expectedList := []*domain.Comment{s.comMother.CommentForVideo(1, video.ID)}

	s.mockVideoRepo.On("GetByID", ctx, video.ID).Return(video, nil)
	s.mockCommentRepo.On("ListByVideo", ctx, video.ID, 10, 0).Return(expectedList, nil)
	s.mockCountCache.On("GetCommentsCount", ctx, video.ID).Return(int64(len(expectedList)), true, nil)

	list, err := s.service.ListByVideo(ctx, video.ID, 10, 0)

	s.NoError(err)
	s.Len(list.Items, 1)
	s.Equal(int64(1), list.TotalCount)
}

func (s *CommentServiceTestSuite) TestListCommentsByVideo_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockVideoRepo.On("GetByID", ctx, 999).Return(nil, nil)

	list, err := s.service.ListByVideo(ctx, 999, 10, 0)

	s.ErrorIs(err, domain.ErrVideoNotFound)
	s.Nil(list)
}

func (s *CommentServiceTestSuite) TestUpdateComment_Positive_StateTransition() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockCommentRepo.On("Update", ctx, existing).Return(nil)

	updated, err := s.service.Update(ctx, existing.ID, existing.UserID, "Updated text")

	s.NoError(err)
	s.Equal("Updated text", updated.Content)
}

func (s *CommentServiceTestSuite) TestUpdateComment_Negative_Forbidden_Combinatorial() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	updated, err := s.service.Update(ctx, existing.ID, 999, "Updated text")

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(updated)
}

func (s *CommentServiceTestSuite) TestDeleteComment_Positive_Author_Combinatorial() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockCommentRepo.On("Delete", ctx, existing.ID).Return(nil)
	s.mockCountCache.On("DecrComments", ctx, existing.VideoID).Return(nil)

	err := s.service.Delete(ctx, existing.ID, existing.UserID, domain.RoleUser)

	s.NoError(err)
}

func (s *CommentServiceTestSuite) TestDeleteComment_Positive_Moderator_Combinatorial() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockCommentRepo.On("Delete", ctx, existing.ID).Return(nil)
	s.mockCountCache.On("DecrComments", ctx, existing.VideoID).Return(nil)

	err := s.service.Delete(ctx, existing.ID, 999, domain.RoleModerator)

	s.NoError(err)
}

func (s *CommentServiceTestSuite) TestDeleteComment_Negative_Forbidden_Combinatorial() {
	ctx := context.Background()
	existing := s.comMother.CommentForVideo(1, 1)
	video := s.vidMother.ReadyVideoForChannel(1)

	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)
	s.mockVideoRepo.On("GetByID", ctx, existing.VideoID).Return(video, nil)
	s.mockChanSvc.On("IsOwner", ctx, video.ChannelID, 999).Return(false, nil)

	err := s.service.Delete(ctx, existing.ID, 999, domain.RoleUser)

	s.ErrorIs(err, domain.ErrForbidden)
}

func (s *CommentServiceTestSuite) TestGetCommentCount_Positive_CacheHit_StateTransition() {
	ctx := context.Background()

	s.mockCountCache.On("GetCommentsCount", ctx, 1).Return(int64(42), true, nil)

	count, err := s.service.GetCount(ctx, 1)

	s.NoError(err)
	s.Equal(int64(42), count)
}

func (s *CommentServiceTestSuite) TestGetCommentCount_Positive_DBFallback_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockCountCache.On("GetCommentsCount", ctx, 1).Return(int64(0), false, errors.New("cache miss"))
	s.mockCommentRepo.On("CountByVideo", ctx, 1).Return(int64(15), nil)
	s.mockCountCache.On("SetCommentsCount", ctx, 1, int64(15)).Return(nil)

	count, err := s.service.GetCount(ctx, 1)

	s.NoError(err)
	s.Equal(int64(15), count)
}

func (s *CommentServiceTestSuite) TestGetCommentCount_Negative_DatabaseError_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockCountCache.On("GetCommentsCount", ctx, 1).Return(int64(0), false, errors.New("cache miss"))
	s.mockCommentRepo.On("CountByVideo", ctx, 1).Return(int64(0), domain.ErrInternalServer)

	count, err := s.service.GetCount(ctx, 1)

	s.ErrorIs(err, domain.ErrInternalServer)
	s.Zero(count)
}

func (s *CommentServiceTestSuite) TestDeleteComment_Negative_MissingVideo_EquivalencePartitioning() {
	ctx := context.Background()

	existing := s.comMother.CommentForVideo(1, 1)
	s.mockCommentRepo.On("GetByID", ctx, existing.ID).Return(existing, nil)

	s.mockVideoRepo.On("GetByID", ctx, existing.VideoID).Return(nil, nil)

	err := s.service.Delete(ctx, existing.ID, 999, domain.RoleUser)
	s.ErrorIs(err, domain.ErrVideoNotFound)
}

func TestCommentServiceSuite(t *testing.T) {
	suite.Run(t, new(CommentServiceTestSuite))
}
