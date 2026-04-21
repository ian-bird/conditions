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
	if i%2 != 1 {
		return condition.Error[int](fmt.Errorf("%v is not odd!", i))
	}

	return i
}

func identity[T any](t T) T {
	return t
}

func doSumming(upto int) int {
	var sum int
	for i := range upto {
		// set up a continue point that skips this iteration
		condition.WithRestarts(func() int {

			// sum this number if its odd
			sum += oddNum(i)

			return 0
		}, condition.Restart(condition.OnContinue, identity[int]))
	}

	return sum
}

func summing(upto int) int {
	var result int
	condition.HandlerBind().Handler(func(e error) {
		// we only want to handle odd num errors
		var i int
		c, _ := fmt.Sscanf(e.Error(), "%v is not odd!", &i)
		if c != 1 {
			return
		}

		// see if someone else wants to handle it
		condition.Signal[int](e)

		// if no one does then skip that number
		condition.Continue(0)
	}).Run(func() {
		result = doSumming(upto)
	})

	return result
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

		condition.UseValue(1)
	}).Run(func() {
		a = oddNum(2) * 5
	})

	// show some more advanced use with default handlers
	var b int
	condition.HandlerBind().Handler(func(e error) {
		var i int
		c, _ := fmt.Sscanf(e.Error(), "%v is not odd!", &i)
		if c != 1 {
			return
		}

		// if its divisble by 4 we use the value anyways,
		// otherwise use the default handler.
		if i%4 == 0 {
			condition.UseValue(i)
		}
	}).Run(func() {
		b = summing(10)
	})

	// show warnings going uncaught
	var y int
	y = evenNum(7)

	fmt.Printf("%v %v %v %v %v\n", x, y, z, a, b) // prints 3 7 9 5 37 and warning for 7
}
