package repository_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/testing/db"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type CommunityRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.CommunityRepository
	userRepo    *repository.UserRepository
	chanRepo    *repository.ChannelRepository
	userMother  mother.UserMother
	chanMother  mother.ChannelMother
	comMother   mother.CommunityMother
	testUser    *domain.User
	testChannel *domain.Channel
	testPost    *domain.CommunityPost
}

func (s *CommunityRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *CommunityRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewCommunityRepository(s.tx)
	s.userRepo = repository.NewUserRepository(s.tx)
	s.chanRepo = repository.NewChannelRepository(s.tx)
	s.userMother = mother.UserMother{}
	s.chanMother = mother.ChannelMother{}
	s.comMother = mother.CommunityMother{}

	u := s.userMother.ValidActiveUser()
	u.ID = 0
	err := s.userRepo.Create(context.Background(), u)
	s.Require().NoError(err)
	s.testUser = u

	c := s.chanMother.ChannelForUser(s.testUser.ID)
	c.ID = 0
	err = s.chanRepo.Create(context.Background(), c)
	s.Require().NoError(err)
	s.testChannel = c

	p := s.comMother.PostForChannel(s.testChannel.ID, s.testUser.ID)
	p.ID = 0
	err = s.repo.CreatePost(context.Background(), p)
	s.Require().NoError(err)
	s.testPost = p
}

func (s *CommunityRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *CommunityRepositoryTestSuite) TestCreatePost_Positive_StateTransition() {
	ctx := context.Background()
	post := s.comMother.PostForChannel(s.testChannel.ID, s.testUser.ID)
	post.ID = 0

	err := s.repo.CreatePost(ctx, post)

	s.NoError(err)
	s.NotZero(post.ID)
}

func (s *CommunityRepositoryTestSuite) TestCreatePost_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	post := s.comMother.PostForChannel(999999, s.testUser.ID)
	post.ID = 0

	err := s.repo.CreatePost(ctx, post)

	s.Error(err)
}

func (s *CommunityRepositoryTestSuite) TestGetPostByID_Positive_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetPostByID(ctx, s.testPost.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.testPost.Content, found.Content)
}

func (s *CommunityRepositoryTestSuite) TestGetPostByID_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetPostByID(ctx, 999)

	s.NoError(err)
	s.Nil(found)
}

func (s *CommunityRepositoryTestSuite) TestUpdatePost_Positive_StateTransition() {
	ctx := context.Background()
	s.testPost.Content = "Updated Content"

	err := s.repo.UpdatePost(ctx, s.testPost)

	s.NoError(err)
	found, _ := s.repo.GetPostByID(ctx, s.testPost.ID)
	s.Equal("Updated Content", found.Content)
}

func (s *CommunityRepositoryTestSuite) TestUpdatePost_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	post := s.comMother.PostForChannel(s.testChannel.ID, s.testUser.ID)
	post.ID = 999999

	err := s.repo.UpdatePost(ctx, post)

	s.ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *CommunityRepositoryTestSuite) TestDeletePost_Positive_StateTransition() {
	ctx := context.Background()

	err := s.repo.DeletePost(ctx, s.testPost.ID)

	s.NoError(err)
	found, _ := s.repo.GetPostByID(ctx, s.testPost.ID)
	s.Nil(found)
}

func (s *CommunityRepositoryTestSuite) TestDeletePost_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.DeletePost(ctx, 999)

	s.ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *CommunityRepositoryTestSuite) TestListPostsByChannel_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()

	list, err := s.repo.ListPostsByChannel(ctx, s.testChannel.ID, 10, 0)

	s.NoError(err)
	s.Len(list, 1)
}

func (s *CommunityRepositoryTestSuite) TestListPostsByChannel_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	list, err := s.repo.ListPostsByChannel(ctx, 999, 10, 0)

	s.NoError(err)
	s.Len(list, 0)
}

func (s *CommunityRepositoryTestSuite) TestCreateComment_Positive_StateTransition() {
	ctx := context.Background()
	comment := s.comMother.CommentForPost(s.testPost.ID, s.testUser.ID)
	comment.ID = 0

	err := s.repo.CreateComment(ctx, comment)

	s.NoError(err)
	s.NotZero(comment.ID)
}

func (s *CommunityRepositoryTestSuite) TestCreateComment_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	comment := s.comMother.CommentForPost(999999, s.testUser.ID)
	comment.ID = 0

	err := s.repo.CreateComment(ctx, comment)

	s.Error(err)
}

func (s *CommunityRepositoryTestSuite) TestGetCommentByID_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	comment := s.comMother.CommentForPost(s.testPost.ID, s.testUser.ID)
	comment.ID = 0
	_ = s.repo.CreateComment(ctx, comment)

	found, err := s.repo.GetCommentByID(ctx, comment.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(comment.Content, found.Content)
}

func (s *CommunityRepositoryTestSuite) TestGetCommentByID_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetCommentByID(ctx, 999)

	s.NoError(err)
	s.Nil(found)
}

func (s *CommunityRepositoryTestSuite) TestUpdateComment_Positive_StateTransition() {
	ctx := context.Background()
	comment := s.comMother.CommentForPost(s.testPost.ID, s.testUser.ID)
	comment.ID = 0
	_ = s.repo.CreateComment(ctx, comment)
	comment.Content = "Updated Comment"

	err := s.repo.UpdateComment(ctx, comment)

	s.NoError(err)
	found, _ := s.repo.GetCommentByID(ctx, comment.ID)
	s.Equal("Updated Comment", found.Content)
}

func (s *CommunityRepositoryTestSuite) TestUpdateComment_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	comment := s.comMother.CommentForPost(s.testPost.ID, s.testUser.ID)
	comment.ID = 999999

	err := s.repo.UpdateComment(ctx, comment)

	s.ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *CommunityRepositoryTestSuite) TestDeleteComment_Positive_StateTransition() {
	ctx := context.Background()
	comment := s.comMother.CommentForPost(s.testPost.ID, s.testUser.ID)
	comment.ID = 0
	_ = s.repo.CreateComment(ctx, comment)

	err := s.repo.DeleteComment(ctx, comment.ID)

	s.NoError(err)
	found, _ := s.repo.GetCommentByID(ctx, comment.ID)
	s.Nil(found)
}

func (s *CommunityRepositoryTestSuite) TestDeleteComment_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.DeleteComment(ctx, 999)

	s.ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *CommunityRepositoryTestSuite) TestCountPostsByChannel_Positive_EquivalencePartitioning() {
	ctx := context.Background()

	count, err := s.repo.CountPostsByChannel(ctx, s.testChannel.ID)

	s.NoError(err)
	s.GreaterOrEqual(count, int64(1))
}

func (s *CommunityRepositoryTestSuite) TestCountPostsByChannel_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err := s.repo.CountPostsByChannel(ctx, s.testChannel.ID)

	s.Error(err)
	s.Zero(count)
}

func (s *CommunityRepositoryTestSuite) TestCountCommentsByPost_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()

	count, err := s.repo.CountCommentsByPost(ctx, s.testPost.ID)

	s.NoError(err)
	s.GreaterOrEqual(count, int64(0))
}

func (s *CommunityRepositoryTestSuite) TestCountCommentsByPost_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err := s.repo.CountCommentsByPost(ctx, s.testPost.ID)

	s.Error(err)
	s.Zero(count)
}

func TestCommunityRepositorySuite(t *testing.T) {
	suite.Run(t, new(CommunityRepositoryTestSuite))
}
