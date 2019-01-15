package observable

import (
	"context"
	"time"
)

// ThrottleTimeConfig is the configuration type for ThrottleTime.
type ThrottleTimeConfig struct {
	Duration time.Duration
	Leading  bool
	Trailing bool
}

// MakeFunc creates an OperatorFunc from this type.
func (conf ThrottleTimeConfig) MakeFunc() OperatorFunc {
	return MakeFunc(throttleTimeOperator(conf).Call)
}

type throttleTimeOperator ThrottleTimeConfig

func (op throttleTimeOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	throttleCtx, _ := Done()

	sink = Finally(sink, cancel)

	var (
		trailingValue    interface{}
		hasTrailingValue bool

		try cancellableLocker
	)

	var doThrottle func()

	doThrottle = func() {
		throttleCtx, _ = scheduleOnce(ctx, op.Duration, func() {
			if op.Trailing {
				if try.Lock() {
					if hasTrailingValue {
						sink.Next(trailingValue)
						hasTrailingValue = false
						doThrottle()
					}
					try.Unlock()
				}
			}
		})
	}

	source.Subscribe(ctx, func(t Notification) {
		if try.Lock() {
			switch {
			case t.HasValue:
				trailingValue = t.Value
				hasTrailingValue = true
				if isDone(throttleCtx) {
					doThrottle()
					if op.Leading {
						sink(t)
						hasTrailingValue = false
					}
				}
				try.Unlock()

			default:
				try.CancelAndUnlock()
				if hasTrailingValue {
					sink.Next(trailingValue)
				}
				sink(t)
			}
		}
	})

	return ctx, cancel
}

// ThrottleTime creates an Observable that emits a value from the source
// Observable, then ignores subsequent source values for a duration, then
// repeats this process.
//
// ThrottleTime lets a value pass, then ignores source values for the next
// duration time.
func (Operators) ThrottleTime(duration time.Duration) OperatorFunc {
	return func(source Observable) Observable {
		op := throttleTimeOperator{
			Duration: duration,
			Leading:  true,
			Trailing: false,
		}
		return source.Lift(op.Call)
	}
}
