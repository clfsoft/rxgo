package observable

import (
	"context"

	"github.com/b97tsk/rxgo/atomic"
)

type repeatWhenOperator struct {
	Notifier func(Observable) Observable
}

func (op repeatWhenOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	sink = Mutex(Finally(sink, cancel))

	var sourceLocker *cancellableLocker
	sourceCtx, sourceCancel := Done()

	var (
		activeCount    = atomic.Uint32(2)
		subject        Subject
		createSubject  func() Subject
		avoidRecursive avoidRecursiveCalls
	)

	subscribe := func() {
		var try cancellableLocker
		sourceLocker = &try
		sourceCtx, sourceCancel = context.WithCancel(ctx)
		source.Subscribe(sourceCtx, func(t Notification) {
			if try.Lock() {
				switch {
				case t.HasValue || t.HasError:
					sink(t)
					try.Unlock()
				default:
					activeCount := activeCount.Sub(1)
					try.CancelAndUnlock()
					if activeCount == 0 {
						sink(t)
						break
					}
					if subject.Observer == nil {
						subject = createSubject()
					}
					subject.Next(nil)
				}
			}
		})
	}

	createSubject = func() Subject {
		subject := NewSubject()
		obs := op.Notifier(subject.Observable)
		obs.Subscribe(ctx, func(t Notification) {
			switch {
			case t.HasValue:
				sourceCancel()
				if sourceLocker.Lock() {
					sourceLocker.CancelAndUnlock()
				} else {
					activeCount.Add(1)
				}
				avoidRecursive.Do(subscribe)

			case t.HasError:
				sink(t)

			default:
				if activeCount.Sub(1) == 0 {
					sink(t)
				}
			}
		})
		return subject
	}

	avoidRecursive.Do(subscribe)

	return ctx, cancel
}

// RepeatWhen creates an Observable that mirrors the source Observable with
// the exception of COMPLETE emission. If the source Observable completes,
// this operator will emit nil to the Observable returned from notifier. If
// that Observable emits a value, this operator will resubscribe to the source
// Observable. Otherwise, this operator will emit a COMPLETE on the child
// subscription.
func (Operators) RepeatWhen(notifier func(Observable) Observable) OperatorFunc {
	return func(source Observable) Observable {
		op := repeatWhenOperator{notifier}
		return source.Lift(op.Call)
	}
}
