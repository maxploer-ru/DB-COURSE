package builder

import (
	"ZVideo/internal/domain"
	"time"
)

type CommunityPostBuilder struct {
	post *domain.CommunityPost
}

func NewCommunityPostBuilder() *CommunityPostBuilder {
	return &CommunityPostBuilder{
		post: &domain.CommunityPost{
			ID:        1,
			ChannelID: 1,
			UserID:    1,
			Content:   "Test Post Content",
			CreatedAt: time.Now(),
		},
	}
}

func (b *CommunityPostBuilder) WithID(id int) *CommunityPostBuilder {
	b.post.ID = id
	return b
}

func (b *CommunityPostBuilder) WithChannelID(channelID int) *CommunityPostBuilder {
	b.post.ChannelID = channelID
	return b
}

func (b *CommunityPostBuilder) WithUserID(userID int) *CommunityPostBuilder {
	b.post.UserID = userID
	return b
}

func (b *CommunityPostBuilder) Build() *domain.CommunityPost {
	return b.post
}

type CommunityCommentBuilder struct {
	comment *domain.CommunityComment
}

func NewCommunityCommentBuilder() *CommunityCommentBuilder {
	return &CommunityCommentBuilder{
		comment: &domain.CommunityComment{
			ID:        1,
			PostID:    1,
			UserID:    2,
			Content:   "Test Comment Content",
			CreatedAt: time.Now(),
		},
	}
}

func (b *CommunityCommentBuilder) WithID(id int) *CommunityCommentBuilder {
	b.comment.ID = id
	return b
}

func (b *CommunityCommentBuilder) WithPostID(postID int) *CommunityCommentBuilder {
	b.comment.PostID = postID
	return b
}

func (b *CommunityCommentBuilder) WithUserID(userID int) *CommunityCommentBuilder {
	b.comment.UserID = userID
	return b
}

func (b *CommunityCommentBuilder) Build() *domain.CommunityComment {
	return b.comment
}
