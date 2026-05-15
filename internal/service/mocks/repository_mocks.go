package mocks

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock of repository.UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	args := m.Called(ctx, id)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Ban(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Unban(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) SetNotificationsEnabled(ctx context.Context, id int, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

// MockRoleRepository is a mock of repository.RoleRepository.
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id int) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if role, ok := args.Get(0).(*domain.Role); ok {
		return role, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if role, ok := args.Get(0).(*domain.Role); ok {
		return role, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleRepository) GetDefaultRole(ctx context.Context) (*domain.Role, error) {
	args := m.Called(ctx)
	if role, ok := args.Get(0).(*domain.Role); ok {
		return role, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockRefreshSessionRepository is a mock of repository.RefreshSessionRepository.
type MockRefreshSessionRepository struct {
	mock.Mock
}

func (m *MockRefreshSessionRepository) Save(ctx context.Context, tokenID string, userID int, expiresAt time.Time) error {
	args := m.Called(ctx, tokenID, userID, expiresAt)
	return args.Error(0)
}

func (m *MockRefreshSessionRepository) GetUserID(ctx context.Context, tokenID string) (int, bool, error) {
	args := m.Called(ctx, tokenID)
	return args.Int(0), args.Bool(1), args.Error(2)
}

func (m *MockRefreshSessionRepository) Rotate(ctx context.Context, oldTokenID, newTokenID string, userID int, expiresAt time.Time) (bool, error) {
	args := m.Called(ctx, oldTokenID, newTokenID, userID, expiresAt)
	return args.Bool(0), args.Error(1)
}

func (m *MockRefreshSessionRepository) Delete(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

// MockChannelRepository is a mock of repository.ChannelRepository.
type MockChannelRepository struct {
	mock.Mock
}

func (m *MockChannelRepository) Create(ctx context.Context, channel *domain.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *MockChannelRepository) GetByID(ctx context.Context, id int) (*domain.Channel, error) {
	args := m.Called(ctx, id)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelRepository) GetByUserID(ctx context.Context, userID int) (*domain.Channel, error) {
	args := m.Called(ctx, userID)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelRepository) GetByName(ctx context.Context, name string) (*domain.Channel, error) {
	args := m.Called(ctx, name)
	if channel, ok := args.Get(0).(*domain.Channel); ok {
		return channel, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChannelRepository) Update(ctx context.Context, channel *domain.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *MockChannelRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChannelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// MockVideoRepository is a mock of repository.VideoRepository.
type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Create(ctx context.Context, video *domain.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) GetByID(ctx context.Context, id int) (*domain.Video, error) {
	args := m.Called(ctx, id)
	if video, ok := args.Get(0).(*domain.Video); ok {
		return video, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVideoRepository) Update(ctx context.Context, video *domain.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVideoRepository) List(ctx context.Context, limit, offset int, sort domain.VideoSort) ([]*domain.Video, error) {
	args := m.Called(ctx, limit, offset, sort)
	if videos, ok := args.Get(0).([]*domain.Video); ok {
		return videos, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVideoRepository) ListByChannel(ctx context.Context, channelID int, limit, offset int, sort domain.VideoSort) ([]*domain.Video, error) {
	args := m.Called(ctx, channelID, limit, offset, sort)
	if videos, ok := args.Get(0).([]*domain.Video); ok {
		return videos, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVideoRepository) ListFilepathsByChannel(ctx context.Context, channelID int) ([]string, error) {
	args := m.Called(ctx, channelID)
	if paths, ok := args.Get(0).([]string); ok {
		return paths, args.Error(1)
	}
	return nil, args.Error(1)
}

// MockSubscriptionRepository is a mock of repository.SubscriptionRepository.
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Subscribe(ctx context.Context, userID, channelID int) (bool, error) {
	args := m.Called(ctx, userID, channelID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) Unsubscribe(ctx context.Context, userID, channelID int) (bool, error) {
	args := m.Called(ctx, userID, channelID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) IsSubscribed(ctx context.Context, userID, channelID int) (bool, error) {
	args := m.Called(ctx, userID, channelID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) GetSubscribersCount(ctx context.Context, channelID int) (int, error) {
	args := m.Called(ctx, channelID)
	return args.Int(0), args.Error(1)
}

func (m *MockSubscriptionRepository) GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) ([]*domain.Subscription, error) {
	args := m.Called(ctx, userID, limit, offset)
	if subs, ok := args.Get(0).([]*domain.Subscription); ok {
		return subs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSubscriptionRepository) NotifySubscribersAboutNewVideo(ctx context.Context, channelID int) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) ResetNewVideosCount(ctx context.Context, userID, channelID int) error {
	args := m.Called(ctx, userID, channelID)
	return args.Error(0)
}

// MockPlaylistRepository is a mock of repository.PlaylistRepository.
type MockPlaylistRepository struct {
	mock.Mock
}

func (m *MockPlaylistRepository) Create(ctx context.Context, playlist *domain.Playlist) error {
	args := m.Called(ctx, playlist)
	return args.Error(0)
}

func (m *MockPlaylistRepository) GetByID(ctx context.Context, playlistID int) (*domain.Playlist, error) {
	args := m.Called(ctx, playlistID)
	if playlist, ok := args.Get(0).(*domain.Playlist); ok {
		return playlist, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlaylistRepository) ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Playlist, error) {
	args := m.Called(ctx, channelID, limit, offset)
	if playlists, ok := args.Get(0).([]*domain.Playlist); ok {
		return playlists, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlaylistRepository) Update(ctx context.Context, playlist *domain.Playlist) error {
	args := m.Called(ctx, playlist)
	return args.Error(0)
}

func (m *MockPlaylistRepository) Delete(ctx context.Context, playlistID int) error {
	args := m.Called(ctx, playlistID)
	return args.Error(0)
}

func (m *MockPlaylistRepository) AddVideo(ctx context.Context, playlistID, videoID int) error {
	args := m.Called(ctx, playlistID, videoID)
	return args.Error(0)
}

func (m *MockPlaylistRepository) RemoveVideo(ctx context.Context, playlistID, videoID int) error {
	args := m.Called(ctx, playlistID, videoID)
	return args.Error(0)
}

// MockCommentRepository is a mock of repository.CommentRepository.
type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id int) (*domain.Comment, error) {
	args := m.Called(ctx, id)
	if comment, ok := args.Get(0).(*domain.Comment); ok {
		return comment, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommentRepository) ListByVideo(ctx context.Context, videoID int, limit, offset int) ([]*domain.Comment, error) {
	args := m.Called(ctx, videoID, limit, offset)
	if comments, ok := args.Get(0).([]*domain.Comment); ok {
		return comments, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) CountByVideo(ctx context.Context, videoID int) (int64, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(int64), args.Error(1)
}

// MockCommentRatingRepository is a mock of repository.CommentRatingRepository.
type MockCommentRatingRepository struct {
	mock.Mock
}

func (m *MockCommentRatingRepository) Create(ctx context.Context, rating *domain.CommentRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockCommentRatingRepository) Update(ctx context.Context, rating *domain.CommentRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockCommentRatingRepository) Delete(ctx context.Context, userID, commentID int) error {
	args := m.Called(ctx, userID, commentID)
	return args.Error(0)
}

func (m *MockCommentRatingRepository) GetByUserAndComment(ctx context.Context, userID, commentID int) (*domain.CommentRating, error) {
	args := m.Called(ctx, userID, commentID)
	if rating, ok := args.Get(0).(*domain.CommentRating); ok {
		return rating, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommentRatingRepository) GetStats(ctx context.Context, commentID int) (int64, int64, error) {
	args := m.Called(ctx, commentID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

// MockVideoRatingRepository is a mock of repository.VideoRatingRepository.
type MockVideoRatingRepository struct {
	mock.Mock
}

func (m *MockVideoRatingRepository) Create(ctx context.Context, rating *domain.VideoRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockVideoRatingRepository) Update(ctx context.Context, rating *domain.VideoRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockVideoRatingRepository) Delete(ctx context.Context, userID, videoID int) error {
	args := m.Called(ctx, userID, videoID)
	return args.Error(0)
}

func (m *MockVideoRatingRepository) GetByUserAndVideo(ctx context.Context, userID, videoID int) (*domain.VideoRating, error) {
	args := m.Called(ctx, userID, videoID)
	if rating, ok := args.Get(0).(*domain.VideoRating); ok {
		return rating, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVideoRatingRepository) GetStats(ctx context.Context, videoID int) (int, int, error) {
	args := m.Called(ctx, videoID)
	return args.Int(0), args.Int(1), args.Error(2)
}

// MockViewingRepository is a mock of repository.ViewingRepository.
type MockViewingRepository struct {
	mock.Mock
}

func (m *MockViewingRepository) Create(ctx context.Context, viewing *domain.Viewing) error {
	args := m.Called(ctx, viewing)
	return args.Error(0)
}

func (m *MockViewingRepository) GetTotalViews(ctx context.Context, videoID int) (int, error) {
	args := m.Called(ctx, videoID)
	return args.Int(0), args.Error(1)
}

// MockCommunityRepository is a mock of repository.CommunityRepository.
type MockCommunityRepository struct {
	mock.Mock
}

func (m *MockCommunityRepository) CreatePost(ctx context.Context, post *domain.CommunityPost) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockCommunityRepository) GetPostByID(ctx context.Context, id int) (*domain.CommunityPost, error) {
	args := m.Called(ctx, id)
	if post, ok := args.Get(0).(*domain.CommunityPost); ok {
		return post, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommunityRepository) ListPostsByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.CommunityPost, error) {
	args := m.Called(ctx, channelID, limit, offset)
	if posts, ok := args.Get(0).([]*domain.CommunityPost); ok {
		return posts, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommunityRepository) UpdatePost(ctx context.Context, post *domain.CommunityPost) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockCommunityRepository) DeletePost(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommunityRepository) CreateComment(ctx context.Context, comment *domain.CommunityComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommunityRepository) GetCommentByID(ctx context.Context, id int) (*domain.CommunityComment, error) {
	args := m.Called(ctx, id)
	if comment, ok := args.Get(0).(*domain.CommunityComment); ok {
		return comment, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommunityRepository) ListCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*domain.CommunityComment, error) {
	args := m.Called(ctx, postID, limit, offset)
	if comments, ok := args.Get(0).([]*domain.CommunityComment); ok {
		return comments, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommunityRepository) UpdateComment(ctx context.Context, comment *domain.CommunityComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommunityRepository) DeleteComment(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCommentStatsCache is a mock of repository.CommentStatsCache.
type MockCommentStatsCache struct {
	mock.Mock
}

func (m *MockCommentStatsCache) IncrLikes(ctx context.Context, commentID int) error {
	args := m.Called(ctx, commentID)
	return args.Error(0)
}

func (m *MockCommentStatsCache) DecrLikes(ctx context.Context, commentID int) error {
	args := m.Called(ctx, commentID)
	return args.Error(0)
}

func (m *MockCommentStatsCache) IncrDislikes(ctx context.Context, commentID int) error {
	args := m.Called(ctx, commentID)
	return args.Error(0)
}

func (m *MockCommentStatsCache) DecrDislikes(ctx context.Context, commentID int) error {
	args := m.Called(ctx, commentID)
	return args.Error(0)
}

func (m *MockCommentStatsCache) GetStats(ctx context.Context, commentID int) (int64, int64, bool, error) {
	args := m.Called(ctx, commentID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Bool(2), args.Error(3)
}

func (m *MockCommentStatsCache) SetStats(ctx context.Context, commentID int, likes, dislikes int64) error {
	args := m.Called(ctx, commentID, likes, dislikes)
	return args.Error(0)
}

// MockVideoStatsCache is a mock of repository.VideoStatsCache.
type MockVideoStatsCache struct {
	mock.Mock
}

func (m *MockVideoStatsCache) IncrViews(ctx context.Context, videoID int) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockVideoStatsCache) IncrLikes(ctx context.Context, videoID int) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockVideoStatsCache) DecrLikes(ctx context.Context, videoID int) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockVideoStatsCache) IncrDislikes(ctx context.Context, videoID int) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockVideoStatsCache) DecrDislikes(ctx context.Context, videoID int) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockVideoStatsCache) IncrComments(ctx context.Context, videoID int) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockVideoStatsCache) DecrComments(ctx context.Context, videoID int) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockVideoStatsCache) GetCommentsCount(ctx context.Context, videoID int) (int64, bool, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(int64), args.Bool(1), args.Error(2)
}

func (m *MockVideoStatsCache) SetCommentsCount(ctx context.Context, videoID int, count int64) error {
	args := m.Called(ctx, videoID, count)
	return args.Error(0)
}

func (m *MockVideoStatsCache) GetStats(ctx context.Context, videoID int) (*domain.VideoStats, bool, error) {
	args := m.Called(ctx, videoID)
	if stats, ok := args.Get(0).(*domain.VideoStats); ok {
		return stats, args.Bool(1), args.Error(2)
	}
	return nil, args.Bool(1), args.Error(2)
}

func (m *MockVideoStatsCache) LoadAll(ctx context.Context) (map[int]domain.VideoStats, error) {
	args := m.Called(ctx)
	if stats, ok := args.Get(0).(map[int]domain.VideoStats); ok {
		return stats, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVideoStatsCache) SetStats(ctx context.Context, videoID int, stats *domain.VideoStats) error {
	args := m.Called(ctx, videoID, stats)
	return args.Error(0)
}

// MockSubscriberCounter is a mock of repository.SubscriberCounter.
type MockSubscriberCounter struct {
	mock.Mock
}

func (m *MockSubscriberCounter) Increment(ctx context.Context, channelID int) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockSubscriberCounter) Decrement(ctx context.Context, channelID int) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockSubscriberCounter) Get(ctx context.Context, channelID int) (int, bool, error) {
	args := m.Called(ctx, channelID)
	return args.Int(0), args.Bool(1), args.Error(2)
}

func (m *MockSubscriberCounter) LoadAll(ctx context.Context) (map[int]int, error) {
	args := m.Called(ctx)
	if data, ok := args.Get(0).(map[int]int); ok {
		return data, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSubscriberCounter) Set(ctx context.Context, channelID int, count int) error {
	args := m.Called(ctx, channelID, count)
	return args.Error(0)
}

var _ repository.UserRepository = (*MockUserRepository)(nil)
var _ repository.RoleRepository = (*MockRoleRepository)(nil)
var _ repository.RefreshSessionRepository = (*MockRefreshSessionRepository)(nil)
var _ repository.ChannelRepository = (*MockChannelRepository)(nil)
var _ repository.VideoRepository = (*MockVideoRepository)(nil)
var _ repository.SubscriptionRepository = (*MockSubscriptionRepository)(nil)
var _ repository.PlaylistRepository = (*MockPlaylistRepository)(nil)
var _ repository.CommentRepository = (*MockCommentRepository)(nil)
var _ repository.CommentRatingRepository = (*MockCommentRatingRepository)(nil)
var _ repository.VideoRatingRepository = (*MockVideoRatingRepository)(nil)
var _ repository.ViewingRepository = (*MockViewingRepository)(nil)
var _ repository.CommunityRepository = (*MockCommunityRepository)(nil)
var _ repository.CommentStatsCache = (*MockCommentStatsCache)(nil)
var _ repository.VideoStatsCache = (*MockVideoStatsCache)(nil)
var _ repository.SubscriberCounter = (*MockSubscriberCounter)(nil)
