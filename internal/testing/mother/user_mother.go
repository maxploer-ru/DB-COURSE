package mother

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/testing/builder"
)

type UserMother struct{}

func (m *UserMother) ValidActiveUser() *domain.User {
	return builder.NewUserBuilder().Build()
}

func (m *UserMother) BannedUser() *domain.User {
	return builder.NewUserBuilder().Banned().Build()
}
