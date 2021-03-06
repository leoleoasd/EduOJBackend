package models

import (
	"github.com/EduOJ/backend/base"
	"gorm.io/gorm"
	"time"
)

const PriorityDefault = uint8(127)

type Submission struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID       uint        `sql:"index" json:"user_id"`
	User         *User       `json:"user"`
	ProblemID    uint        `sql:"index" json:"problem_id"`
	Problem      *Problem    `json:"problem"`
	ProblemSetId uint        `sql:"index" json:"problem_set_id"`
	ProblemSet   *ProblemSet `json:"problem_set"`
	LanguageName string      `json:"language_name"`
	Language     *Language   `json:"language"`
	FileName     string      `json:"file_name"`
	Priority     uint8       `json:"priority"`

	Judged bool `json:"judged"`
	Score  uint `json:"score"`

	/*
		PENDING  / JUDGEMENT_FAILED / NO_COMMENT
		ACCEPTED / WRONG_ANSWER / RUNTIME_ERROR / TIME_LIMIT_EXCEEDED / MEMORY_LIMIT_EXCEEDED / DANGEROUS_SYSTEM_CALLS
	*/
	Status string `json:"status"`

	Runs []Run `json:"runs"`

	CreatedAt time.Time      `sql:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Run struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID       uint        `sql:"index" json:"user_id"`
	User         *User       `json:"user"`
	ProblemID    uint        `sql:"index" json:"problem_id"`
	Problem      *Problem    `json:"problem"`
	ProblemSetId uint        `sql:"index" json:"problem_set_id"`
	TestCaseID   uint        `json:"test_case_id"`
	TestCase     *TestCase   `json:"test_case"`
	Sample       bool        `json:"sample" gorm:"not null"`
	SubmissionID uint        `json:"submission_id"`
	Submission   *Submission `json:"submission"`
	Priority     uint8       `json:"priority"`

	Judged        bool `json:"judged"`
	JudgerName    string
	JudgerMessage string

	/*
		PENDING / JUDGING / JUDGEMENT_FAILED / NO_COMMENT
		ACCEPTED / WRONG_ANSWER / RUNTIME_ERROR / TIME_LIMIT_EXCEEDED / MEMORY_LIMIT_EXCEEDED / DANGEROUS_SYSTEM_CALLS
	*/
	Status             string `json:"status"`
	MemoryUsed         uint   `json:"memory_used"` // Byte
	TimeUsed           uint   `json:"time_used"`   // ms
	OutputStrippedHash string `json:"output_stripped_hash"`

	CreatedAt time.Time      `sql:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (s *Submission) LoadRuns() {
	err := base.DB.Model(s).Association("Runs").Find(&s.Runs)
	if err != nil {
		panic(err)
	}
}

func (s *Submission) AfterDelete(tx *gorm.DB) (err error) {
	return tx.Where("submission_id = ?", s.ID).Delete(&Run{}).Error
}
