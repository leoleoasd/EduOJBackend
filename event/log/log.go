package log

import "github.com/EduOJ/backend/base/log"

// EventArgs is the arguments of "log" event.
// Only contains the log itself.
type EventArgs = log.Log

// EventRst is the result of "log" event.
// Cause there is no need for result, this
// is a empty struct.
type EventRst struct{}
