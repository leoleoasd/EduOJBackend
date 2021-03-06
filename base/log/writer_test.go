package log

import (
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/event"
	"github.com/EduOJ/backend/base/exit"
	"github.com/EduOJ/backend/database"
	"github.com/EduOJ/backend/database/models"
	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConsoleWriter(t *testing.T) {
	levels := []struct {
		Level
	}{
		{
			DEBUG,
		},
		{
			INFO,
		},
		{
			WARNING,
		},
		{
			ERROR,
		},
		{
			FATAL,
		},
	}
	w := &consoleWriter{}
	ti := time.Now()
	for _, level := range levels {
		out := capturer.CaptureOutput(func() {
			w.log(Log{
				Level:   level.Level,
				Time:    ti,
				Message: "sample log output",
				Caller:  "main.main.func",
			})
		})
		txt := colors[level.Level]("[%s][%s] ▶ %s ", ti.Format("15:04:05"), "main.main.func", level.Level.String()) + "sample log output\n"
		assert.Equal(t, txt, out)
	}
}

func TestEventWriter(t *testing.T) {
	lastLog := Log{}
	done := make(chan struct{})
	event.RegisterListener("log", func(arg Log) {
		lastLog = arg
		done <- struct{}{}
	})
	w := &eventWriter{}
	log := Log{
		Level:   DEBUG,
		Time:    time.Now(),
		Message: "123",
		Caller:  "233",
	}
	w.log(log)
	<-done
	assert.Equal(t, log, lastLog)
}

// This test contains a data race on exit's base context.
// So this file isn't included in the race test.
// This race won't happen in real situation.
// Cause the exit lock shouldn't be replaced out of test.
func TestDatabaseWriter(t *testing.T) {
	t.Cleanup(database.SetupDatabaseForTest())
	t.Cleanup(exit.SetupExitForTest())
	log := Log{
		Level:   DEBUG,
		Time:    time.Now(),
		Message: "123",
		Caller:  "233",
	}
	w := databaseWriter{}
	w.queue = make(chan Log, 100)
	for i := 0; i < 1000; i += 1 {
		w.log(log)
	}
	assert.Equal(t, 100, len(w.queue))
	w.init()
	for i := 0; i < 1000; i += 1 {
		w.log(log)
	}
	exit.Close()
	exit.QuitWG.Wait()
	lm := models.Log{}
	count := int64(-1)
	assert.NoError(t, base.DB.Table("logs").Count(&count).Error)
	assert.LessOrEqual(t, int64(100), count)

	assert.NoError(t, base.DB.First(&lm).Error)
	ll := int(DEBUG)
	assert.Equal(t, models.Log{
		ID:        lm.ID,
		Level:     &ll,
		Message:   "123",
		Caller:    "233",
		CreatedAt: lm.CreatedAt,
	}, lm)
}
