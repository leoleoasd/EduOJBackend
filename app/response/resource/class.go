package resource

import "github.com/EduOJ/backend/database/models"

type Class struct {
	ID uint `json:"id"`

	Name        string `json:"name"`
	CourseName  string `json:"course_name"`
	Description string `json:"description"`

	Managers    []User       `json:"managers"`
	Students    []User       `json:"students"`
	ProblemSets []ProblemSet `json:"problem_sets"`
}

type ClassDetail struct {
	ID uint `json:"id"`

	Name        string `json:"name"`
	CourseName  string `json:"course_name"`
	Description string `json:"description"`
	InviteCode  string `json:"invite_code"`

	Managers    []User       `json:"managers"`
	Students    []User       `json:"students"`
	ProblemSets []ProblemSet `json:"problem_sets"`
}

func (c *Class) convert(class *models.Class) {
	c.ID = class.ID
	c.Name = class.Name
	c.CourseName = class.CourseName
	c.Description = class.Description
	c.Managers = GetUserSlice(class.Managers)
	c.Students = GetUserSlice(class.Students)
	c.ProblemSets = GetProblemSetSlice(class.ProblemSets)
}

func (c *ClassDetail) convert(class *models.Class) {
	c.ID = class.ID
	c.Name = class.Name
	c.CourseName = class.CourseName
	c.Description = class.Description
	c.InviteCode = class.InviteCode
	c.Managers = GetUserSlice(class.Managers)
	c.Students = GetUserSlice(class.Students)
	c.ProblemSets = GetProblemSetSlice(class.ProblemSets)
}

func GetClass(class *models.Class) *Class {
	c := Class{}
	c.convert(class)
	return &c
}

func GetClassDetail(class *models.Class) *ClassDetail {
	c := ClassDetail{}
	c.convert(class)
	return &c
}

func GetClassSlice(classes []models.Class) (c []Class) {
	c = make([]Class, len(classes))
	for i, class := range classes {
		c[i].convert(&class)
	}
	return
}
