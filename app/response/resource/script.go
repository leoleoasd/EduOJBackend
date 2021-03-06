package resource

import (
	"github.com/EduOJ/backend/database/models"
	"time"
)

type Script struct {
	Name      string    `json:"name"`
	Filename  string    `json:"file_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Script) convert(script *models.Script) {
	s.Name = script.Name
	s.Filename = script.Filename
	s.CreatedAt = script.CreatedAt
	s.UpdatedAt = script.UpdatedAt
}

func GetScript(script *models.Script) *Script {
	s := Script{}
	s.convert(script)
	return &s
}

func GetScriptSlice(scripts []*models.Script) []*Script {
	s := make([]*Script, len(scripts))
	for i, script := range scripts {
		s[i] = &Script{}
		s[i].convert(script)
	}
	return s
}
