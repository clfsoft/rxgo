package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_MergeAll(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just(
				Just("A", "B").Pipe(addLatencyToValue(3, 5)),
				Just("C", "D").Pipe(addLatencyToValue(2, 4)),
				Just("E", "F").Pipe(addLatencyToValue(1, 3)),
			).Pipe(operators.MergeAll()),
		},
		"E", "C", "A", "F", "D", "B", xComplete,
	)
}
