package observable

import (
	"context"
)

type takeWhileOperator struct {
	Predicate func(interface{}, int) bool
}

func (op takeWhileOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	var (
		outerIndex = -1
		observer   Observer
	)

	observer = func(t Notification) {
		switch {
		case t.HasValue:
			outerIndex++

			if op.Predicate(t.Value, outerIndex) {
				sink(t)
				break
			}

			observer = NopObserver
			sink.Complete()
			cancel()

		default:
			sink(t)
			cancel()
		}
	}

	source.Subscribe(ctx, observer.Notify)

	return ctx, cancel
}

// TakeWhile creates an Observable that emits values emitted by the source
// Observable so long as each value satisfies the given predicate, and then
// completes as soon as this predicate is not satisfied.
//
// TakeWhile takes values from the source only while they pass the condition
// given. When the first value does not satisfy, it completes.
func (o Observable) TakeWhile(predicate func(interface{}, int) bool) Observable {
	op := takeWhileOperator{predicate}
	return o.Lift(op.Call)
}
