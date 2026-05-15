package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestViewingRepository_CreateAndGetTotalViews(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewViewingRepository(db)

	roleID := insertRole(t, "role_view_a", false)
	userID := insertUser(t, roleID, "user_view_a")
	channelID := insertChannel(t, userID, "channel_view_a")
	videoID := insertVideo(t, channelID, "video_view_a")

	viewing := &domain.Viewing{UserID: userID, VideoID: videoID}
	require.NoError(t, repo.Create(context.Background(), viewing))
	require.NoError(t, repo.Create(context.Background(), viewing))

	count, err := repo.GetTotalViews(context.Background(), videoID)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}
