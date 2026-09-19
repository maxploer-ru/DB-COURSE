package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CommunityServiceTestSuite struct {
	suite.Suite
	mockRepo    *mocks.CommunityRepository
	mockChanSvc *mocks.ChannelService
	service     service.CommunityService
	mother      mother.CommunityMother
	chanMother  mother.ChannelMother
}

func (s *CommunityServiceTestSuite) SetupTest() {
	s.mockRepo = mocks.NewCommunityRepository(s.T())
	s.mockChanSvc = mocks.NewChannelService(s.T())
	s.service = service.NewCommunityService(s.mockRepo, s.mockChanSvc)
	s.mother = mother.CommunityMother{}
	s.chanMother = mother.ChannelMother{}
}

func (s *CommunityServiceTestSuite) TestGetChannelCommunity_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	ch := s.chanMother.ChannelForUser(1)
	expectedPosts := []*domain.CommunityPost{s.mother.PostForChannel(ch.ID, 1)}

	s.mockChanSvc.On("GetChannel", ctx, ch.ID).Return(ch, nil)
	s.mockRepo.On("ListPostsByChannel", ctx, ch.ID, 10, 0).Return(expectedPosts, nil)
	s.mockRepo.On("CountPostsByChannel", ctx, ch.ID).Return(int64(len(expectedPosts)), nil)

	community, err := s.service.GetChannelCommunity(ctx, ch.ID, 10, 0)

	s.NoError(err)
	s.NotNil(community)
	s.Equal(ch.ID, community.Channel.ID)
	s.Len(community.Posts, 1)
	s.Equal(int64(1), community.TotalCount)
}

func (s *CommunityServiceTestSuite) TestGetChannelCommunity_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockChanSvc.On("GetChannel", ctx, 999).Return(nil, domain.ErrChannelNotFound)

	community, err := s.service.GetChannelCommunity(ctx, 999, 10, 0)

	s.ErrorIs(err, domain.ErrChannelNotFound)
	s.Nil(community)
}

func (s *CommunityServiceTestSuite) TestCreatePost_Positive_StateTransition() {
	ctx := context.Background()

	s.mockChanSvc.On("IsOwner", ctx, 1, 1).Return(true, nil)
	s.mockRepo.On("CreatePost", ctx, mock.AnythingOfType("*domain.CommunityPost")).Return(nil)

	post, err := s.service.CreatePost(ctx, 1, 1, "Content")

	s.NoError(err)
	s.NotNil(post)
	s.Equal("Content", post.Content)
}

func (s *CommunityServiceTestSuite) TestCreatePost_Negative_Forbidden_Combinatorial() {
	ctx := context.Background()

	s.mockChanSvc.On("IsOwner", ctx, 1, 999).Return(false, nil)

	post, err := s.service.CreatePost(ctx, 1, 999, "Content")

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(post)
}

func (s *CommunityServiceTestSuite) TestCreatePost_Negative_EmptyContent_BoundaryValueAnalysis() {
	ctx := context.Background()

	s.mockChanSvc.On("IsOwner", ctx, 1, 1).Return(true, nil)

	post, err := s.service.CreatePost(ctx, 1, 1, "   ")

	s.ErrorIs(err, domain.ErrCommunityPostContentEmpty)
	s.Nil(post)
}

func (s *CommunityServiceTestSuite) TestUpdatePost_Positive_StateTransition() {
	ctx := context.Background()
	post := s.mother.PostForChannel(1, 1)

	s.mockRepo.On("GetPostByID", ctx, post.ID).Return(post, nil)
	s.mockChanSvc.On("IsOwner", ctx, post.ChannelID, 1).Return(true, nil)
	s.mockRepo.On("UpdatePost", ctx, post).Return(nil)

	updated, err := s.service.UpdatePost(ctx, post.ID, 1, "New Content")

	s.NoError(err)
	s.Equal("New Content", updated.Content)
}

func (s *CommunityServiceTestSuite) TestUpdatePost_Negative_Combinatorial() {
	ctx := context.Background()
	post := s.mother.PostForChannel(1, 1)

	s.mockRepo.On("GetPostByID", ctx, post.ID).Return(post, nil)
	s.mockChanSvc.On("IsOwner", ctx, post.ChannelID, 999).Return(false, nil)

	updated, err := s.service.UpdatePost(ctx, post.ID, 999, "New Content")

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(updated)
}

func (s *CommunityServiceTestSuite) TestDeletePost_Positive_StateTransition() {
	ctx := context.Background()
	post := s.mother.PostForChannel(1, 1)

	s.mockRepo.On("GetPostByID", ctx, post.ID).Return(post, nil)
	s.mockChanSvc.On("IsOwner", ctx, post.ChannelID, 1).Return(true, nil)
	s.mockRepo.On("DeletePost", ctx, post.ID).Return(nil)

	err := s.service.DeletePost(ctx, post.ID, 1)

	s.NoError(err)
}

func (s *CommunityServiceTestSuite) TestDeletePost_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockRepo.On("GetPostByID", ctx, 999).Return(nil, nil)

	err := s.service.DeletePost(ctx, 999, 1)

	s.ErrorIs(err, domain.ErrCommunityPostNotFound)
}

func (s *CommunityServiceTestSuite) TestCreateComment_Positive_StateTransition() {
	ctx := context.Background()
	post := s.mother.PostForChannel(1, 1)

	s.mockRepo.On("GetPostByID", ctx, post.ID).Return(post, nil)
	s.mockRepo.On("CreateComment", ctx, mock.AnythingOfType("*domain.CommunityComment")).Return(nil)

	comment, err := s.service.CreateComment(ctx, post.ID, 2, "Comment Content")

	s.NoError(err)
	s.NotNil(comment)
	s.Equal("Comment Content", comment.Content)
}

func (s *CommunityServiceTestSuite) TestCreateComment_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	s.mockRepo.On("GetPostByID", ctx, 999).Return(nil, nil)

	comment, err := s.service.CreateComment(ctx, 999, 2, "Comment Content")

	s.ErrorIs(err, domain.ErrCommunityPostNotFound)
	s.Nil(comment)
}

func (s *CommunityServiceTestSuite) TestUpdateComment_Positive_StateTransition() {
	ctx := context.Background()
	comment := s.mother.CommentForPost(1, 2)

	s.mockRepo.On("GetCommentByID", ctx, comment.ID).Return(comment, nil)
	s.mockRepo.On("UpdateComment", ctx, comment).Return(nil)

	updated, err := s.service.UpdateComment(ctx, comment.ID, 2, "New Content")

	s.NoError(err)
	s.Equal("New Content", updated.Content)
}

func (s *CommunityServiceTestSuite) TestUpdateComment_Negative_Combinatorial() {
	ctx := context.Background()
	comment := s.mother.CommentForPost(1, 2)

	s.mockRepo.On("GetCommentByID", ctx, comment.ID).Return(comment, nil)

	updated, err := s.service.UpdateComment(ctx, comment.ID, 999, "New Content")

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(updated)
}

func (s *CommunityServiceTestSuite) TestDeleteComment_Positive_StateTransition() {
	ctx := context.Background()
	comment := s.mother.CommentForPost(1, 2)

	s.mockRepo.On("GetCommentByID", ctx, comment.ID).Return(comment, nil)
	s.mockRepo.On("DeleteComment", ctx, comment.ID).Return(nil)

	err := s.service.DeleteComment(ctx, comment.ID, 2)

	s.NoError(err)
}

func (s *CommunityServiceTestSuite) TestDeleteComment_Negative_Combinatorial() {
	ctx := context.Background()
	comment := s.mother.CommentForPost(1, 2)

	s.mockRepo.On("GetCommentByID", ctx, comment.ID).Return(comment, nil)

	err := s.service.DeleteComment(ctx, comment.ID, 999)

	s.ErrorIs(err, domain.ErrForbidden)
}

func TestCommunityServiceSuite(t *testing.T) {
	suite.Run(t, new(CommunityServiceTestSuite))
}
