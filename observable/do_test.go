package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestOperators_Do(t *testing.T) {
	n := 0
	op := operators.Do(func(Notification) { n++ })
	obs := Defer(func() Observable { return Just(n) })
	subscribe(
		t,
		[]Observable{
			Concat(Empty().Pipe(op), obs),
			Concat(Just("A").Pipe(op), obs),
			Concat(Just("A", "B").Pipe(op), obs),
			Concat(Concat(Just("A", "B"), Throw(xErrTest)).Pipe(op), obs),
			obs,
		},
		1, xComplete,
		"A", 3, xComplete,
		"A", "B", 6, xComplete,
		"A", "B", xErrTest,
		9, xComplete,
	)
}

func TestOperators_DoOnEach(t *testing.T) {
	n := 0
	op := operators.DoOnEach(func(interface{}) { n++ })
	obs := Defer(func() Observable { return Just(n) })
	subscribe(
		t,
		[]Observable{
			Concat(Empty().Pipe(op), obs),
			Concat(Just("A").Pipe(op), obs),
			Concat(Just("A", "B").Pipe(op), obs),
			Concat(Concat(Just("A", "B"), Throw(xErrTest)).Pipe(op), obs),
			obs,
		},
		0, xComplete,
		"A", 1, xComplete,
		"A", "B", 3, xComplete,
		"A", "B", xErrTest,
		5, xComplete,
	)
}

func TestOperators_DoOnError(t *testing.T) {
	n := 0
	op := operators.DoOnError(func(error) { n++ })
	obs := Defer(func() Observable { return Just(n) })
	subscribe(
		t,
		[]Observable{
			Concat(Empty().Pipe(op), obs),
			Concat(Just("A").Pipe(op), obs),
			Concat(Just("A", "B").Pipe(op), obs),
			Concat(Concat(Just("A", "B"), Throw(xErrTest)).Pipe(op), obs),
			obs,
		},
		0, xComplete,
		"A", 0, xComplete,
		"A", "B", 0, xComplete,
		"A", "B", xErrTest,
		1, xComplete,
	)
}

func TestOperators_DoOnComplete(t *testing.T) {
	n := 0
	op := operators.DoOnComplete(func() { n++ })
	obs := Defer(func() Observable { return Just(n) })
	subscribe(
		t,
		[]Observable{
			Concat(Empty().Pipe(op), obs),
			Concat(Just("A").Pipe(op), obs),
			Concat(Just("A", "B").Pipe(op), obs),
			Concat(Concat(Just("A", "B"), Throw(xErrTest)).Pipe(op), obs),
			obs,
		},
		1, xComplete,
		"A", 2, xComplete,
		"A", "B", 3, xComplete,
		"A", "B", xErrTest,
		3, xComplete,
	)
}
