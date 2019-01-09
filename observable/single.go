package observable

import (
	"context"
)

type singleOperator struct{}

func (op singleOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	var (
		value    interface{}
		hasValue bool
		observer Observer
	)

	observer = func(t Notification) {
		switch {
		case t.HasValue:
			if hasValue {
				observer = NopObserver
				sink.Error(ErrTooMany)
				cancel()
			} else {
				value = t.Value
				hasValue = true
			}
		case t.HasError:
			sink(t)
			cancel()
		default:
			if hasValue {
				sink.Next(value)
				sink.Complete()
			} else {
				sink.Error(ErrEmpty)
			}
			cancel()
		}
	}

	source.Subscribe(ctx, observer.Notify)

	return ctx, cancel
}

// Single creates an Observable that emits the single item emitted by the
// source Observable. If the source emits more than one item or no items,
// notify of an ErrTooMany or ErrEmpty respectively.
func (Operators) Single() OperatorFunc {
	return func(source Observable) Observable {
		op := singleOperator{}
		return source.Lift(op.Call)
	}
}
