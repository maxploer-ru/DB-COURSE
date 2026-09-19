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

type CommentRatingRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.CommentRatingRepository
	userRepo    *repository.UserRepository
	chanRepo    *repository.ChannelRepository
	vidRepo     *repository.VideoRepository
	comRepo     *repository.CommentRepository
	userMother  mother.UserMother
	chanMother  mother.ChannelMother
	vidMother   mother.VideoMother
	comMother   mother.CommentMother
	rateMother  mother.CommentRatingMother
	testUser1   *domain.User
	testUser2   *domain.User
	testComment *domain.Comment
}

func (s *CommentRatingRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *CommentRatingRepositoryTestSuite) SetupTest() {
	ctx := context.Background()
	s.tx = s.db.Begin()
	s.repo = repository.NewCommentRatingRepository(s.tx)
	s.userRepo = repository.NewUserRepository(s.tx)
	s.chanRepo = repository.NewChannelRepository(s.tx)
	s.vidRepo = repository.NewVideoRepository(s.tx)
	s.comRepo = repository.NewCommentRepository(s.tx)
	s.userMother = mother.UserMother{}
	s.chanMother = mother.ChannelMother{}
	s.vidMother = mother.VideoMother{}
	s.comMother = mother.CommentMother{}
	s.rateMother = mother.CommentRatingMother{}

	u1 := s.userMother.ValidActiveUser()
	u1.ID = 0
	u1.Username = "user1"
	u1.Email = "u1@test.com"
	_ = s.userRepo.Create(ctx, u1)
	s.testUser1 = u1

	u2 := s.userMother.ValidActiveUser()
	u2.ID = 0
	u2.Username = "user2"
	u2.Email = "u2@test.com"
	_ = s.userRepo.Create(ctx, u2)
	s.testUser2 = u2

	c := s.chanMother.ChannelForUser(s.testUser1.ID)
	c.ID = 0
	_ = s.chanRepo.Create(ctx, c)

	v := s.vidMother.ReadyVideoForChannel(c.ID)
	v.ID = 0
	_ = s.vidRepo.Create(ctx, v)

	com := s.comMother.CommentForVideo(s.testUser1.ID, v.ID)
	com.ID = 0
	_ = s.comRepo.Create(ctx, com)
	s.testComment = com
}

func (s *CommentRatingRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *CommentRatingRepositoryTestSuite) TestCreate_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.rateMother.LikeForComment(s.testUser2.ID, s.testComment.ID)

	err := s.repo.Create(ctx, rating)

	s.NoError(err)
}

func (s *CommentRatingRepositoryTestSuite) TestCreate_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	rating := s.rateMother.LikeForComment(s.testUser2.ID, s.testComment.ID)
	_ = s.repo.Create(ctx, rating)
	duplicate := s.rateMother.DislikeForComment(s.testUser2.ID, s.testComment.ID)

	err := s.repo.Create(ctx, duplicate)

	s.ErrorIs(err, domain.ErrAlreadyRated)
}

func (s *CommentRatingRepositoryTestSuite) TestUpdate_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.rateMother.LikeForComment(s.testUser2.ID, s.testComment.ID)
	_ = s.repo.Create(ctx, rating)
	rating.Liked = false

	err := s.repo.Update(ctx, rating)

	s.NoError(err)
	found, _ := s.repo.GetByUserAndComment(ctx, s.testUser2.ID, s.testComment.ID)
	s.False(found.Liked)
}

func (s *CommentRatingRepositoryTestSuite) TestUpdate_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	rating := s.rateMother.LikeForComment(999999, s.testComment.ID)

	err := s.repo.Update(ctx, rating)

	s.Error(err)
}

func (s *CommentRatingRepositoryTestSuite) TestDelete_Positive_StateTransition() {
	ctx := context.Background()
	rating := s.rateMother.LikeForComment(s.testUser2.ID, s.testComment.ID)
	_ = s.repo.Create(ctx, rating)

	err := s.repo.Delete(ctx, s.testUser2.ID, s.testComment.ID)

	s.NoError(err)
}

func (s *CommentRatingRepositoryTestSuite) TestDelete_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, 999999, 999999)

	s.ErrorIs(err, domain.ErrCommentRatingNotFound)
}

func (s *CommentRatingRepositoryTestSuite) TestGetByUserAndComment_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	rating := s.rateMother.LikeForComment(s.testUser2.ID, s.testComment.ID)
	_ = s.repo.Create(ctx, rating)

	found, err := s.repo.GetByUserAndComment(ctx, s.testUser2.ID, s.testComment.ID)

	s.NoError(err)
	s.NotNil(found)
	s.True(found.Liked)
}

func (s *CommentRatingRepositoryTestSuite) TestGetByUserAndComment_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetByUserAndComment(ctx, 999999, 999999)

	s.NoError(err)
	s.Nil(found)
}

func (s *CommentRatingRepositoryTestSuite) TestGetStats_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	r1 := s.rateMother.LikeForComment(s.testUser1.ID, s.testComment.ID)
	_ = s.repo.Create(ctx, r1)
	r2 := s.rateMother.DislikeForComment(s.testUser2.ID, s.testComment.ID)
	_ = s.repo.Create(ctx, r2)

	likes, dislikes, err := s.repo.GetStats(ctx, s.testComment.ID)

	s.NoError(err)
	s.Equal(int64(1), likes)
	s.Equal(int64(1), dislikes)
}

func (s *CommentRatingRepositoryTestSuite) TestGetStats_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	likes, dislikes, err := s.repo.GetStats(ctx, 999999)

	s.NoError(err)
	s.Equal(int64(0), likes)
	s.Equal(int64(0), dislikes)
}

func TestCommentRatingRepositorySuite(t *testing.T) {
	suite.Run(t, new(CommentRatingRepositoryTestSuite))
}
