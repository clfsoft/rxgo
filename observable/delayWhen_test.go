package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_DelayWhen(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Range(1, 5).Pipe(operators.DelayWhen(
				func(val interface{}, idx int) Observable {
					return Interval(step(val.(int)))
				},
			)),
		},
		1, 2, 3, 4, xComplete,
	)
}
