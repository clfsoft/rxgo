package observable

import (
	"context"
)

type toObservablesOperator struct{}

func (op toObservablesOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

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
				cancel()
			}
		case t.HasError:
			sink(t)
			cancel()
		default:
			sink.Next(observables)
			sink.Complete()
			cancel()
		}
	}

	source.Subscribe(ctx, observer.Notify)

	return ctx, cancel
}

// ToObservables creates an Observable that collects all the Observables the
// source emits, then emits them as an slice of Observable when the source
// completes.
func (o Observable) ToObservables() Observable {
	op := toObservablesOperator{}
	return o.Lift(op.Call)
}
