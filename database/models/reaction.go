package models

import (
	"gorm.io/gorm"
	"time"
)

type Reaction struct {
	ID         uint `gorm:"primaryKey"`
	TargetType string
	TargetID   uint

	Details string

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`

	DeletedAt gorm.DeletedAt `sql:"index" json:"deleted_at"`
}
