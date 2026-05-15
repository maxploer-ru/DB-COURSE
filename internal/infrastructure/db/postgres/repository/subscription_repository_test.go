package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionRepository_SubscribeAndIsSubscribed(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewSubscriptionRepository(db)

	roleID := insertRole(t, "role_sub_a", false)
	userID := insertUser(t, roleID, "user_sub_a")
	channelID := insertChannel(t, userID, "channel_sub_a")

	created, err := repo.Subscribe(context.Background(), userID, channelID)
	require.NoError(t, err)
	require.True(t, created)

	created, err = repo.Subscribe(context.Background(), userID, channelID)
	require.NoError(t, err)
	require.False(t, created)

	ok, err := repo.IsSubscribed(context.Background(), userID, channelID)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestSubscriptionRepository_Unsubscribe(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewSubscriptionRepository(db)

	roleID := insertRole(t, "role_sub_b", false)
	userID := insertUser(t, roleID, "user_sub_b")
	channelID := insertChannel(t, userID, "channel_sub_b")
	insertSubscription(t, userID, channelID, 0)

	removed, err := repo.Unsubscribe(context.Background(), userID, channelID)
	require.NoError(t, err)
	require.True(t, removed)

	removed, err = repo.Unsubscribe(context.Background(), userID, channelID)
	require.NoError(t, err)
	require.False(t, removed)
}

func TestSubscriptionRepository_GetSubscribersCount(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewSubscriptionRepository(db)

	roleID := insertRole(t, "role_sub_c", false)
	userA := insertUser(t, roleID, "user_sub_c1")
	userB := insertUser(t, roleID, "user_sub_c2")
	channelID := insertChannel(t, userA, "channel_sub_c")
	insertSubscription(t, userA, channelID, 0)
	insertSubscription(t, userB, channelID, 0)

	count, err := repo.GetSubscribersCount(context.Background(), channelID)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestSubscriptionRepository_GetUserSubscriptions(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewSubscriptionRepository(db)

	roleID := insertRole(t, "role_sub_d", false)
	userID := insertUser(t, roleID, "user_sub_d")
	channelA := insertChannel(t, userID, "channel_sub_d1")
	channelB := insertChannel(t, userID, "channel_sub_d2")
	insertSubscription(t, userID, channelA, 1)
	insertSubscription(t, userID, channelB, 2)

	subs, err := repo.GetUserSubscriptions(context.Background(), userID, 10, 0)
	require.NoError(t, err)
	require.Len(t, subs, 2)
}

func TestSubscriptionRepository_NotifySubscribersAboutNewVideo(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewSubscriptionRepository(db)

	roleID := insertRole(t, "role_sub_e", false)
	userID := insertUser(t, roleID, "user_sub_e")
	channelID := insertChannel(t, userID, "channel_sub_e")
	insertSubscription(t, userID, channelID, 0)

	err := repo.NotifySubscribersAboutNewVideo(context.Background(), channelID)
	require.NoError(t, err)

	sqlDB := ensureTx(t, db)
	var count int
	err = sqlDB.QueryRow("SELECT new_videos_count FROM subscriptions WHERE user_id = $1 AND channel_id = $2", userID, channelID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestSubscriptionRepository_ResetNewVideosCount(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewSubscriptionRepository(db)

	roleID := insertRole(t, "role_sub_f", false)
	userID := insertUser(t, roleID, "user_sub_f")
	channelID := insertChannel(t, userID, "channel_sub_f")
	insertSubscription(t, userID, channelID, 3)

	err := repo.ResetNewVideosCount(context.Background(), userID, channelID)
	require.NoError(t, err)

	sqlDB := ensureTx(t, db)
	var count int
	err = sqlDB.QueryRow("SELECT new_videos_count FROM subscriptions WHERE user_id = $1 AND channel_id = $2", userID, channelID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
