package observable

import (
	"context"
)

type skipOperator struct {
	Count int
}

func (op skipOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	var (
		count    = op.Count
		observer Observer
	)

	observer = func(t Notification) {
		switch {
		case t.HasValue:
			if count > 1 {
				count--
			} else {
				observer = sink
			}
		default:
			sink(t)
		}
	}

	return source.Subscribe(ctx, observer.Notify)
}

// Skip creates an Observable that skips the first count items emitted by the
// source Observable.
func (o Observable) Skip(count int) Observable {
	if count <= 0 {
		return o
	}
	op := skipOperator{count}
	return o.Lift(op.Call)
}
