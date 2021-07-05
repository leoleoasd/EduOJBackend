package models

import (
	"encoding/json"
	"github.com/EduOJ/backend/base"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type ProblemSet struct {
	ID uint `gorm:"primaryKey" json:"id"`

	ClassID     uint   `sql:"index" json:"class_id" gorm:"not null"`
	Class       *Class `json:"class"`
	Name        string `json:"name" gorm:"not null;size:255"`
	Description string `json:"description"`

	Problems []*Problem `json:"problems" gorm:"many2many:problems_in_problem_sets"`
	Grades   []*Grade   `json:"grades"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Grade struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID       uint        `json:"user_id"`
	User         *User       `json:"user"`
	ProblemSetID uint        `json:"problem_set_id"`
	ProblemSet   *ProblemSet `json:"problem_set"`
	ClassID      uint        `json:"class_id"`
	Class        *Class      `json:"class"`

	Detail datatypes.JSON `json:"detail"`
	Total  uint           `json:"total"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`
}

func (p *ProblemSet) AddProblems(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	existingIds := make([]uint, len(p.Problems))
	for i, problem := range p.Problems {
		existingIds[i] = problem.ID
	}
	var problems []Problem
	query := base.DB.Preload("TestCases")
	if len(existingIds) != 0 {
		query = query.Where("id not in (?)", existingIds)
	}
	if err := query.Find(&problems, ids).Error; err != nil {
		return err
	}
	return base.DB.Model(p).Association("Problems").Append(&problems)
}

func (p *ProblemSet) DeleteProblems(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	var problems []Problem
	if err := base.DB.Find(&problems, ids).Error; err != nil {
		return err
	}
	return base.DB.Model(p).Association("Problems").Delete(&problems)
}

func (g *Grade) BeforeSave(tx *gorm.DB) (err error) {
	detail := make(map[uint]uint)
	err = json.Unmarshal(g.Detail, &detail)
	if err != nil {
		return
	}
	g.Total = 0
	for _, score := range detail {
		g.Total += score
	}
	return nil
}

func (p *ProblemSet) AfterDelete(tx *gorm.DB) error {
	err := tx.Model(p).Association("Grades").Clear()
	if err != nil {
		return err
	}
	var submissions []Submission
	err = tx.Where("problem_set_id = ?", p.ID).Find(&submissions).Error
	if err != nil {
		return err
	}
	if len(submissions) == 0 {
		return nil
	}
	return tx.Delete(&submissions).Error
}
