package models

import (
	"gorm.io/gorm"
)

type User struct {
	Model
	// Email string `gorm:"type:varchar(100);unique_index"`
	// Role         string  `gorm:"size:255"`        // set field size to 255
	// MemberNumber *string `gorm:"unique;not null"` // 指针类型用以规避默认0值陷阱
	// Num          int     `gorm:"AUTO_INCREMENT"`  // set num to auto incrementable
	// Address      string  `gorm:"index:addr"`      // create index with name `addr` for address
	Name      string
	RoleCode  int
	Suspended bool
	Actived   bool
	IgnoreMe  int `gorm:"-"` // ignore this field
}

func (user *User) BeforeCreate(scope *gorm.DB) error {
	// scope.SetColumn("ID", uuid.New())
	return nil
}
