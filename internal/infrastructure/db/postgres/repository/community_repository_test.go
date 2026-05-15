package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommunityRepository_CreatePost(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cp1", false)
	userID := insertUser(t, roleID, "user_cp1")
	channelID := insertChannel(t, userID, "channel_cp1")

	post := &domain.CommunityPost{ChannelID: channelID, UserID: userID, Content: "post"}
	err := repo.CreatePost(context.Background(), post)
	require.NoError(t, err)
	require.NotZero(t, post.ID)
}

func TestCommunityRepository_GetPostByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cp2", false)
	userID := insertUser(t, roleID, "user_cp2")
	channelID := insertChannel(t, userID, "channel_cp2")
	postID := insertCommunityPost(t, channelID, userID, "post")

	post, err := repo.GetPostByID(context.Background(), postID)
	require.NoError(t, err)
	require.Equal(t, postID, post.ID)
}

func TestCommunityRepository_ListPostsByChannel(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cp3", false)
	userID := insertUser(t, roleID, "user_cp3")
	channelID := insertChannel(t, userID, "channel_cp3")
	_ = insertCommunityPost(t, channelID, userID, "post")

	posts, err := repo.ListPostsByChannel(context.Background(), channelID, 10, 0)
	require.NoError(t, err)
	require.Len(t, posts, 1)
}

func TestCommunityRepository_UpdatePost(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cp4", false)
	userID := insertUser(t, roleID, "user_cp4")
	channelID := insertChannel(t, userID, "channel_cp4")
	postID := insertCommunityPost(t, channelID, userID, "old")

	post, err := repo.GetPostByID(context.Background(), postID)
	require.NoError(t, err)
	post.Content = "new"

	err = repo.UpdatePost(context.Background(), post)
	require.NoError(t, err)
}

func TestCommunityRepository_DeletePost(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cp5", false)
	userID := insertUser(t, roleID, "user_cp5")
	channelID := insertChannel(t, userID, "channel_cp5")
	postID := insertCommunityPost(t, channelID, userID, "post")

	err := repo.DeletePost(context.Background(), postID)
	require.NoError(t, err)
}

func TestCommunityRepository_CreateComment(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cc1", false)
	userID := insertUser(t, roleID, "user_cc1")
	channelID := insertChannel(t, userID, "channel_cc1")
	postID := insertCommunityPost(t, channelID, userID, "post")

	comment := &domain.CommunityComment{PostID: postID, UserID: userID, Content: "comment"}
	err := repo.CreateComment(context.Background(), comment)
	require.NoError(t, err)
	require.NotZero(t, comment.ID)
}

func TestCommunityRepository_GetCommentByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cc2", false)
	userID := insertUser(t, roleID, "user_cc2")
	channelID := insertChannel(t, userID, "channel_cc2")
	postID := insertCommunityPost(t, channelID, userID, "post")
	commentID := insertCommunityComment(t, postID, userID, "comment")

	comment, err := repo.GetCommentByID(context.Background(), commentID)
	require.NoError(t, err)
	require.Equal(t, commentID, comment.ID)
}

func TestCommunityRepository_ListCommentsByPost(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cc3", false)
	userID := insertUser(t, roleID, "user_cc3")
	channelID := insertChannel(t, userID, "channel_cc3")
	postID := insertCommunityPost(t, channelID, userID, "post")
	_ = insertCommunityComment(t, postID, userID, "comment")

	comments, err := repo.ListCommentsByPost(context.Background(), postID, 10, 0)
	require.NoError(t, err)
	require.Len(t, comments, 1)
}

func TestCommunityRepository_UpdateComment(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cc4", false)
	userID := insertUser(t, roleID, "user_cc4")
	channelID := insertChannel(t, userID, "channel_cc4")
	postID := insertCommunityPost(t, channelID, userID, "post")
	commentID := insertCommunityComment(t, postID, userID, "old")

	comment, err := repo.GetCommentByID(context.Background(), commentID)
	require.NoError(t, err)
	comment.Content = "new"

	err = repo.UpdateComment(context.Background(), comment)
	require.NoError(t, err)
}

func TestCommunityRepository_DeleteComment(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommunityRepository(db)

	roleID := insertRole(t, "role_cc5", false)
	userID := insertUser(t, roleID, "user_cc5")
	channelID := insertChannel(t, userID, "channel_cc5")
	postID := insertCommunityPost(t, channelID, userID, "post")
	commentID := insertCommunityComment(t, postID, userID, "comment")

	err := repo.DeleteComment(context.Background(), commentID)
	require.NoError(t, err)
}
