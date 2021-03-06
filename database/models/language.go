package models

import (
	"github.com/EduOJ/backend/database"
	"time"
)

type Language struct {
	Name             string               `gorm:"primaryKey" json:"name"`
	ExtensionAllowed database.StringArray `gorm:"type:string" json:"extension_allowed"`
	BuildScriptName  string               `json:"-"`
	BuildScript      *Script              `gorm:"foreignKey:BuildScriptName" json:"build_script"`
	RunScriptName    string               `json:"-"`
	RunScript        *Script              `gorm:"foreignKey:RunScriptName" json:"run_script"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}
