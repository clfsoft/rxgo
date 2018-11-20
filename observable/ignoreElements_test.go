package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_IgnoreElements(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Empty().Pipe(operators.IgnoreElements()),
			Just("A", "B", "C").Pipe(operators.IgnoreElements()),
			Concat(Just("A", "B", "C"), Throw(xErrTest)).Pipe(operators.IgnoreElements()),
		},
		xComplete,
		xComplete,
		xErrTest,
	)
}
