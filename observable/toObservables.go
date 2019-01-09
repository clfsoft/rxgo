package observable

import (
	"context"
)

type toObservablesOperator struct{}

func (op toObservablesOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	sink = Finally(sink, cancel)

	var (
		observables []Observable
		observer    Observer
	)

	observer = func(t Notification) {
		switch {
		case t.HasValue:
			if obsv, ok := t.Value.(Observable); ok {
				observables = append(observables, obsv)
			} else {
				observer = NopObserver
				sink.Error(ErrNotObservable)
			}
		case t.HasError:
			sink(t)
		default:
			sink.Next(observables)
			sink.Complete()
		}
	}

	source.Subscribe(ctx, observer.Notify)

	return ctx, cancel
}

// ToObservables creates an Observable that collects all the Observables the
// source emits, then emits them as an slice of Observable when the source
// completes.
func (Operators) ToObservables() OperatorFunc {
	return func(source Observable) Observable {
		op := toObservablesOperator{}
		return source.Lift(op.Call)
	}
}
