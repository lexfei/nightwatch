// Package actions provides API to implement cr-monitor actions.
package actions

import (
	"errors"
	"sync"
	"time"
)

// Actor is the interface for actions.
type Actor interface {
	// Init is called when cr-monitor starts monitoring.
	//
	// name is the monitor name.
	// Non-nil error is logged, and STOPS the monitor.
	Init(name string) error

	// Fail is called when a probe is start failing.
	//
	// name is the monitor name.
	// v is the returned value from the probe (or a value from the filter).
	// Non-nil error is logged, but will not stop the monitor.
	Fail(name string, v float64) error

	// Recover is called when a probe is recovered from failure.
	//
	// name is the monitor name.
	// d is the failure duration.
	// Non-nil error is logged, but will not stop the monitor.
	//
	// Note that this may not always be called if cr-monitor is stopped
	// during failure.  Init is the good place to correct such status.
	Recover(name string, d time.Duration) error

	// String returns a descriptive string for this action.
	String() string
}

// Constructor is a function to create an action.
//
// params are configuration options for the action.
type Constructor func(params map[string]interface{}) (Actor, error)

// Errors for actions.
var (
	ErrNotFound = errors.New("action not found")
)

var (
	registryLock = new(sync.Mutex)
	registry     = make(map[string]Constructor)
)

// Register registers a constructor of a kind of probes.
func Register(name string, ctor Constructor) {
	registryLock.Lock()
	defer registryLock.Unlock()

	if _, ok := registry[name]; ok {
		panic("duplicate action entry: " + name)
	}

	registry[name] = ctor
}

// Construct constructs a named action.
// This function is used internally in cr-monitor.
func Construct(name string, params map[string]interface{}) (Actor, error) {
	registryLock.Lock()
	ctor, ok := registry[name]
	registryLock.Unlock()

	if !ok {
		return nil, ErrNotFound
	}

	return ctor(params)
}
