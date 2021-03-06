package submission

import "github.com/EduOJ/backend/database/models"

// EventArgs is the arguments of "submission" event.
type EventArgs = *models.Submission

// EventRst is the result of "submission" event.
type EventRst error
