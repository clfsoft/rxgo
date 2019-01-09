package observable

import (
	"context"
)

type finallyOperator struct {
	Func func()
}

func (op finallyOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	return source.Subscribe(ctx, Finally(sink, op.Func))
}

// Finally creates an Observer that passes all emissions to the specified
// Observer, in the case that an Error or Complete emission is passed, makes
// a call to the specified function.
func Finally(sink Observer, finally func()) Observer {
	return func(t Notification) {
		if t.HasValue {
			sink(t)
			return
		}
		defer finally()
		sink(t)
	}
}

// Finally creates an Observable that mirrors the source Observable, in the
// case that an Error or Complete emission is mirrored, makes a call to the
// specified function.
func (Operators) Finally(finally func()) OperatorFunc {
	return func(source Observable) Observable {
		op := finallyOperator{finally}
		return source.Lift(op.Call)
	}
}
