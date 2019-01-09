package observable

import (
	"context"
)

type doOperator struct {
	sink Observer
}

func (op doOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	return source.Subscribe(ctx, func(t Notification) {
		op.sink(t)
		sink(t)
	})
}

// Do creates an Observable that mirrors the source Observable, but perform
// a side effect before each emission.
func (o Observable) Do(sink Observer) Observable {
	op := doOperator{sink}
	return o.Lift(op.Call)
}
