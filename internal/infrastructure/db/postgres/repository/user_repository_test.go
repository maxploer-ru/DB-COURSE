package repository_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/db/postgres/models"
	"ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/testing/db"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.UserRepository
	mother      mother.UserMother
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
}

func (s *UserRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewUserRepository(s.tx)
	s.mother = mother.UserMother{}
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *UserRepositoryTestSuite) TestCreate_Positive() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0

	err := s.repo.Create(ctx, user)

	s.NoError(err)
	s.NotZero(user.ID)

	var savedUser models.User
	s.tx.First(&savedUser, user.ID)
	s.Equal(user.Email, savedUser.Email)
}

func (s *UserRepositoryTestSuite) TestCreate_Negative() {
	ctx := context.Background()
	user1 := s.mother.ValidActiveUser()
	user1.ID = 0
	_ = s.repo.Create(ctx, user1)

	user2 := s.mother.ValidActiveUser()
	user2.ID = 0
	user2.Username = "another_name"

	err := s.repo.Create(ctx, user2)

	s.Error(err)
}

func (s *UserRepositoryTestSuite) TestGetByID_Positive() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	foundUser, err := s.repo.GetByID(ctx, user.ID)

	s.NoError(err)
	s.NotNil(foundUser)
	s.Equal(user.Username, foundUser.Username)
	s.NotNil(foundUser.Role)
}

func (s *UserRepositoryTestSuite) TestGetByID_Negative() {
	ctx := context.Background()

	foundUser, err := s.repo.GetByID(ctx, 999)

	s.NoError(err)
	s.Nil(foundUser)
}

func (s *UserRepositoryTestSuite) TestDelete_Positive() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	err := s.repo.Delete(ctx, user.ID)

	s.NoError(err)

	var count int64
	s.tx.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	s.Equal(int64(0), count)
}

func (s *UserRepositoryTestSuite) TestDelete_Negative() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, 999)

	s.ErrorIs(err, domain.ErrUserNotFound)
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
