package observable

import (
	"context"
)

type filterOperator struct {
	predicate func(interface{}, int) bool
}

func (op filterOperator) Call(ctx context.Context, ob Observer, source Observable) (context.Context, context.CancelFunc) {
	var outerIndex = -1
	return source.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			outerIndex++

			if op.predicate(t.Value, outerIndex) {
				ob.Next(t.Value)
			}

		case t.HasError:
			ob.Error(t.Value.(error))

		default:
			ob.Complete()
		}
	})
}

// Filter creates an Observable that filter items emitted by the source
// Observable by only emitting those that satisfy a specified predicate.
func (o Observable) Filter(predicate func(interface{}, int) bool) Observable {
	op := filterOperator{predicate}
	return o.Lift(op.Call)
}
