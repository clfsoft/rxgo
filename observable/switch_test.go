package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_Switch(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just(
				Just("A", "B", "C", "D").Pipe(addLatencyToValue(0, 2)),
				Just("E", "F", "G", "H").Pipe(addLatencyToValue(0, 3)),
				Just("I", "J", "K", "L").Pipe(addLatencyToValue(0, 2)),
			).Pipe(addLatencyToValue(0, 5), operators.Switch()),
			Just(
				Just("A", "B", "C", "D").Pipe(addLatencyToValue(0, 2)),
				Just("E", "F", "G", "H").Pipe(addLatencyToValue(0, 3)),
				Just("I", "J", "K", "L").Pipe(addLatencyToValue(0, 2)),
				Empty(),
			).Pipe(addLatencyToValue(0, 5), operators.Switch()),
			Just(
				Just("A", "B", "C", "D").Pipe(addLatencyToValue(0, 2)),
				Just("E", "F", "G", "H").Pipe(addLatencyToValue(0, 3)),
				Just("I", "J", "K", "L").Pipe(addLatencyToValue(0, 2)),
				Throw(xErrTest),
			).Pipe(addLatencyToValue(0, 5), operators.Switch()),
		},
		"A", "B", "C", "E", "F", "I", "J", "K", "L", xComplete,
		"A", "B", "C", "E", "F", "I", "J", "K", xComplete,
		"A", "B", "C", "E", "F", "I", "J", "K", xErrTest,
	)
}
