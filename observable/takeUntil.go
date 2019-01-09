package observable

import (
	"context"
)

type takeUntilOperator struct {
	notifier Observable
}

func (op takeUntilOperator) Call(ctx context.Context, ob Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	done := ctx.Done()

	op.notifier.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			ob.Complete()
		default:
			t.Observe(ob)
		}
		cancel()
	})

	select {
	case <-done:
		return ctx, cancel
	default:
	}

	source.Subscribe(ctx, withFinalizer(ob, cancel))

	return ctx, cancel
}

// TakeUntil creates an Observable that emits the values emitted by the source
// Observable until a notifier Observable emits a value.
//
// TakeUntil lets values pass until a second Observable, notifier, emits
// something. Then, it completes.
func (o Observable) TakeUntil(notifier Observable) Observable {
	op := takeUntilOperator{notifier}
	return o.Lift(op.Call).Mutex()
}
