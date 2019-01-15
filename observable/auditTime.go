package observable

import (
	"context"
	"time"
)

type auditTimeOperator struct {
	Duration time.Duration
}

func (op auditTimeOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	scheduleCtx, _ := Done()

	sink = Finally(sink, cancel)

	var (
		latestValue interface{}
		try         cancellableLocker
	)

	doSchedule := func() {
		if !isDone(scheduleCtx) {
			return
		}

		scheduleCtx, _ = scheduleOnce(ctx, op.Duration, func() {
			if try.Lock() {
				sink.Next(latestValue)
				try.Unlock()
			}
		})
	}

	source.Subscribe(ctx, func(t Notification) {
		if try.Lock() {
			switch {
			case t.HasValue:
				latestValue = t.Value
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

// AuditTime ignores source values for a duration, then emits the most recent
// value from the source Observable, then repeats this process.
//
// When it sees a source values, it ignores that plus the next ones for a
// duration, and then it emits the most recent value from the source.
func (Operators) AuditTime(duration time.Duration) OperatorFunc {
	return func(source Observable) Observable {
		op := auditTimeOperator{duration}
		return source.Lift(op.Call)
	}
}
