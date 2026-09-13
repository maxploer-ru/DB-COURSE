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

type CommentRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.CommentRepository
	userRepo    *repository.UserRepository
	chanRepo    *repository.ChannelRepository
	vidRepo     *repository.VideoRepository
	userMother  mother.UserMother
	chanMother  mother.ChannelMother
	vidMother   mother.VideoMother
	comMother   mother.CommentMother
	testUser    *domain.User
	testVideo   *domain.Video
}

func (s *CommentRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *CommentRepositoryTestSuite) SetupTest() {
	ctx := context.Background()
	s.tx = s.db.Begin()
	s.repo = repository.NewCommentRepository(s.tx)
	s.userRepo = repository.NewUserRepository(s.tx)
	s.chanRepo = repository.NewChannelRepository(s.tx)
	s.vidRepo = repository.NewVideoRepository(s.tx)
	s.userMother = mother.UserMother{}
	s.chanMother = mother.ChannelMother{}
	s.vidMother = mother.VideoMother{}
	s.comMother = mother.CommentMother{}

	u := s.userMother.ValidActiveUser()
	u.ID = 0
	_ = s.userRepo.Create(ctx, u)
	s.testUser = u

	c := s.chanMother.ChannelForUser(s.testUser.ID)
	c.ID = 0
	_ = s.chanRepo.Create(ctx, c)

	v := s.vidMother.ReadyVideoForChannel(c.ID)
	v.ID = 0
	_ = s.vidRepo.Create(ctx, v)
	s.testVideo = v
}

func (s *CommentRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *CommentRepositoryTestSuite) TestCreate_Positive() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(s.testUser.ID, s.testVideo.ID)
	comment.ID = 0

	err := s.repo.Create(ctx, comment)

	s.NoError(err)
	s.NotZero(comment.ID)
}

func (s *CommentRepositoryTestSuite) TestCreate_Negative() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(999999, s.testVideo.ID)
	comment.ID = 0

	err := s.repo.Create(ctx, comment)

	s.Error(err)
}

func (s *CommentRepositoryTestSuite) TestGetByID_Positive() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(s.testUser.ID, s.testVideo.ID)
	comment.ID = 0
	_ = s.repo.Create(ctx, comment)

	found, err := s.repo.GetByID(ctx, comment.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(comment.Content, found.Content)
}

func (s *CommentRepositoryTestSuite) TestGetByID_Negative() {
	ctx := context.Background()

	found, err := s.repo.GetByID(ctx, 999999)

	s.NoError(err)
	s.Nil(found)
}

func (s *CommentRepositoryTestSuite) TestListByVideo_Positive() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(s.testUser.ID, s.testVideo.ID)
	comment.ID = 0
	_ = s.repo.Create(ctx, comment)

	list, err := s.repo.ListByVideo(ctx, s.testVideo.ID, 10, 0)

	s.NoError(err)
	s.Len(list, 1)
}

func (s *CommentRepositoryTestSuite) TestListByVideo_Negative() {
	ctx := context.Background()

	list, err := s.repo.ListByVideo(ctx, 999999, 10, 0)

	s.NoError(err)
	s.Len(list, 0)
}

func (s *CommentRepositoryTestSuite) TestUpdate_Positive() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(s.testUser.ID, s.testVideo.ID)
	comment.ID = 0
	_ = s.repo.Create(ctx, comment)
	comment.Content = "Updated content"

	err := s.repo.Update(ctx, comment)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, comment.ID)
	s.Equal("Updated content", found.Content)
}

func (s *CommentRepositoryTestSuite) TestUpdate_Negative() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(s.testUser.ID, s.testVideo.ID)
	comment.ID = 999999

	err := s.repo.Update(ctx, comment)

	s.ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *CommentRepositoryTestSuite) TestDelete_Positive() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(s.testUser.ID, s.testVideo.ID)
	comment.ID = 0
	_ = s.repo.Create(ctx, comment)

	err := s.repo.Delete(ctx, comment.ID)

	s.NoError(err)
}

func (s *CommentRepositoryTestSuite) TestDelete_Negative() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, 999999)

	s.ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *CommentRepositoryTestSuite) TestCountByVideo_Positive() {
	ctx := context.Background()
	comment := s.comMother.CommentForVideo(s.testUser.ID, s.testVideo.ID)
	comment.ID = 0
	_ = s.repo.Create(ctx, comment)

	count, err := s.repo.CountByVideo(ctx, s.testVideo.ID)

	s.NoError(err)
	s.Equal(int64(1), count)
}

func (s *CommentRepositoryTestSuite) TestCountByVideo_Negative() {
	ctx := context.Background()

	count, err := s.repo.CountByVideo(ctx, 999999)

	s.NoError(err)
	s.Equal(int64(0), count)
}

func TestCommentRepositorySuite(t *testing.T) {
	suite.Run(t, new(CommentRepositoryTestSuite))
}
