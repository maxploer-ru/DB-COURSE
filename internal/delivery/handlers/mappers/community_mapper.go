package mappers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/domain"
)

func ToCommunityCommentResponse(comment *domain.CommunityComment) dto.CommunityCommentResponse {
	return dto.CommunityCommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Username:  comment.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func ToCommunityPostResponse(post *domain.CommunityPost, comments []*domain.CommunityComment) dto.CommunityPostResponse {
	respComments := make([]dto.CommunityCommentResponse, len(comments))
	for i, comment := range comments {
		respComments[i] = ToCommunityCommentResponse(comment)
	}

	return dto.CommunityPostResponse{
		ID:        post.ID,
		ChannelID: post.ChannelID,
		UserID:    post.UserID,
		Username:  post.Username,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		Comments:  respComments,
	}
}

func ToCommunityResponse(community *domain.Community, subscribersCount int) *dto.CommunityResponse {
	posts := make([]dto.CommunityPostResponse, len(community.Posts))
	for i, item := range community.Posts {
		posts[i] = ToCommunityPostResponse(item.Post, item.Comments)
	}

	return &dto.CommunityResponse{
		Channel: *ToGetChannelResponse(community.Channel, subscribersCount),
		Posts:   posts,
	}
}
