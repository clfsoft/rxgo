package observable

import (
	"context"
)

type everyOperator struct {
	Predicate func(interface{}, int) bool
}

func (op everyOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	var (
		outerIndex = -1
		observer   Observer
	)

	observer = func(t Notification) {
		switch {
		case t.HasValue:
			outerIndex++

			if !op.Predicate(t.Value, outerIndex) {
				observer = NopObserver
				sink.Next(false)
				sink.Complete()
				cancel()
			}

		case t.HasError:
			sink(t)
			cancel()

		default:
			sink.Next(true)
			sink.Complete()
			cancel()
		}
	}

	source.Subscribe(ctx, observer.Notify)

	return ctx, cancel
}

// Every creates an Observable that emits whether or not every item of the source
// satisfies the condition specified.
//
// Every emits true or false, then completes.
func (Operators) Every(predicate func(interface{}, int) bool) OperatorFunc {
	return func(source Observable) Observable {
		op := everyOperator{predicate}
		return source.Lift(op.Call)
	}
}
