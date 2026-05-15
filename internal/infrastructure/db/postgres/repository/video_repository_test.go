package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVideoRepository_CreateAndGetByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRepository(db)

	roleID := insertRole(t, "role_video_a", false)
	userID := insertUser(t, roleID, "user_video_a")
	channelID := insertChannel(t, userID, "channel_video_a")

	video := &domain.Video{ChannelID: channelID, Title: "title_a", Description: "desc", Filepath: "a.mp4"}
	err := repo.Create(context.Background(), video)
	require.NoError(t, err)
	require.NotZero(t, video.ID)

	loaded, err := repo.GetByID(context.Background(), video.ID)
	require.NoError(t, err)
	require.Equal(t, video.ID, loaded.ID)
	require.Equal(t, "title_a", loaded.Title)
	require.Equal(t, channelID, loaded.ChannelID)
}

func TestVideoRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRepository(db)

	roleID := insertRole(t, "role_video_b", false)
	userID := insertUser(t, roleID, "user_video_b")
	channelID := insertChannel(t, userID, "channel_video_b")

	video := &domain.Video{ChannelID: channelID, Title: "old", Description: "desc", Filepath: "old.mp4"}
	err := repo.Create(context.Background(), video)
	require.NoError(t, err)

	video.Title = "new"
	video.Description = "new-desc"
	video.Filepath = "new.mp4"
	err = repo.Update(context.Background(), video)
	require.NoError(t, err)

	updated, err := repo.GetByID(context.Background(), video.ID)
	require.NoError(t, err)
	require.Equal(t, "new", updated.Title)
	require.Equal(t, "new.mp4", updated.Filepath)
}

func TestVideoRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRepository(db)

	roleID := insertRole(t, "role_video_c", false)
	userID := insertUser(t, roleID, "user_video_c")
	channelID := insertChannel(t, userID, "channel_video_c")

	video := &domain.Video{ChannelID: channelID, Title: "title_c", Description: "desc", Filepath: "c.mp4"}
	err := repo.Create(context.Background(), video)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), video.ID)
	require.NoError(t, err)

	loaded, err := repo.GetByID(context.Background(), video.ID)
	require.NoError(t, err)
	require.Nil(t, loaded)
}

func TestVideoRepository_ListAndListByChannel(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRepository(db)

	roleID := insertRole(t, "role_video_d", false)
	userID := insertUser(t, roleID, "user_video_d")
	channelA := insertChannel(t, userID, "channel_video_d1")
	channelB := insertChannel(t, userID, "channel_video_d2")

	_ = insertVideo(t, channelA, "video_d1")
	_ = insertVideo(t, channelA, "video_d2")
	_ = insertVideo(t, channelB, "video_d3")

	all, err := repo.List(context.Background(), 10, 0, domain.VideoSortNewest)
	require.NoError(t, err)
	require.Len(t, all, 3)

	byChannel, err := repo.ListByChannel(context.Background(), channelA, 10, 0, domain.VideoSortNewest)
	require.NoError(t, err)
	require.Len(t, byChannel, 2)
}

func TestVideoRepository_ListSortViews(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRepository(db)

	roleID := insertRole(t, "role_video_e", false)
	userID := insertUser(t, roleID, "user_video_e")
	channelID := insertChannel(t, userID, "channel_video_e")

	videoA := insertVideo(t, channelID, "video_e1")
	videoB := insertVideo(t, channelID, "video_e2")
	insertViewing(t, userID, videoA)
	insertViewing(t, userID, videoA)

	videos, err := repo.List(context.Background(), 10, 0, domain.VideoSortViews)
	require.NoError(t, err)
	require.Len(t, videos, 2)
	require.Equal(t, videoA, videos[0].ID)
	require.Equal(t, videoB, videos[1].ID)
}

func TestVideoRepository_ListSortRating(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRepository(db)

	roleID := insertRole(t, "role_video_f", false)
	userID := insertUser(t, roleID, "user_video_f")
	channelID := insertChannel(t, userID, "channel_video_f")

	videoA := insertVideo(t, channelID, "video_f1")
	videoB := insertVideo(t, channelID, "video_f2")
	insertVideoRating(t, userID, videoA, true)
	insertVideoRating(t, userID, videoB, false)

	videos, err := repo.List(context.Background(), 10, 0, domain.VideoSortRating)
	require.NoError(t, err)
	require.Len(t, videos, 2)
	require.Equal(t, videoA, videos[0].ID)
	require.Equal(t, videoB, videos[1].ID)
}

func TestVideoRepository_ListFilepathsByChannel(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRepository(db)

	roleID := insertRole(t, "role_video_g", false)
	userID := insertUser(t, roleID, "user_video_g")
	channelID := insertChannel(t, userID, "channel_video_g")

	videoA := &domain.Video{ChannelID: channelID, Title: "title_g1", Description: "desc", Filepath: "g1.mp4"}
	videoB := &domain.Video{ChannelID: channelID, Title: "title_g2", Description: "desc", Filepath: "g2.mp4"}
	require.NoError(t, repo.Create(context.Background(), videoA))
	require.NoError(t, repo.Create(context.Background(), videoB))

	paths, err := repo.ListFilepathsByChannel(context.Background(), channelID)
	require.NoError(t, err)
	require.Len(t, paths, 2)
	require.ElementsMatch(t, []string{"g1.mp4", "g2.mp4"}, paths)
}
