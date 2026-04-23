package condition

import (
	"errors"
	"fmt"
	"os"

	"github.com/ian-bird/conditions/handler"
	"github.com/ian-bird/conditions/restart"
)

// this is the equivalent to a traditional try/catch. If any code
// signals an error, control jumps to yourHandler and the overall
// result of the HandlerCase is the return from the handler.
func HandlerCase[T any](code func() T, yourHandler func(error) T) T {
	return restart.Case(func() T {
		var result T
		handler.Bind(func() {
			result = code()
		}, handler.Handler(func(e error) {
			restart.Invoke[T, restart.RestartT](yourHandler(e))
		}))
		return result
	}, restart.Restart[T, restart.RestartT](func(t T) T { return t }))
}

func Error(e error) {
	if e == nil {
		panic(errors.New("attempted to throw nil error!"))
	}

	handler.BaseSignal(e, func() {
		panic(fmt.Errorf("%w was unhandled", e))
	})
}

type MuffleWarnings struct{}

var WHERE = os.Stderr

// warn also introduces a muffle-warnings restart case implicitly
func Warn(e error) {
	restart.Case(func() any {
		handler.BaseSignal(e, func() {
			fmt.Fprintf(WHERE, "%v was unhandled", e)
		})
		return struct{}{}
	}, restart.Restart[any, MuffleWarnings](func(a any) any { return a }))

	if e == nil {
		fmt.Fprintf(WHERE, "attempted to raise nil warning!")
	}
}

func IgnoreErrors[T any](code func() T) *T {
	return HandlerCase(func() *T { return new(code()) }, func(error) *T { return nil })
}

type Abort struct{}

type Continue struct{}
