package observable

import (
	"context"
)

// MapTo creates an Observable that emits the given constant value on the
// output Observable every time the source Observable emits a value.
//
// It's like Map, but it maps every source value to the same output value
// every time.
func (Operators) MapTo(value interface{}) OperatorFunc {
	return func(source Observable) Observable {
		return source.Lift(
			func(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
				return source.Subscribe(ctx, func(t Notification) {
					switch {
					case t.HasValue:
						sink.Next(value)
					default:
						sink(t)
					}
				})
			},
		)
	}
}
