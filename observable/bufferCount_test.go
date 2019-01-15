package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_BufferCount(t *testing.T) {
	subscribe(
		t,
		[]Observable{
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(operators.BufferCount(2), toString),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(operators.BufferCount(3), toString),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(BufferCountOperator{3, 1}.MakeFunc(), toString),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(BufferCountOperator{3, 2}.MakeFunc(), toString),
			Just("A", "B", "C", "D", "E", "F", "G").Pipe(BufferCountOperator{3, 4}.MakeFunc(), toString),
		},
		"[A B]", "[C D]", "[E F]", "[G]", xComplete,
		"[A B C]", "[D E F]", "[G]", xComplete,
		"[A B C]", "[B C D]", "[C D E]", "[D E F]", "[E F G]", "[F G]", "[G]", xComplete,
		"[A B C]", "[C D E]", "[E F G]", "[G]", xComplete,
		"[A B C]", "[E F G]", xComplete,
	)
}
