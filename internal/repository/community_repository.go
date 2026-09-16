package repository

import (
	"ZVideo/internal/domain"
	"context"
)

type CommunityRepository interface {
	CreatePost(ctx context.Context, post *domain.CommunityPost) error
	GetPostByID(ctx context.Context, id int) (*domain.CommunityPost, error)
	ListPostsByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.CommunityPost, error)
	CountPostsByChannel(ctx context.Context, channelID int) (int64, error)
	UpdatePost(ctx context.Context, post *domain.CommunityPost) error
	DeletePost(ctx context.Context, id int) error

	CreateComment(ctx context.Context, comment *domain.CommunityComment) error
	GetCommentByID(ctx context.Context, id int) (*domain.CommunityComment, error)
	ListCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*domain.CommunityComment, error)
	CountCommentsByPost(ctx context.Context, postID int) (int64, error)
	UpdateComment(ctx context.Context, comment *domain.CommunityComment) error
	DeleteComment(ctx context.Context, id int) error
}
