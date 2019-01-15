package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_Expand(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just(8).Pipe(operators.Expand(
				func(val interface{}, idx int) Observable {
					i := val.(int)
					if i < 1 {
						return Empty()
					}
					return Just(i - 1)
				},
			)),
		},
		8, 7, 6, 5, 4, 3, 2, 1, 0, xComplete,
	)
}
