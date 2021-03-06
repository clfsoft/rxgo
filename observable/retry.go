package observable

import (
	"context"
)

type retryOperator struct {
	Count int
}

func (op retryOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	sink = Finally(sink, cancel)

	var (
		count          = op.Count
		observer       Observer
		avoidRecursive avoidRecursiveCalls
	)

	subscribe := func() {
		source.Subscribe(ctx, observer)
	}

	observer = func(t Notification) {
		switch {
		case t.HasValue:
			sink(t)
		case t.HasError:
			if count == 0 {
				sink(t)
			} else {
				if count > 0 {
					count--
				}
				avoidRecursive.Do(subscribe)
			}
		default:
			sink(t)
		}
	}

	avoidRecursive.Do(subscribe)

	return ctx, cancel
}

// Retry creates an Observable that mirrors the source Observable with the
// exception of ERROR emission. If the source Observable errors, this
// operator will resubscribe to the source Observable for a maximum of count
// resubscriptions rather than propagating the ERROR emission.
func (Operators) Retry(count int) OperatorFunc {
	return func(source Observable) Observable {
		if count == 0 {
			return source
		}
		op := retryOperator{count}
		return source.Lift(op.Call)
	}
}
