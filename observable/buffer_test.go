package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_Buffer(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(
				addLatencyToValue(1, 2),
				operators.Buffer(Interval(step(2))),
				toString,
			),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(
				addLatencyToValue(1, 2),
				operators.Buffer(Interval(step(4))),
				toString,
			),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(
				addLatencyToValue(1, 2),
				operators.Buffer(Interval(step(6))),
				toString,
			),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(
				addLatencyToValue(1, 2),
				operators.Buffer(Interval(step(8))),
				toString,
			),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(
				addLatencyToValue(1, 2),
				operators.Buffer(Throw(xErrTest)),
				toString,
			),
		},
		"[A]", "[B]", "[C]", "[D]", "[E]", "[F]", xComplete,
		"[A B]", "[C D]", "[E F]", xComplete,
		"[A B C]", "[D E F]", xComplete,
		"[A B C D]", xComplete,
		xErrTest,
	)
}
