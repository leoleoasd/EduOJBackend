package resource

import "github.com/EduOJ/backend/database/models"

type Permission struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

type Role struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Target      *string      `json:"target"`
	Permissions []Permission `json:"permissions"`
	TargetID    uint         `json:"target_id"`
}

func (p *Permission) convert(perm *models.Permission) {
	p.ID = perm.ID
	p.Name = perm.Name
}

func (p *Role) convert(userHasRole *models.UserHasRole) {
	p.Name = userHasRole.Role.Name
	p.Target = userHasRole.Role.Target
	p.TargetID = userHasRole.TargetID
	p.Permissions = make([]Permission, len(userHasRole.Role.Permissions))
	for i, perm := range userHasRole.Role.Permissions {
		p.Permissions[i].convert(&perm)
	}
}

func GetPermission(perm *models.Permission) *Permission {
	p := Permission{}
	p.convert(perm)
	return &p
}

func GetRole(role *models.UserHasRole) *Role {
	p := Role{}
	p.convert(role)
	return &p
}

func GetRoleSlice(roles []models.UserHasRole) (profiles []Role) {
	profiles = make([]Role, len(roles))
	for i, role := range roles {
		profiles[i].convert(&role)
	}
	return
}
