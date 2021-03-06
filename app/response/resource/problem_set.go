package resource

import (
	"github.com/EduOJ/backend/database/models"
	"github.com/pkg/errors"
	"time"
)

type ProblemSetDetail struct {
	ID uint `json:"id"`

	ClassID     uint   `json:"class_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Problems []Problem `json:"problems"`
	Grades   []Grade   `json:"grades"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type ProblemSet struct {
	ID uint `json:"id"`

	ClassID     uint   `json:"class_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Problems []Problem `json:"problems"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type Grade struct {
	ID uint `json:"id"`

	UserID       uint `json:"user_id"`
	ProblemSetID uint `json:"problem_set_id"`

	Detail string `json:"detail"`
	Total  uint   `json:"total"`
}

func (p *ProblemSetDetail) convert(problemSet *models.ProblemSet) {
	p.ID = problemSet.ID
	p.ClassID = problemSet.ClassID
	p.Name = problemSet.Name
	p.Description = problemSet.Description
	p.Problems = GetProblemSlice(problemSet.Problems)
	p.Grades = GetGradeSlice(problemSet.Grades)
	p.StartTime = problemSet.StartTime
	p.EndTime = problemSet.EndTime
}

func (p *ProblemSet) convert(problemSet *models.ProblemSet) {
	p.ID = problemSet.ID
	p.ClassID = problemSet.ClassID
	p.Name = problemSet.Name
	p.Description = problemSet.Description
	p.Problems = GetProblemSlice(problemSet.Problems)
	p.StartTime = problemSet.StartTime
	p.EndTime = problemSet.EndTime
}

func GetProblemSet(problemSet *models.ProblemSet) *ProblemSet {
	p := ProblemSet{}
	p.convert(problemSet)
	return &p
}

func GetProblemSetDetail(problemSet *models.ProblemSet) *ProblemSetDetail {
	p := ProblemSetDetail{}
	p.convert(problemSet)
	return &p
}

func GetProblemSetSlice(problemSetSlice []*models.ProblemSet) (ps []ProblemSet) {
	ps = make([]ProblemSet, len(problemSetSlice))
	for i, problemSet := range problemSetSlice {
		ps[i].convert(problemSet)
	}
	return
}

func (g *Grade) convert(grade *models.Grade) {
	g.ID = grade.ID
	g.UserID = grade.UserID
	g.ProblemSetID = grade.ProblemSetID
	b, err := grade.Detail.MarshalJSON()
	if err != nil {
		panic(errors.Wrap(err, "could not marshal json for converting grade"))
	}
	g.Detail = string(b)
	g.Total = grade.Total
}

func GetGrade(grade *models.Grade) *Grade {
	g := Grade{}
	g.convert(grade)
	return &g
}

func GetGradeSlice(grades []*models.Grade) (g []Grade) {
	g = make([]Grade, len(grades))
	for i, grade := range grades {
		g[i].convert(grade)
	}
	return
}
