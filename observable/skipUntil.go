package observable

import (
	"context"
	"sync/atomic"
)

type skipUntilOperator struct {
	notifier Observable
}

func (op skipUntilOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	done := ctx.Done()

	var (
		noSkipping   uint32
		hasCompleted uint32
	)

	op.notifier.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			atomic.StoreUint32(&noSkipping, 1)
		case t.HasError:
			sink(t)
			cancel()
		default:
			if atomic.CompareAndSwapUint32(&hasCompleted, 0, 1) {
				break
			}
			sink(t)
			cancel()
		}
	})

	select {
	case <-done:
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
			cancel()
		default:
			if atomic.CompareAndSwapUint32(&hasCompleted, 0, 1) {
				break
			}
			sink(t)
			cancel()
		}
	})

	return ctx, cancel
}

// SkipUntil creates an Observable that skips items emitted by the source
// Observable until a second Observable emits an item.
func (o Observable) SkipUntil(notifier Observable) Observable {
	op := skipUntilOperator{notifier}
	return o.Lift(op.Call).Mutex()
}
