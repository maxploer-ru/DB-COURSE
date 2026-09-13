package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type CommunityMother struct{}

func (m *CommunityMother) PostForChannel(channelID, userID int) *domain.CommunityPost {
	return builder.NewCommunityPostBuilder().
		WithChannelID(channelID).
		WithUserID(userID).
		Build()
}

func (m *CommunityMother) CommentForPost(postID, userID int) *domain.CommunityComment {
	return builder.NewCommunityCommentBuilder().
		WithPostID(postID).
		WithUserID(userID).
		Build()
}
