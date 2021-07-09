package models

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID uint `gorm:"primaryKey"`

	UserID uint
	User   User `gorm:"foreignKey:UserID"`

	ReactionID uint
	Reaction   Reaction `gorm:"foreignKey:ReactionID" gorm:"polymorphic:Target"`

	Content string

	TargetID   uint
	TargetType string

	FatherID      uint
	RootCommentID uint

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `sql:"index" json:"deleted_at"`
}

func (comment *Comment) AfterDelete(tx *gorm.DB) (err error) {
	if comment.ID == 0 {
		return nil
	}
	if err := tx.Where("father_id = ?", comment.ID).Delete(&Comment{}).Error; err != nil {
		return err
	}
	return tx.Where("id = ?", comment.ReactionID).Delete(&Reaction{}).Error
}
