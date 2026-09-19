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

type RoleRepositoryTestSuite struct {
	suite.Suite
	pgContainer *db.PostgresContainer
	db          *gorm.DB
	tx          *gorm.DB
	repo        *repository.RoleRepository
	mother      mother.RoleMother
}

func (s *RoleRepositoryTestSuite) SetupSuite() {
	s.db = sharedDB
	s.mother = mother.RoleMother{}
}

func (s *RoleRepositoryTestSuite) SetupTest() {
	s.tx = s.db.Begin()
	s.repo = repository.NewRoleRepository(s.tx)
}

func (s *RoleRepositoryTestSuite) TearDownTest() {
	s.tx.Rollback()
}

func (s *RoleRepositoryTestSuite) TestCreate_Positive_StateTransition() {
	ctx := context.Background()
	role := s.mother.AdminRole()
	role.ID = 0
	role.Name = "super_admin"

	err := s.repo.Create(ctx, role)
	s.NoError(err)
	s.NotZero(role.ID)
}

func (s *RoleRepositoryTestSuite) TestCreate_Negative_AlreadyExists_EquivalencePartitioning() {
	role := s.mother.AdminRole()
	role.ID = 0
	role.Name = "admin"

	err := s.repo.Create(context.Background(), role)

	s.Error(err)
}

func (s *RoleRepositoryTestSuite) TestGetByName_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	role := s.mother.AdminRole()
	role.ID = 0
	role.Name = "super_admin"
	s.NoError(s.repo.Create(ctx, role))

	found, err := s.repo.GetByName(ctx, "super_admin")
	s.NoError(err)
	s.Equal(role.Name, found.Name)
}

func (s *RoleRepositoryTestSuite) TestGetByID_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	role := s.mother.AdminRole()
	role.ID = 0
	role.Name = "temporary_admin"

	s.NoError(s.repo.Create(ctx, role))

	found, err := s.repo.GetByID(ctx, role.ID)

	s.NoError(err)
	s.Equal(role.Name, found.Name)
}

func (s *RoleRepositoryTestSuite) TestGetByID_Negative_EquivalencePartitioning() {
	found, err := s.repo.GetByID(context.Background(), 99999)

	s.NoError(err)
	s.Nil(found)
}

func (s *RoleRepositoryTestSuite) TestGetByName_Negative_NotFound_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetByName(ctx, "non_existent")

	s.NoError(err)
	s.Nil(found)
}

func (s *RoleRepositoryTestSuite) TestGetDefaultRole_Positive_EquivalencePartitioning() {
	found, err := s.repo.GetDefaultRole(context.Background())

	s.NoError(err)
	s.NotNil(found)
	s.True(found.IsDefault)
}

func (s *RoleRepositoryTestSuite) TestGetDefaultRole_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	found, err := s.repo.GetDefaultRole(ctx)

	s.Error(err)
	s.Nil(found)
}

func (s *RoleRepositoryTestSuite) TestUpdate_Positive_StateTransition() {
	ctx := context.Background()
	role := s.mother.AdminRole()
	role.ID = 0
	role.Name = "role_before_update"
	s.NoError(s.repo.Create(ctx, role))
	role.Name = "role_after_update"

	err := s.repo.Update(ctx, role)
	found, getErr := s.repo.GetByID(ctx, role.ID)

	s.NoError(err)
	s.NoError(getErr)
	s.Equal("role_after_update", found.Name)
}

func (s *RoleRepositoryTestSuite) TestUpdate_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.repo.Update(ctx, &domain.Role{ID: 99999, Name: "cancelled"})

	s.Error(err)
}

func (s *RoleRepositoryTestSuite) TestDelete_Positive_StateTransition() {
	ctx := context.Background()
	role := s.mother.AdminRole()
	role.ID = 0
	role.Name = "role_to_delete"
	s.NoError(s.repo.Create(ctx, role))

	err := s.repo.Delete(ctx, role.ID)
	found, getErr := s.repo.GetByID(ctx, role.ID)

	s.NoError(err)
	s.NoError(getErr)
	s.Nil(found)
}

func (s *RoleRepositoryTestSuite) TestDelete_Negative_CancelledContext_Exception() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.repo.Delete(ctx, 99999)

	s.Error(err)
}

func TestRoleRepositorySuite(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}
