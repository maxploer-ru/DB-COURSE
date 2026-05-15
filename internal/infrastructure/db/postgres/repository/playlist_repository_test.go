package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlaylistRepository_Create(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewPlaylistRepository(db)

	roleID := insertRole(t, "role_p1", false)
	userID := insertUser(t, roleID, "user_p1")
	channelID := insertChannel(t, userID, "channel_p1")

	playlist := &domain.Playlist{ChannelID: channelID, Name: "pl", Description: "d", CreatedAt: time.Now()}
	err := repo.Create(context.Background(), playlist)
	require.NoError(t, err)
	require.NotZero(t, playlist.ID)
}

func TestPlaylistRepository_GetByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewPlaylistRepository(db)

	roleID := insertRole(t, "role_p2", false)
	userID := insertUser(t, roleID, "user_p2")
	channelID := insertChannel(t, userID, "channel_p2")
	videoID := insertVideo(t, channelID, "video_p2")
	playlistID := insertPlaylist(t, channelID, "pl_p2")
	insertPlaylistItem(t, playlistID, videoID, 1)

	playlist, err := repo.GetByID(context.Background(), playlistID)
	require.NoError(t, err)
	require.Equal(t, playlistID, playlist.ID)
	require.Len(t, playlist.Items, 1)
}

func TestPlaylistRepository_ListByChannel(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewPlaylistRepository(db)

	roleID := insertRole(t, "role_p3", false)
	userID := insertUser(t, roleID, "user_p3")
	channelID := insertChannel(t, userID, "channel_p3")
	_ = insertPlaylist(t, channelID, "pl_p3")

	playlists, err := repo.ListByChannel(context.Background(), channelID, 10, 0)
	require.NoError(t, err)
	require.Len(t, playlists, 1)
}

func TestPlaylistRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewPlaylistRepository(db)

	roleID := insertRole(t, "role_p4", false)
	userID := insertUser(t, roleID, "user_p4")
	channelID := insertChannel(t, userID, "channel_p4")
	playlistID := insertPlaylist(t, channelID, "pl_p4")

	playlist, err := repo.GetByID(context.Background(), playlistID)
	require.NoError(t, err)
	playlist.Name = "pl_p4_new"

	err = repo.Update(context.Background(), playlist)
	require.NoError(t, err)
}

func TestPlaylistRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewPlaylistRepository(db)

	roleID := insertRole(t, "role_p5", false)
	userID := insertUser(t, roleID, "user_p5")
	channelID := insertChannel(t, userID, "channel_p5")
	playlistID := insertPlaylist(t, channelID, "pl_p5")

	err := repo.Delete(context.Background(), playlistID)
	require.NoError(t, err)
}

func TestPlaylistRepository_AddVideo(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewPlaylistRepository(db)

	roleID := insertRole(t, "role_p6", false)
	userID := insertUser(t, roleID, "user_p6")
	channelID := insertChannel(t, userID, "channel_p6")
	videoID := insertVideo(t, channelID, "video_p6")
	playlistID := insertPlaylist(t, channelID, "pl_p6")

	err := repo.AddVideo(context.Background(), playlistID, videoID)
	require.NoError(t, err)
}

func TestPlaylistRepository_RemoveVideo(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewPlaylistRepository(db)

	roleID := insertRole(t, "role_p7", false)
	userID := insertUser(t, roleID, "user_p7")
	channelID := insertChannel(t, userID, "channel_p7")
	videoID := insertVideo(t, channelID, "video_p7")
	playlistID := insertPlaylist(t, channelID, "pl_p7")
	insertPlaylistItem(t, playlistID, videoID, 1)

	err := repo.RemoveVideo(context.Background(), playlistID, videoID)
	require.NoError(t, err)
}
