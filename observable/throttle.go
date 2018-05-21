package observable

import (
	"context"
)

type throttleOperator struct {
	DurationSelector func(interface{}) Observable
}

func (op throttleOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	scheduleCtx, scheduleCancel := canceledCtx, doNothing
	scheduleDone := scheduleCtx.Done()

	source.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			select {
			case <-scheduleDone:
			default:
				return
			}

			sink(t)

			scheduleCtx, scheduleCancel = context.WithCancel(ctx)
			scheduleDone = scheduleCtx.Done()

			var observer Observer
			observer = func(t Notification) {
				observer = NopObserver
				scheduleCancel()
				if t.HasError {
					sink(t)
					cancel()
					return
				}
			}

			obsv := op.DurationSelector(t.Value)
			obsv.Subscribe(scheduleCtx, observer.Notify)

		default:
			sink(t)
			cancel()
		}
	})

	return ctx, cancel
}

// Throttle creates an Observable that emits a value from the source
// Observable, then ignores subsequent source values for a duration determined
// by another Observable, then repeats this process.
//
// It's like ThrottleTime, but the silencing duration is determined by a second
// Observable.
func (o Observable) Throttle(durationSelector func(interface{}) Observable) Observable {
	op := throttleOperator{durationSelector}
	return o.Lift(op.Call)
}
