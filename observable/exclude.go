package observable

import (
	"context"
)

type excludeOperator struct {
	predicate func(interface{}, int) bool
}

func (op excludeOperator) Call(ctx context.Context, ob Observer, source Observable) (context.Context, context.CancelFunc) {
	var outerIndex = -1
	return source.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			outerIndex++

			if !op.predicate(t.Value, outerIndex) {
				ob.Next(t.Value)
			}

		case t.HasError:
			ob.Error(t.Value.(error))

		default:
			ob.Complete()
		}
	})
}

// Exclude creates an Observable that filter items emitted by the source
// Observable by only emitting those that do not satisfy a specified predicate.
func (o Observable) Exclude(predicate func(interface{}, int) bool) Observable {
	op := excludeOperator{predicate}
	return o.Lift(op.Call)
}
