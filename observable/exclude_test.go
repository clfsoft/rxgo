package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_Exclude(t *testing.T) {
	excludeLessThan5 := operators.Exclude(
		func(val interface{}, idx int) bool {
			return val.(int) < 5
		},
	)
	subscribe(
		t,
		[]Observable{
			Just(1, 2, 3, 4, 5, 4, 3, 2, 1).Pipe(excludeLessThan5),
			Range(1, 9).Pipe(excludeLessThan5),
			Concat(Range(1, 9), Throw(xErrTest)).Pipe(excludeLessThan5),
		},
		5, xComplete,
		5, 6, 7, 8, xComplete,
		5, 6, 7, 8, xErrTest,
	)
}
