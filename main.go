package main

import (
	"errors"
	"fmt"

	"github.com/ian-bird/conditions/condition"
	"github.com/ian-bird/conditions/handler"
	"github.com/ian-bird/conditions/restart"
)

type divideByZero error

type useValue restart.RestartT

func div(a, b int) int {
	return restart.Case(func() int {
		if b == 0 {
			var e divideByZero = errors.New("divide by 0")
			condition.Error(e)
		}
		return a / b
	}, restart.Restart[int, useValue](func(i int) int { return i }))
}

// demo of default handler behavior. Outer handler declines the signal
// and control passes back here, where we invoke the useValue restart
// that div offers.
func defaultZeroDiv(a, b int) int {
	var result int
	handler.Bind(func() {
		result = div(a, b)
	}, handler.Handler(func(e divideByZero) {
		handler.BaseSignal(e, func() {})
		
		restart.Invoke[int, useValue](0)
	}))
	return result
}

func main() {
	var j int
	handler.Bind(func() {
		j = restart.Case(func() int {
			handler.BaseSignal(errors.New("err!"), func() {})
			return j
		}, restart.Restart[int, restart.RestartT](func(i int) int { return i }))
	}, handler.Handler(func(e error) {
		restart.Invoke[int, restart.RestartT](-1)
	}))

	k := condition.HandlerCase(func() int {
		return 0
	}, func(e error) int {
		return -1
	})

	var l int
	handler.Bind(func() {
		l = defaultZeroDiv(10, 0)
	}, handler.Handler(func(e divideByZero) {
		condition.Warn(e)
	}))

	fmt.Printf("%v %v %v\n", j, k, l)
}
