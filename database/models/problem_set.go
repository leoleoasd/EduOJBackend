package models

import (
	"encoding/json"
	"github.com/EduOJ/backend/base"
	"github.com/pkg/errors"
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

	Detail datatypes.JSON `json:"detail"`
	Total  uint           `json:"total"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (p *ProblemSet) AddProblems(ids []uint) error {
	existingIds := make([]uint, len(p.Problems))
	for i, problem := range p.Problems {
		existingIds[i] = problem.ID
	}
	var problems []Problem
	if err := base.DB.Not("id in ?", existingIds).Preload("TestCases").Find(&problems, ids).Error; err != nil {
		return err
	}
	return base.DB.Model(p).Association("Problems").Append(&problems)
}

func (p *ProblemSet) DeleteProblems(ids []uint) error {
	var problems []Problem
	if err := base.DB.Find(&problems, ids).Error; err != nil {
		return err
	}
	return base.DB.Model(p).Association("Problems").Delete(&problems)
}

func UpdateGrade(submission *Submission) error {
	if submission.ProblemSetId == 0 {
		return nil
	}
	grade := Grade{}
	detail := make(map[uint]uint)
	var err error
	err = base.DB.First(&grade, "problem_set_id = ? and user_id = ?", submission.ProblemSetId, submission.UserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			grade = Grade{
				UserID:       submission.UserID,
				ProblemSetID: submission.ProblemSetId,
				Detail:       datatypes.JSON("{}"),
				Total:        0,
			}
		} else {
			return err
		}
	}
	err = json.Unmarshal(grade.Detail, &detail)
	if err != nil {
		return err
	}
	detail[submission.ProblemID] = submission.Score
	grade.Detail, err = json.Marshal(detail)
	return base.DB.Save(&grade).Error
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
	return tx.Delete(&Grade{}, "problem_set_id = ?", p.ID).Error
}
