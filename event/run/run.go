package submission

import "github.com/EduOJ/backend/database/models"

// EventArgs is the arguments of "run" event.
type EventArgs = *models.Run

// EventRst is the result of "run" event.
type EventRst struct{}
