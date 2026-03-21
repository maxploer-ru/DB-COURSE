package models

type Role struct {
	ID        int    `gorm:"type:serial;primary_key"`
	Name      string `gorm:"type:varchar(32);unique;not null"`
	IsDefault bool   `gorm:"type:boolean;not null;default:false"`

	Users []User `gorm:"foreignkey:RoleID"`
}
