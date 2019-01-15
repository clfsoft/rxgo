package observable_test

import (
	"context"
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestCreate(t *testing.T) {
	obs := Create(func(ctx context.Context, sink Observer) (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		sink.Next("A")
		sink.Next("B")
		sink.Complete()
		sink.Next("C")
		return ctx, cancel
	})
	value := 0
	subscribe(
		t,
		[]Observable{
			obs.Pipe(operators.Mutex(), operators.Finally(func() { value++ })),
		},
		"A", "B", xComplete,
	)
	if value != 1 {
		t.Fail()
	}
}
