package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AdminServiceTestSuite struct {
	suite.Suite
	mockUserRepo *mocks.UserRepository
	mockRoleRepo *mocks.RoleRepository
	service      service.AdminService
	userMother   mother.UserMother
	roleMother   mother.RoleMother
}

func (s *AdminServiceTestSuite) SetupTest() {
	s.mockUserRepo = mocks.NewUserRepository(s.T())
	s.mockRoleRepo = mocks.NewRoleRepository(s.T())
	s.service = service.NewAdminService(s.mockUserRepo, s.mockRoleRepo)
	s.userMother = mother.UserMother{}
	s.roleMother = mother.RoleMother{}
}

func (s *AdminServiceTestSuite) TestBanUser_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	targetID := 2
	admin := s.userMother.ValidActiveUser()
	admin.ID = 1
	admin.Role = s.roleMother.AdminRole()

	s.mockUserRepo.On("GetByID", ctx, 1).Return(admin, nil)
	s.mockUserRepo.On("Ban", ctx, targetID).Return(nil)

	err := s.service.BanUser(ctx, 1, targetID)

	s.NoError(err)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *AdminServiceTestSuite) TestChangeUserRole_Negative_SelfChange_Combinatorial() {
	ctx := context.Background()
	adminID := 1

	err := s.service.ChangeUserRole(ctx, adminID, adminID, domain.RoleUser)

	s.ErrorIs(err, domain.ErrForbidden)
}

func (s *AdminServiceTestSuite) TestChangeUserRole_Positive_StateTransition() {
	ctx := context.Background()
	adminID := 1
	targetUser := s.userMother.ValidActiveUser()
	targetUser.ID = 2
	targetUser.Role = s.roleMother.DefaultUserRole()
	newRole := s.roleMother.AdminRole()
	admin := s.userMother.ValidActiveUser()
	admin.ID = adminID
	admin.Role = s.roleMother.AdminRole()

	s.mockUserRepo.On("GetByID", ctx, adminID).Return(admin, nil)
	s.mockRoleRepo.On("GetByName", ctx, domain.RoleAdmin).Return(newRole, nil)
	s.mockUserRepo.On("GetByID", ctx, targetUser.ID).Return(targetUser, nil)
	s.mockUserRepo.On("Update", ctx, targetUser).Return(nil)

	err := s.service.ChangeUserRole(ctx, adminID, targetUser.ID, domain.RoleAdmin)

	s.NoError(err)
	s.Equal(newRole.Name, targetUser.Role.Name)
	s.mockUserRepo.AssertExpectations(s.T())
}

func TestAdminServiceSuite(t *testing.T) {
	suite.Run(t, new(AdminServiceTestSuite))
}
