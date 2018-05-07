package observable

import (
	"context"
	"time"
)

type emptyOperator struct {
	delay     time.Duration
	scheduler Scheduler
}

func (op emptyOperator) Call(ctx context.Context, ob Observer) (context.Context, context.CancelFunc) {
	if op.scheduler != nil {
		return op.scheduler.ScheduleOnce(ctx, op.delay, ob.Complete)
	}

	ob.Complete()
	return canceledCtx, noopFunc
}

// Empty creates an Observable that emits no items to the Observer and
// immediately emits a Complete notification.
func Empty() Observable {
	op := emptyOperator{}
	return Observable{op}
}

// EmptyOn creates an Observable that emits no items to the Observer and
// immediately emits a Complete notification, on the specified Scheduler.
func EmptyOn(s Scheduler, delay time.Duration) Observable {
	op := emptyOperator{
		delay:     delay,
		scheduler: s,
	}
	return Observable{op}
}