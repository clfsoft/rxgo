package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_Single(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just("A", "B").Pipe(operators.Single()),
			Just("A").Pipe(operators.Single()),
			Empty().Pipe(operators.Single()),
			Throw(xErrTest).Pipe(operators.Single()),
		},
		ErrTooMany,
		"A", xComplete,
		ErrEmpty,
		xErrTest,
	)
}
