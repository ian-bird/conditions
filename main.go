package main

import (
	"errors"
	"fmt"

	condition "github.com/ian-bird/conditions/conditions"
	"github.com/ian-bird/conditions/restart"
)

func main() {
	var j int
	condition.HandlerBind(func() {
		j = restart.WithRestarts(func() int {
			condition.BaseSignal(errors.New("err!"), func() {})
			return j
		}, restart.Restart[int, restart.RestartT](func(i int) int { return i }))
	}, condition.Handler(func(e error) {
		restart.InvokeRestart[string, restart.RestartT]("a")
	}))

	fmt.Printf("%v\n", j)
}
