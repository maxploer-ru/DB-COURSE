package repository_test

import (
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

func (s *RoleRepositoryTestSuite) TestCreateAndGetByName_Positive_StateTransition() {
	ctx := context.Background()
	role := s.mother.AdminRole()
	role.ID = 0
	role.Name = "super_admin"

	err := s.repo.Create(ctx, role)
	s.NoError(err)
	s.NotZero(role.ID)

	found, err := s.repo.GetByName(ctx, "super_admin")
	s.NoError(err)
	s.Equal(role.Name, found.Name)
}

func (s *RoleRepositoryTestSuite) TestGetByName_Negative_NotFound_EquivalencePartitioning() {
	ctx := context.Background()

	found, err := s.repo.GetByName(ctx, "non_existent")

	s.NoError(err)
	s.Nil(found)
}

func TestRoleRepositorySuite(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}
