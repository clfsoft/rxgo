package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_SampleTime(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(
				addLatencyToValue(1, 2),
				operators.SampleTime(step(4)),
			),
		},
		"B", "D", "F", xComplete,
	)
}
