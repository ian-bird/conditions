package main

import (
	"errors"
	"fmt"

	"github.com/ian-bird/conditions/condition"
	"github.com/ian-bird/conditions/handler"
	"github.com/ian-bird/conditions/restart"
)

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
		// condition.Error(errors.New("err!"))
		return 0
	}, func(e error) int {
		return -1
	})

	fmt.Printf("%v %v\n", j, k)
}
