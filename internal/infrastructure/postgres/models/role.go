package models

type Role struct {
	ID   int    `gorm:"type:serial"`
	Name string `gorm:"type:varchar(32);unique;not null"`

	Users []User `gorm:"foreignkey:RoleID"`
}
