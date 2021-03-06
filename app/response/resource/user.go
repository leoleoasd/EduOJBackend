package resource

import "github.com/EduOJ/backend/database/models"

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

type UserForAdmin struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`

	Roles  []Role  `json:"roles"`
	Grades []Grade `json:"grades"`
}

func (p *User) convert(user *models.User) {
	if user == nil {
		return
	}
	p.ID = user.ID
	p.Username = user.Username
	p.Nickname = user.Nickname
	p.Email = user.Email
}

func (p *UserForAdmin) convert(user *models.User) {
	if user == nil {
		return
	}
	p.ID = user.ID
	p.Username = user.Username
	p.Nickname = user.Nickname
	p.Email = user.Email
	p.Roles = GetRoleSlice(user.Roles)
	p.Grades = GetGradeSlice(user.Grades)
}

func GetUser(user *models.User) *User {
	p := User{}
	p.convert(user)
	return &p
}

func GetUserForAdmin(user *models.User) *UserForAdmin {
	p := UserForAdmin{}
	p.convert(user)
	return &p
}

func GetUserSlice(users []*models.User) (profiles []User) {
	profiles = make([]User, len(users))
	for i, user := range users {
		profiles[i] = *GetUser(user)
	}
	return
}
