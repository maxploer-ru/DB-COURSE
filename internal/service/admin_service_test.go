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

func (s *AdminServiceTestSuite) TestBanUser_Negative_Forbidden_Combinatorial() {
	ctx := context.Background()
	admin := s.userMother.ValidActiveUser()
	admin.ID = 1
	admin.Role = s.roleMother.AdminRole()

	s.mockUserRepo.On("GetByID", ctx, admin.ID).Return(admin, nil)

	err := s.service.BanUser(ctx, admin.ID, admin.ID)

	s.ErrorIs(err, domain.ErrForbidden)
}

func (s *AdminServiceTestSuite) TestUnbanUser_Positive_StateTransition() {
	ctx := context.Background()
	admin := s.userMother.ValidActiveUser()
	admin.ID = 1
	admin.Role = s.roleMother.AdminRole()

	s.mockUserRepo.On("GetByID", ctx, admin.ID).Return(admin, nil)
	s.mockUserRepo.On("Unban", ctx, 2).Return(nil)

	err := s.service.UnbanUser(ctx, admin.ID, 2)

	s.NoError(err)
}

func (s *AdminServiceTestSuite) TestUnbanUser_Negative_Forbidden_Combinatorial() {
	ctx := context.Background()
	admin := s.userMother.ValidActiveUser()
	admin.ID = 1
	admin.Role = s.roleMother.AdminRole()

	s.mockUserRepo.On("GetByID", ctx, admin.ID).Return(admin, nil)

	err := s.service.UnbanUser(ctx, admin.ID, admin.ID)

	s.ErrorIs(err, domain.ErrForbidden)
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

func (s *AdminServiceTestSuite) TestListUsers_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	admin := s.userMother.ValidActiveUser()
	admin.ID = 1
	admin.Role = s.roleMother.AdminRole()
	users := []*domain.User{s.userMother.ValidActiveUser()}

	s.mockUserRepo.On("GetByID", ctx, admin.ID).Return(admin, nil)
	s.mockUserRepo.On("ListUsers", ctx, 10, 0).Return(users, nil)
	s.mockUserRepo.On("CountUsers", ctx).Return(int64(1), nil)

	page, err := s.service.ListUsers(ctx, admin.ID, 10, 0)

	s.NoError(err)
	s.Len(page.Items, 1)
	s.Equal(int64(1), page.TotalCount)
}

func (s *AdminServiceTestSuite) TestListUsers_Negative_Forbidden_EquivalencePartitioning() {
	ctx := context.Background()
	nonAdmin := s.userMother.ValidActiveUser()
	nonAdmin.ID = 1
	nonAdmin.Role = s.roleMother.DefaultUserRole()

	s.mockUserRepo.On("GetByID", ctx, nonAdmin.ID).Return(nonAdmin, nil)

	page, err := s.service.ListUsers(ctx, nonAdmin.ID, 10, 0)

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(page)
}

func TestAdminServiceSuite(t *testing.T) {
	suite.Run(t, new(AdminServiceTestSuite))
}
