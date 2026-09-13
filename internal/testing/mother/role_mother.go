package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type RoleMother struct{}

func (m *RoleMother) DefaultUserRole() *domain.Role {
	return builder.NewRoleBuilder().Build()
}

func (m *RoleMother) AdminRole() *domain.Role {
	return builder.NewRoleBuilder().WithID(2).Admin().Build()
}
