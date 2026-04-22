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
		}, restart.Restart(restart.New(), func(i int) int { return i }))
	}, condition.Handler(func(e error) {
		restart.InvokeRestart(restart.New(), -1)
	}))

	fmt.Printf("%v\n", j)
}
