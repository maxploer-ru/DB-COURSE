package mappers

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
)

func ToDomainCommunityPost(model *models.CommunityPost) *domain.CommunityPost {
	if model == nil {
		return nil
	}
	return &domain.CommunityPost{
		ID:        model.ID,
		ChannelID: model.ChannelID,
		UserID:    model.UserID,
		Content:   model.Content,
		CreatedAt: model.CreatedAt,
	}
}

func FromDomainCommunityPost(post *domain.CommunityPost) *models.CommunityPost {
	if post == nil {
		return nil
	}
	return &models.CommunityPost{
		ID:        post.ID,
		ChannelID: post.ChannelID,
		UserID:    post.UserID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	}
}

func ToDomainCommunityPostList(items []models.CommunityPost) []*domain.CommunityPost {
	posts := make([]*domain.CommunityPost, len(items))
	for i, item := range items {
		posts[i] = ToDomainCommunityPost(&item)
	}
	return posts
}

func ToDomainCommunityComment(model *models.CommunityComment) *domain.CommunityComment {
	if model == nil {
		return nil
	}
	return &domain.CommunityComment{
		ID:        model.ID,
		PostID:    model.PostID,
		UserID:    model.UserID,
		Content:   model.Content,
		CreatedAt: model.CreatedAt,
	}
}

func FromDomainCommunityComment(comment *domain.CommunityComment) *models.CommunityComment {
	if comment == nil {
		return nil
	}
	return &models.CommunityComment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func ToDomainCommunityCommentList(items []models.CommunityComment) []*domain.CommunityComment {
	comments := make([]*domain.CommunityComment, len(items))
	for i, item := range items {
		comments[i] = ToDomainCommunityComment(&item)
	}
	return comments
}
