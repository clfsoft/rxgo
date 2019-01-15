package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_DebounceTime(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just("A", "B", "C").Pipe(
				addLatencyToValue(1, 2),
				operators.DebounceTime(step(3)),
			),
			Just("A", "B", "C").Pipe(
				addLatencyToValue(1, 3),
				operators.DebounceTime(step(2)),
			),
		},
		"C", xComplete,
		"A", "B", "C", xComplete,
	)
}
