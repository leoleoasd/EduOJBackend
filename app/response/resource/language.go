package resource

import (
	"time"
)

type Language struct {
	Name             string    `json:"name"`
	ExtensionAllowed []string  `json:"extension_allowed"`
	BuildScript      *Script   `json:"build_script"`
	RunScript        *Script   `json:"run_script"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
