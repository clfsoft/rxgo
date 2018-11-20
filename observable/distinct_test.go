package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_Distinct(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just("A", "B", "B", "A", "C", "C", "A").Pipe(operators.Distinct()),
		},
		"A", "B", "C", xComplete,
	)
}
