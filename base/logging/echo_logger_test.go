package logging

import (
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

type errJson struct {
	log.JSON
}

func (j errJson) MarshalJSON() ([]byte, error) {
	return nil, errors.New("test error")
}

func TestEchoLogger(t *testing.T) {
	oldLogger := logger0
	t.Cleanup(func() {
		logger0 = oldLogger
		Debug = logger0.Debug
		Info = logger0.Info
		Warning = logger0.Warning
		Error = logger0.Error
		Fatal = logger0.Fatal
		Debugf = logger0.Debugf
		Infof = logger0.Infof
		Warningf = logger0.Warningf
		Errorf = logger0.Errorf
		Fatalf = logger0.Fatalf
	})
	fl := &fakeLogger{}
	logger0 = fl
	Debug = logger0.Debug
	Info = logger0.Info
	Warning = logger0.Warning
	Error = logger0.Error
	Fatal = logger0.Fatal
	Debugf = logger0.Debugf
	Infof = logger0.Infof
	Warningf = logger0.Warningf
	Errorf = logger0.Errorf
	Fatalf = logger0.Fatalf
	el := &EchoLogger{}
	assert.Equal(t, os.Stdout, el.Output())
	el.SetOutput(os.Stdout)
	assert.Equal(t, "", el.Prefix())
	el.SetPrefix("test_echo_logger")
	assert.Equal(t, "test_echo_logger", el.Prefix())
	assert.Equal(t, log.Lvl(0), el.Level())
	el.SetLevel(0)
	el.SetHeader("")

	tests := []struct {
		f interface{}
		string
	}{
		{el.Print, "Info"},
		{el.Printf, "Infof"},
		{el.Printj, "Info"},
		{el.Debug, "Debug"},
		{el.Debugf, "Debugf"},
		{el.Debugj, "Debug"},
		{el.Info, "Info"},
		{el.Infof, "Infof"},
		{el.Infoj, "Info"},
		{el.Warn, "Warning"},
		{el.Warnf, "Warningf"},
		{el.Warnj, "Warning"},
		{el.Error, "Error"},
		{el.Errorf, "Errorf"},
		{el.Errorj, "Error"},
		{el.Fatal, "Fatal"},
		{el.Fatalf, "Fatalf"},
		{el.Fatalj, "Fatal"},
		{el.Panic, "Fatal"},
		{el.Panicf, "Fatalf"},
		{el.Panicj, "Fatal"},
	}
	for _, test := range tests {
		t.Run("testEchoLogger"+test.string, func(t *testing.T) {
			ff := reflect.ValueOf(test.f)
			ty := reflect.TypeOf(test.f)
			ff.Call([]reflect.Value{
				reflect.New(ty.In(0)).Elem(),
			})
			assert.Equal(t, test.string, fl.lastFunctionCalled)
		})
	}
}
