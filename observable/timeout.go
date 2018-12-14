package observable

import (
	"context"
	"time"
)

// TimeoutConfig is the configuration type for Timeout.
type TimeoutConfig struct {
	Duration   time.Duration
	Observable Observable
}

// MakeFunc creates an OperatorFunc from this type.
func (conf TimeoutConfig) MakeFunc() OperatorFunc {
	return MakeFunc(timeoutOperator(conf).Call)
}

type timeoutOperator TimeoutConfig

func (op timeoutOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	childCtx, childCancel := context.WithCancel(ctx)

	sink = Finally(sink, cancel)

	var (
		scheduleCancel = nothingToDo

		try cancellableLocker
	)

	doSchedule := func() {
		scheduleCancel()

		_, scheduleCancel = scheduleOnce(childCtx, op.Duration, func() {
			if try.Lock() {
				try.CancelAndUnlock()
				childCancel()
				op.Observable.Subscribe(ctx, sink)
			}
		})
	}

	doSchedule()

	source.Subscribe(childCtx, func(t Notification) {
		if try.Lock() {
			switch {
			case t.HasValue:
				sink(t)
				try.Unlock()
				doSchedule()
			default:
				try.CancelAndUnlock()
				sink(t)
			}
		}
	})

	return ctx, cancel
}

// Timeout creates an Observable that mirrors the source Observable or notify
// of an ErrTimeout if the source does not emit a value in given time span.
func (Operators) Timeout(timeout time.Duration) OperatorFunc {
	return func(source Observable) Observable {
		op := timeoutOperator{timeout, Throw(ErrTimeout)}
		return source.Lift(op.Call)
	}
}
