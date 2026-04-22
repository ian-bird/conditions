package restart

import (
	"fmt"
	"maps"
	"reflect"

	"github.com/ian-bird/conditions/exit_continuation"
)

// this is an interface that libraries can use to create additional
// restart types
type RestartT interface {
}

func New() RestartT {
	return struct{}{}
}

type UnwrappableRestart interface {
	Unwrap() RestartT
}

// another box allowing for different restart types to be grouped together
type singleRestart struct {
	r RestartT
	f any
	k func(any)
}

// global map of all restarts for the dynamic scope
var restarts map[reflect.Type][]singleRestart

func init () {
	restarts = make(map[reflect.Type][]singleRestart)
}


// this type is created by restart.Restart
type typedRestart[T any] func(T) singleRestart

// recursively build up let/ecs for each handler.
// inject the continuations into the restarts and extend the context.
// after all have been injected, run the code
//
// if the code exits non-locally, the restart was called and we need
// to actually run the code related to it, and return that.
//
// this effectively creates a dynamic binding for a set of restart functions
// associated with the specific point in the code this appears.
// we're able to substantially reduce the amount of code needed by modelling
// the non-local return with an escape continuation.
func WithRestarts[T any](code func() T, bindings ...typedRestart[T]) T {
	if len(bindings) == 0 {
		return code()
	}

	var t T
	nonLocalExit := true

	result := exit_continuation.CallEc(func(k func(any)) any {
		oldRestarts := maps.Clone(restarts)

		binding := bindings[0](t)
		binding.k = k

		restarts[reflect.TypeOf(t)] = append(restarts[reflect.TypeOf(t)], binding)
		defer func() {
			restarts = oldRestarts
		}()

		toReturn := WithRestarts(code, bindings[1:]...)
		nonLocalExit = false
		return toReturn
	}).(T)

	if !nonLocalExit {
		return result
	}

	return bindings[0](t).f.(func(T) T)(result)
}

// this finds the innermost appropriate handler and calls its continuation with the provided value
func InvokeRestart[T any](r RestartT, t T) {
	var zeroT T
	restartsForT := restarts[reflect.TypeOf(zeroT)]
	for _, restart := range restartsForT {
		if Is(r, restart.r) {
			restart.k(t)
		}
	}

	panic(fmt.Sprintf("failed to find handler for restart of type %T\n", r))
}

// create a new type checked restart
func Restart[T any](r RestartT, code func(T) T) typedRestart[T] {
	return func(_ T) singleRestart {
		return singleRestart{
			r: r,
			f: code,
		}
	}
}

// check if a restart is another type
func Is(unknown, known RestartT) bool {
	if reflect.TypeOf(unknown) == reflect.TypeOf(known) {
		return true
	}

	unwrappable, ok := unknown.(UnwrappableRestart)
	if ok {
		return Is(unwrappable.Unwrap(), known)
	}

	return false
}
