package models

import (
	"time"
)

type Model struct {
	ID        int `gorm:"primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
	//   DeletedAt *time.Time
}
