package main

import (
	"fmt"

	condition "github.com/ian-bird/conditions/conditions"
)

func evenNum(i int) int {
	if i%2 != 0 {
		condition.Warn(fmt.Errorf("%v is not even!", i))
	}

	return i
}

func main() {
	// standard try/catch with dynamic unwinding
	x := condition.HandlerCase(func() int {
		return evenNum(3) * 3
	}, func(err error) int {
		var i int
		c, _ := fmt.Sscanf(err.Error(), "%v is not even!", &i)
		if c != 1 {
			panic(err)
		}

		return i
	})

	// show useValue injecting alternate value into warning
	var z int
	condition.HandlerBind().Handler(func(e error) {
		var i int
		c, _ := fmt.Sscanf(e.Error(), "%v is not even!", &i)
		if c != 1 {
			return
		}

		condition.MuffleWarning()
	}).Run(func() {
		z = evenNum(3) * 3
	})

	// show warnings going uncaught
	var y int
	y = evenNum(7)

	fmt.Printf("%v %v %v\n", x, y, z)
}
