package observable

import (
	"context"
)

type catchOperator struct {
	selector func(error) Observable
}

func (op catchOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	source.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			sink(t)
		case t.HasError:
			obsv := op.selector(t.Value.(error))
			obsv.Subscribe(ctx, withFinalizer(sink, cancel))
		default:
			sink(t)
			cancel()
		}
	})

	return ctx, cancel
}

// Catch catches errors on the observable to be handled by returning a new
// observable.
func (o Observable) Catch(selector func(error) Observable) Observable {
	op := catchOperator{selector}
	return o.Lift(op.Call)
}
