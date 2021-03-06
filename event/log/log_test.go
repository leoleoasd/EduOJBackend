package log

import (
	"github.com/EduOJ/backend/base/event"
	"github.com/EduOJ/backend/base/log"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLogEvent(t *testing.T) {
	lastLog := log.Log{}
	event.RegisterListener("test_log_event", func(arg EventArgs) {
		lastLog = arg
	})
	log := log.Log{
		Level:   log.DEBUG,
		Time:    time.Now(),
		Message: "123",
		Caller:  "233",
	}
	event.FireEvent("test_log_event", log)
	assert.Equal(t, log, lastLog)
}
