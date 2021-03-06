package event

import (
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

var eventLock = sync.RWMutex{}
var listeners = map[string][]interface{}{}

// RegisterListener registers a listener with a event name.
// The listener should be a function.
func RegisterListener(eventName string, listener interface{}) {
	eventLock.Lock()
	defer eventLock.Unlock()
	listenerValue := reflect.ValueOf(listener)
	if listenerValue.Kind() != reflect.Func {
		panic("Trying to register a non-func listener!")
	}
	listeners[eventName] = append(listeners[eventName], listener)
}

// FireEvent Fires a given event name with args.
// Args will be passed to all registered listeners.
// Returns a slice of results, each result is a slice of interface {},
// representing the return value of each call.
func FireEvent(eventName string, args ...interface{}) (result [][]interface{}, err error) {
	eventLock.RLock()
	defer eventLock.RUnlock()
	defer func() {
		if p := recover(); p != nil {
			result = nil
			err = errors.New(p.(string))
			return
		}
	}()
	argsValue := make([]reflect.Value, len(args))
	for i, arg := range args {
		argsValue[i] = reflect.ValueOf(arg)
	}
	result = make([][]interface{}, len(listeners[eventName]))
	for i, listener := range listeners[eventName] {
		listenerFunc := reflect.ValueOf(listener)
		rst := listenerFunc.Call(argsValue)
		result[i] = make([]interface{}, len(rst))
		for j, v := range rst {
			result[i][j] = v.Interface()
		}
	}
	return
}
