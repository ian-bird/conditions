package handler

import (
	"errors"
	"os"
)

var breakOnSignal = false

var WarningOutput = os.Stderr

const (
	OnContinue restartType = iota
	OnAbort
	OnMuffleWarning
	useValue
)

type restartType int

// this is a way of creating a homogenous struct where
// we can dynamically select based on the type of the argument,
// without forcing the use of objects.
type singleHandler struct {
	concrete error
	f        func(error)
}

var handlers [][]singleHandler

// run the code under the set of bindings defined.  each binding
// should be for a different error type. The first one that is capable
// of handling is the only handler attempted in the binding.
func Bind(code func(), bindings ...singleHandler) {
	oldVal := handlers
	handlers = append(oldVal, bindings)
	defer func() { handlers = oldVal }()

	code()
}

// create a new handler for a given bind statement that is
// specialized to a specific error type
func Handler[E error](h func(E)) singleHandler {
	var e E
	return singleHandler{
		concrete: e,
		f: func(err error) {
			h(err.(E))
		},
	}
}

// this is the default way of signalling exceptions. it calls a function when nothing
// catches it.
func BaseSignal(e error, whenUncaught func()) {
	// start looking for an appropriate handler
	if e == nil {
		return
	}

	// restore handlers when this exits.
	// we need to keep it restricted during handling of this signal
	// otherwise a handler raising another signal would loop.
	handlerCopy := handlers
	defer func() { handlers = handlerCopy }()

	for i := len(handlers) - 1; i >= 0; i-- {
		// trim the current level out before looing for a handler
		handlers = handlers[0:i]
		for _, handler := range handlerCopy[i] {
			// only one per frame even if the handler declines it
			if errors.As(e, &handler.concrete) {
				handler.f(e)
				break
			}
		}
	}

	whenUncaught()
}
