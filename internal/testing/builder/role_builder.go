package builder

import "ZVideo/internal/domain"

type RoleBuilder struct {
	role *domain.Role
}

func NewRoleBuilder() *RoleBuilder {
	return &RoleBuilder{
		role: &domain.Role{
			ID:        1,
			Name:      domain.RoleUser,
			IsDefault: true,
		},
	}
}

func (b *RoleBuilder) WithID(id int) *RoleBuilder {
	b.role.ID = id
	return b
}

func (b *RoleBuilder) WithName(name string) *RoleBuilder {
	b.role.Name = name
	return b
}

func (b *RoleBuilder) Admin() *RoleBuilder {
	b.role.Name = domain.RoleAdmin
	b.role.IsDefault = false
	return b
}

func (b *RoleBuilder) Build() *domain.Role {
	return b.role
}
