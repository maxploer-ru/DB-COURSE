package service_test

import (
	"ZVideo/internal/domain"
	service "ZVideo/internal/service"
	"ZVideo/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommunityService_GetChannelCommunity(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	channel := &domain.Channel{ID: 1, Name: "c"}
	post := &domain.CommunityPost{ID: 2, ChannelID: 1, UserID: 10, Content: "post"}
	comment := &domain.CommunityComment{ID: 3, PostID: 2, UserID: 11, Content: "comment"}

	channelSvc.On("GetChannel", ctx, 1).Return(channel, nil)
	communityRepo.On("ListPostsByChannel", ctx, 1, 100, 0).Return([]*domain.CommunityPost{post}, nil)
	communityRepo.On("ListCommentsByPost", ctx, 2, 100, 0).Return([]*domain.CommunityComment{comment}, nil)
	userRepo.On("GetByID", ctx, 10).Return(&domain.User{ID: 10, Username: "u10"}, nil)
	userRepo.On("GetByID", ctx, 11).Return(&domain.User{ID: 11, Username: "u11"}, nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	community, err := svc.GetChannelCommunity(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "u10", community.Posts[0].Post.Username)
	require.Equal(t, "u11", community.Posts[0].Comments[0].Username)
}

func TestCommunityService_GetMyCommunity(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	channel := &domain.Channel{ID: 2, UserID: 5}
	channelSvc.On("GetChannelByUserID", ctx, 5).Return(channel, nil)
	channelSvc.On("GetChannel", ctx, 2).Return(channel, nil)
	communityRepo.On("ListPostsByChannel", ctx, 2, 100, 0).Return([]*domain.CommunityPost{}, nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	community, err := svc.GetMyCommunity(ctx, 5)
	require.NoError(t, err)
	require.Equal(t, 2, community.Channel.ID)
}

func TestCommunityService_CreatePost(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	channelSvc.On("IsOwner", ctx, 1, 7).Return(true, nil)
	communityRepo.On("CreatePost", ctx, mock.MatchedBy(func(p *domain.CommunityPost) bool {
		return p.ChannelID == 1 && p.UserID == 7 && p.Content == "post"
	})).Return(nil)
	userRepo.On("GetByID", ctx, 7).Return(&domain.User{ID: 7, Username: "u"}, nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	post, err := svc.CreatePost(ctx, 1, 7, "post")
	require.NoError(t, err)
	require.Equal(t, "u", post.Username)
}

func TestCommunityService_UpdatePost(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	post := &domain.CommunityPost{ID: 3, ChannelID: 1, UserID: 7, Content: "old"}
	communityRepo.On("GetPostByID", ctx, 3).Return(post, nil)
	channelSvc.On("IsOwner", ctx, 1, 7).Return(true, nil)
	communityRepo.On("UpdatePost", ctx, post).Return(nil)
	userRepo.On("GetByID", ctx, 7).Return(&domain.User{ID: 7, Username: "u"}, nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	updated, err := svc.UpdatePost(ctx, 3, 7, "new")
	require.NoError(t, err)
	require.Equal(t, "new", updated.Content)
}

func TestCommunityService_DeletePost(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	post := &domain.CommunityPost{ID: 4, ChannelID: 1, UserID: 7}
	communityRepo.On("GetPostByID", ctx, 4).Return(post, nil)
	channelSvc.On("IsOwner", ctx, 1, 7).Return(true, nil)
	communityRepo.On("DeletePost", ctx, 4).Return(nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	err := svc.DeletePost(ctx, 4, 7)
	require.NoError(t, err)
}

func TestCommunityService_CreateComment(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	communityRepo.On("GetPostByID", ctx, 5).Return(&domain.CommunityPost{ID: 5}, nil)
	communityRepo.On("CreateComment", ctx, mock.MatchedBy(func(c *domain.CommunityComment) bool {
		return c.PostID == 5 && c.UserID == 9 && c.Content == "comment"
	})).Return(nil)
	userRepo.On("GetByID", ctx, 9).Return(&domain.User{ID: 9, Username: "u"}, nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	comment, err := svc.CreateComment(ctx, 5, 9, "comment")
	require.NoError(t, err)
	require.Equal(t, "u", comment.Username)
}

func TestCommunityService_UpdateComment(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	comment := &domain.CommunityComment{ID: 6, PostID: 5, UserID: 9, Content: "old"}
	communityRepo.On("GetCommentByID", ctx, 6).Return(comment, nil)
	communityRepo.On("UpdateComment", ctx, comment).Return(nil)
	userRepo.On("GetByID", ctx, 9).Return(&domain.User{ID: 9, Username: "u"}, nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	updated, err := svc.UpdateComment(ctx, 6, 9, "new")
	require.NoError(t, err)
	require.Equal(t, "new", updated.Content)
}

func TestCommunityService_DeleteComment(t *testing.T) {
	ctx := context.Background()
	communityRepo := mocks.NewCommunityRepository(t)
	channelSvc := mocks.NewChannelService(t)
	userRepo := mocks.NewUserRepository(t)

	comment := &domain.CommunityComment{ID: 7, PostID: 5, UserID: 9}
	communityRepo.On("GetCommentByID", ctx, 7).Return(comment, nil)
	communityRepo.On("DeleteComment", ctx, 7).Return(nil)

	svc := service.NewCommunityService(communityRepo, channelSvc, userRepo)
	err := svc.DeleteComment(ctx, 7, 9)
	require.NoError(t, err)
}
