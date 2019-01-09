package observable

import (
	"context"
	"time"
)

type sampleTimeOperator struct {
	interval time.Duration
}

func (op sampleTimeOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	var (
		latestValue    interface{}
		hasLatestValue bool
		try            cancellableLocker
	)

	schedule(ctx, op.interval, func() {
		if try.Lock() {
			defer try.Unlock()
			if hasLatestValue {
				sink.Next(latestValue)
				hasLatestValue = false
			}
		}
	})

	source.Subscribe(ctx, func(t Notification) {
		if try.Lock() {
			switch {
			case t.HasValue:
				latestValue = t.Value
				hasLatestValue = true
				try.Unlock()
			default:
				try.CancelAndUnlock()
				sink(t)
				cancel()
			}
		}
	})

	return ctx, cancel
}

// SampleTime creates an Observable that emits the most recently emitted value
// from the source Observable within periodic time intervals.
func (o Observable) SampleTime(interval time.Duration) Observable {
	op := sampleTimeOperator{interval}
	return o.Lift(op.Call)
}
