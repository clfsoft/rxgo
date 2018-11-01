package observable

import (
	"context"
	"sync/atomic"
)

type skipUntilOperator struct {
	Notifier Observable
}

func (op skipUntilOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	sink = Mutex(Finally(sink, cancel))

	var (
		noSkipping   uint32
		hasCompleted uint32
	)

	op.Notifier.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			atomic.StoreUint32(&noSkipping, 1)
		case t.HasError:
			sink(t)
		default:
			if atomic.CompareAndSwapUint32(&hasCompleted, 0, 1) {
				break
			}
			sink(t)
		}
	})

	select {
	case <-ctx.Done():
		return ctx, cancel
	default:
	}

	source.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			if atomic.LoadUint32(&noSkipping) != 0 {
				sink(t)
			}
		case t.HasError:
			sink(t)
		default:
			if atomic.CompareAndSwapUint32(&hasCompleted, 0, 1) {
				break
			}
			sink(t)
		}
	})

	return ctx, cancel
}

// SkipUntil creates an Observable that skips items emitted by the source
// Observable until a second Observable emits an item.
func (Operators) SkipUntil(notifier Observable) OperatorFunc {
	return func(source Observable) Observable {
		op := skipUntilOperator{notifier}
		return source.Lift(op.Call)
	}
}
