package observable

import (
	"context"
)

type catchOperator struct {
	Selector func(error) Observable
}

func (op catchOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	sink = Finally(sink, cancel)

	source.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			sink(t)
		case t.HasError:
			obs := op.Selector(t.Value.(error))
			obs.Subscribe(ctx, sink)
		default:
			sink(t)
		}
	})

	return ctx, cancel
}

// Catch catches errors on the observable to be handled by returning a new
// observable.
func (Operators) Catch(selector func(error) Observable) OperatorFunc {
	return func(source Observable) Observable {
		op := catchOperator{selector}
		return source.Lift(op.Call)
	}
}
