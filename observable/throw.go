package observable

import (
	"context"
)

// Throw creates an Observable that emits no items to the Observer and
// immediately emits an ERROR notification.
func Throw(err error) Observable {
	return Observable{}.Lift(
		func(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
			sink.Error(err)
			return Done()
		},
	)
}
