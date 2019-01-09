package observable

import (
	"context"
)

type repeatOperator struct {
	count int
}

func (op repeatOperator) Call(ctx context.Context, ob Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	var (
		count    = op.count
		observer Observer
	)

	observer = func(t Notification) {
		switch {
		case t.HasValue:
			ob.Next(t.Value)
		case t.HasError:
			ob.Error(t.Value.(error))
			cancel()
		default:
			if count == 0 {
				ob.Complete()
				cancel()
			} else {
				if count > 0 {
					count--
				}
				source.Subscribe(ctx, observer)
			}
		}
	}

	source.Subscribe(ctx, observer)

	return ctx, cancel
}

// Repeat creates an Observable that repeats the stream of items emitted by the
// source Observable at most count times.
func (o Observable) Repeat(count int) Observable {
	if count == 0 {
		return Empty()
	}
	if count > 0 {
		count--
	}
	op := repeatOperator{count}
	return o.Lift(op.Call)
}
