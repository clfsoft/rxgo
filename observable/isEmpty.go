package observable

import (
	"context"
)

type isEmptyOperator struct{}

func (op isEmptyOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	sink = Finally(sink, cancel)

	var observer Observer
	observer = func(t Notification) {
		switch {
		case t.HasValue:
			observer = NopObserver
			sink.Next(false)
			sink.Complete()
		case t.HasError:
			sink(t)
		default:
			sink.Next(true)
			sink.Complete()
		}
	}
	source.Subscribe(ctx, observer.Notify)

	return ctx, cancel
}

// IsEmpty creates an Observable that emits true if the source Observable
// emits no items, otherwise, it emits false.
func (Operators) IsEmpty() OperatorFunc {
	return func(source Observable) Observable {
		op := isEmptyOperator{}
		return source.Lift(op.Call)
	}
}
