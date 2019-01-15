package observable

import (
	"context"
)

// DistinctConfig is the configuration type for Distinct.
type DistinctConfig struct {
	KeySelector func(interface{}) interface{}
}

// MakeFunc creates an OperatorFunc from this type.
func (conf DistinctConfig) MakeFunc() OperatorFunc {
	return MakeFunc(distinctOperator(conf).Call)
}

type distinctOperator DistinctConfig

func (op distinctOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	var keys = make(map[interface{}]struct{})
	return source.Subscribe(ctx, func(t Notification) {
		if t.HasValue {
			key := op.KeySelector(t.Value)
			if _, exists := keys[key]; exists {
				return
			}
			keys[key] = struct{}{}
		}
		sink(t)
	})
}

// Distinct creates an Observable that emits all items emitted by the source
// Observable that are distinct by comparison from previous items.
//
// If a keySelector function is provided, then it will project each value from
// the source observable into a new value that it will check for equality with
// previously projected values. If a keySelector function is not provided, it
// will use each value from the source observable directly with an equality
// check against previous values.
func (Operators) Distinct() OperatorFunc {
	return func(source Observable) Observable {
		op := distinctOperator{defaultKeySelector}
		return source.Lift(op.Call)
	}
}
