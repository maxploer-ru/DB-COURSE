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

func (s *UserRepositoryTestSuite) TestCreate_Positive_StateTransition() {
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

func (s *UserRepositoryTestSuite) TestCreate_Negative_EquivalencePartitioning() {
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

func (s *UserRepositoryTestSuite) TestGetByID_Positive_EquivalencePartitioning() {
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

func (s *UserRepositoryTestSuite) TestGetByID_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	foundUser, err := s.repo.GetByID(ctx, 999)

	s.NoError(err)
	s.Nil(foundUser)
}

func (s *UserRepositoryTestSuite) TestListUsers_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	list, err := s.repo.ListUsers(ctx, 10, 0)

	s.NoError(err)
	s.GreaterOrEqual(len(list), 1)
}

func (s *UserRepositoryTestSuite) TestListUsers_Negative_BoundaryValueAnalysis() {
	ctx := context.Background()

	list, err := s.repo.ListUsers(ctx, 10, 999)

	s.NoError(err)
	s.Len(list, 0)
}

func (s *UserRepositoryTestSuite) TestGetByIDs_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	foundUsers, err := s.repo.GetByIDs(ctx, []int{user.ID})

	s.NoError(err)
	s.Len(foundUsers, 1)
	s.Equal(user.Username, foundUsers[0].Username)
}

func (s *UserRepositoryTestSuite) TestGetByIDs_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	foundUsers, err := s.repo.GetByIDs(ctx, []int{999})

	s.NoError(err)
	s.Len(foundUsers, 0)
}

func (s *UserRepositoryTestSuite) TestGetByUsername_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	foundUser, err := s.repo.GetByUsername(ctx, user.Username)

	s.NoError(err)
	s.NotNil(foundUser)
	s.Equal(user.ID, foundUser.ID)
}

func (s *UserRepositoryTestSuite) TestGetByUsername_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	foundUser, err := s.repo.GetByUsername(ctx, "unknown")

	s.NoError(err)
	s.Nil(foundUser)
}

func (s *UserRepositoryTestSuite) TestGetByEmail_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	foundUser, err := s.repo.GetByEmail(ctx, user.Email)

	s.NoError(err)
	s.NotNil(foundUser)
	s.Equal(user.ID, foundUser.ID)
}

func (s *UserRepositoryTestSuite) TestGetByEmail_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	foundUser, err := s.repo.GetByEmail(ctx, "unknown@test.com")

	s.NoError(err)
	s.Nil(foundUser)
}

func (s *UserRepositoryTestSuite) TestUpdate_Positive_StateTransition() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)
	user.Username = "updated_name"

	err := s.repo.Update(ctx, user)

	s.NoError(err)
	foundUser, _ := s.repo.GetByID(ctx, user.ID)
	s.Equal("updated_name", foundUser.Username)
}

func (s *UserRepositoryTestSuite) TestUpdate_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	user1 := s.mother.ValidActiveUser()
	user1.ID = 0
	_ = s.repo.Create(ctx, user1)
	user2 := s.mother.ValidActiveUser()
	user2.ID = 0
	user2.Email = "another@test.com"
	user2.Username = "another"
	_ = s.repo.Create(ctx, user2)
	user2.Email = user1.Email

	err := s.repo.Update(ctx, user2)

	s.Error(err)
}

func (s *UserRepositoryTestSuite) TestDelete_Positive_StateTransition() {
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

func (s *UserRepositoryTestSuite) TestDelete_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, 999)

	s.ErrorIs(err, domain.ErrUserNotFound)
}

func (s *UserRepositoryTestSuite) TestExistsByEmail_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	exists, err := s.repo.ExistsByEmail(ctx, user.Email)

	s.NoError(err)
	s.True(exists)
}

func (s *UserRepositoryTestSuite) TestExistsByEmail_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	exists, err := s.repo.ExistsByEmail(ctx, "unknown@test.com")

	s.NoError(err)
	s.False(exists)
}

func (s *UserRepositoryTestSuite) TestExistsByUsername_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	exists, err := s.repo.ExistsByUsername(ctx, user.Username)

	s.NoError(err)
	s.True(exists)
}

func (s *UserRepositoryTestSuite) TestExistsByUsername_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	exists, err := s.repo.ExistsByUsername(ctx, "unknown")

	s.NoError(err)
	s.False(exists)
}

func (s *UserRepositoryTestSuite) TestBan_Positive_StateTransition() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	err := s.repo.Ban(ctx, user.ID)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, user.ID)
	s.False(found.IsActive)
}

func (s *UserRepositoryTestSuite) TestBan_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.Ban(ctx, 999)

	s.ErrorIs(err, domain.ErrUserNotFound)
}

func (s *UserRepositoryTestSuite) TestUnban_Positive_StateTransition() {
	ctx := context.Background()
	user := s.mother.BannedUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	err := s.repo.Unban(ctx, user.ID)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, user.ID)
	s.True(found.IsActive)
}

func (s *UserRepositoryTestSuite) TestUnban_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.Unban(ctx, 999)

	s.ErrorIs(err, domain.ErrUserNotFound)
}

func (s *UserRepositoryTestSuite) TestSetNotificationsEnabled_Positive_StateTransition() {
	ctx := context.Background()
	user := s.mother.ValidActiveUser()
	user.ID = 0
	_ = s.repo.Create(ctx, user)

	err := s.repo.SetNotificationsEnabled(ctx, user.ID, false)

	s.NoError(err)
	found, _ := s.repo.GetByID(ctx, user.ID)
	s.False(found.NotificationsEnabled)
}

func (s *UserRepositoryTestSuite) TestSetNotificationsEnabled_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	err := s.repo.SetNotificationsEnabled(ctx, 999, false)

	s.ErrorIs(err, domain.ErrUserNotFound)
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
