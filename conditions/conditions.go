package condition

import (
	"errors"
	"fmt"
	"os"
)

var handlers [][]func(error)

var breakOnSignal = false

var WarningOutput = os.Stderr

const (
	OnContinue restartType = iota
	OnAbort
	OnMuffleWarning
	useValue
)

type restartType int

type handler struct {
	handlers []func(error)
}

// set up a handler bind that allows for the handling of exceptions.
// a handler should end by calling a restart. If it does not,
// it is said to have declined to handle the function, and the next handler
// attempts.
func HandlerBind() handler {
	return handler{}
}

func (wh handler) Handler(h func(e error)) handler {
	wh.handlers = append(wh.handlers, h)
	return wh
}

// code to run under the given handlers.
func (h handler) Run(code func()) {
	oldVal := handlers
	handlers = append(oldVal, h.handlers)
	defer func() { handlers = oldVal }()

	code()
}

type resume[T any] struct {
	v       T
	restart restartType
}

// wrapper to allow for idiomatic go to fit cleanly into
// the condition system. If the error is non-nil, an error condition
// is generated and attempted to be handled.
func MaybeError[T any](v T, e error) (result T) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		res, ok := r.(resume[T])
		if !ok || res.restart != useValue {
			panic(r)
		}

		result = res.v
	}()

	if e == nil {
		return v
	}

	scopedHandlers := handlers
	defer func() { handlers = scopedHandlers }()

	for i := len(scopedHandlers) - 1; i >= 0; i-- {
		handlers = scopedHandlers[0:i]
		for _, h := range scopedHandlers[i] {
			h(e)
		}
	}

	panic(e)
}

func MaybeBoolError[T any](v T, e bool) (result T) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		res, ok := r.(resume[T])
		if !ok || res.restart != useValue {
			panic(r)
		}

		result = res.v
	}()

	if e == true {
		return v
	}

	scopedHandlers := handlers
	defer func() { handlers = scopedHandlers }()

	for i := len(scopedHandlers) - 1; i >= 0; i-- {
		handlers = scopedHandlers[0:i]
		for _, h := range scopedHandlers[i] {
			h(fmt.Errorf("not ok"))
		}
	}

	panic(e)
}

// raise an error. The most recent handlers are tried first
// and older ones are tried afterwards. Failure to handle
// causes a runtime panic.
func Error[T any](e error) (result T) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		res, ok := r.(resume[T])
		if !ok || res.restart != useValue {
			panic(r)
		}

		result = res.v
	}()

	scopedHandlers := handlers
	defer func() { handlers = scopedHandlers }()

	for i := len(scopedHandlers) - 1; i >= 0; i-- {
		handlers = scopedHandlers[0:i]
		for _, h := range scopedHandlers[i] {
			// we only want to use a handler if its return type
			// matches the one for this error
			func() {
				defer func() {
					r := recover()
					if r == nil {
						return
					}

					_, ok := r.(resume[T])
					if ok {
						panic(r)
					}
				}()
				h(e)
			}()
		}
	}

	panic(e)
}

type Warning struct {
	e error
}

func (w Warning) Unwrap() error {
	return w.e
}

func (w Warning) Error() string {
	return w.e.Error()
}

// warn about something. Caller can decide what to do.
// an unhandled warning is printed to the specified output stream.
func Warn(warn error) {
	defer func() {
		r := recover()
		if r == nil {
			fmt.Fprintf(WarningOutput, "warning: %v\n", warn)
			return
		}

		res, ok := r.(resume[any])
		if !ok {
			panic(r)
		}

		if res.restart == OnMuffleWarning {
			return
		}

		fmt.Fprintf(WarningOutput, "warning: %v\n", warn)
		panic(r)
	}()

	scopedHandlers := handlers
	defer func() { handlers = scopedHandlers }()

	for i := len(scopedHandlers) - 1; i >= 0; i-- {
		handlers = scopedHandlers[0:i]
		for _, h := range scopedHandlers[i] {
			h(Warning{e: warn})
		}
	}

	if breakOnSignal {
		panic(warn)
	}
}

// follow the useValue restart, where the provided value
// is injected in place of the error that occured.
func UseValue[T any](with T) {
	panic(resume[T]{
		v:       with,
		restart: useValue,
	})
}

func MuffleWarning() {
	var t any
	panic(resume[any]{
		v:       t,
		restart: OnMuffleWarning,
	})
}

// unwind the stack to the nearest continue restart case and use it.
func Continue[T any](with T) {
	panic(resume[T]{
		v:       with,
		restart: OnContinue,
	})
}

// unwind the stack to the nearest abort restart case and use it.
func Abort[T any](with T) {
	panic(resume[T]{
		v:       with,
		restart: OnAbort,
	})
}

type restartCase[T any] struct {
	handler func(T) T
	when    restartType
}

// creates a new restart case for WithRestarts that catches
// either a continue or an abort restart.
func Restart[T any](rt restartType, restartCode func(T) T) restartCase[T] {
	return restartCase[T]{
		handler: restartCode,
		when:    rt,
	}
}

// creates a point that can be associated with either the continue
// or the abort restart, or both.
func WithRestarts[T any](code func() T, restarts ...restartCase[T]) (result T) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		if breakOnSignal {
			panic(r)
		}

		res, ok := r.(resume[T])
		if !ok {
			panic(r)
		}

		for _, restart := range restarts {
			if restart.when == res.restart {
				result = restart.handler(res.v)
				return
			}
		}

		panic(r)
	}()

	return code()
}

// within the scope of break on signals,
// all uncaught warnings will panic.
func BreakOnSignals(code func()) {
	breakOnSignal = true
	defer func() { breakOnSignal = false }()
	code()
}

// when an uncaught error occurs, the stack unwinds to the
// handler case statement, and the error is passed to the handler.
func HandlerCase[T any](code func() T, handler func(error) T) T {
	return WithRestarts(func() T {
		var result T

		HandlerBind().Handler(func(e error) {
			Continue(handler(e))
		}).Run(func() {

			result = code()
		})

		return result
	}, Restart(OnContinue, func(t T) T {
		return t
	}))
}

// any errors generated that go unhandled are caught and the zero value is returned.
func IgnoreErrors[T any](code func() T) (result T) {
	return HandlerCase(code, func(err error) T {
		return result
	})
}
