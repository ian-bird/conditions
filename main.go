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

func oddNum(i int) int {
	if i % 2 != 1 {
		return condition.Error[int](fmt.Errorf("%v is not odd!", i))
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

	// show muffled warning
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

	// show useValue overriding erroneous value
	var a int
	condition.HandlerBind().Handler(func(e error) {
		var i int
		c, _ := fmt.Sscanf(e.Error(), "%v is not odd!", &i)
		if c != 1 {
			return
		}

		condition.UseValue[int](1)
	}).Run(func() {
		a = oddNum(2) * 5
	})	

	// show warnings going uncaught
	var y int
	y = evenNum(7)

	fmt.Printf("%v %v %v %v\n", x, y, z, a) // prints 3 7 9 5 and warning for 7
}
