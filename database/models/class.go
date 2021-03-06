package models

import (
	"github.com/EduOJ/backend/base"
	"gorm.io/gorm"
	"time"
)

type Class struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `json:"name" gorm:"size:255;default:'';not null"`
	CourseName  string  `json:"course_name" gorm:"size:255;default:'';not null"`
	Description string  `json:"description"`
	InviteCode  string  `json:"invite_code" gorm:"size:255;default:'';not null"`
	Managers    []*User `json:"managers" gorm:"many2many:user_manage_classes"`
	Students    []*User `json:"students" gorm:"many2many:user_in_classes"`

	ProblemSets []*ProblemSet `json:"problem_sets"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (c Class) GetID() uint {
	return c.ID
}

func (c Class) TypeName() string {
	return "class"
}

func (c *Class) AddStudents(ids []uint) error {
	existingIds := make([]uint, len(c.Students))
	for i, s := range c.Students {
		existingIds[i] = s.ID
	}
	var users []User
	query := base.DB
	if len(existingIds) != 0 {
		query = base.DB.Where("id not in (?)", existingIds)
	}
	if err := query.Find(&users, ids).Error; err != nil {
		return err
	}
	return base.DB.Model(c).Association("Students").Append(&users)
}

func (c *Class) DeleteStudents(ids []uint) error {
	var users []User
	if err := base.DB.Find(&users, ids).Error; err != nil {
		return err
	}
	return base.DB.Model(c).Association("Students").Delete(&users)
}

func (c *Class) AfterDelete(tx *gorm.DB) (err error) {
	return tx.Delete(&ProblemSet{}, "class_id = ?", c.ID).Error
}
