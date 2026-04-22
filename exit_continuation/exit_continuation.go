package exit_continuation

import "fmt"

type transport struct {
	v any
	n int
}

func (t transport) Error() string {
	return fmt.Sprintf("No handler h#%v in scope for ec call with %v\n", t.n, t.v)
}

var n int = 0

var token chan any

func init() {
	token = make(chan any, 1)
	token <- struct{}{}
}

func nextVal() int {
	var result int
	tk := <-token
	result = n
	n++
	token <- tk
	return result
}

// call the body with an explicit continuation with dynamic extent equivalent to the body.
func CallEc(body func(func(any)) any) (result any) {
	id := nextVal()
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		k, ok := r.(transport)
		if !ok || k.n != id {
			panic(r)
		}

		result = k.v
	}()

	return body(func(v any) {
		panic(transport{
			v: v,
			n: id,
		})
	})
}
