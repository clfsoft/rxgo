package observable

import (
	"context"
)

type ignoreElementsOperator struct{}

func (op ignoreElementsOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	return source.Subscribe(ctx, IgnoreElements(sink))
}

// IgnoreElements creates an Observer that ignores all values but only passes
// Complete or Error emission to the specified Observer.
func IgnoreElements(sink Observer) Observer {
	return func(t Notification) {
		switch {
		case t.HasValue:
		default:
			sink(t)
		}
	}
}

// IgnoreElements creates an Observable that ignores all values emitted by the
// source Observable and only passes Complete or Error emission.
func (o Observable) IgnoreElements() Observable {
	op := ignoreElementsOperator{}
	return o.Lift(op.Call)
}
