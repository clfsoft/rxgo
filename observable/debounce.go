package observable

import (
	"context"
)

type debounceOperator struct {
	DurationSelector func(interface{}) Observable
}

func (op debounceOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	scheduleCtx, scheduleCancel := canceledCtx, nothingToDo

	sink = Finally(sink, cancel)

	var (
		latestValue interface{}
		try         cancellableLocker
	)

	doSchedule := func(val interface{}) {
		scheduleCancel()

		scheduleCtx, scheduleCancel = context.WithCancel(ctx)

		var observer Observer
		observer = func(t Notification) {
			observer = NopObserver
			scheduleCancel()
			if try.Lock() {
				if t.HasError {
					try.CancelAndUnlock()
					sink(t)
					return
				}
				defer try.Unlock()
				sink.Next(latestValue)
			}
		}

		obsv := op.DurationSelector(val)
		obsv.Subscribe(scheduleCtx, observer.Notify)
	}

	source.Subscribe(ctx, func(t Notification) {
		if try.Lock() {
			switch {
			case t.HasValue:
				latestValue = t.Value
				try.Unlock()
				doSchedule(t.Value)
			default:
				try.CancelAndUnlock()
				sink(t)
			}
		}
	})

	return ctx, cancel
}

// Debounce creates an Observable that emits a value from the source Observable
// only after a particular time span, determined by another Observable, has
// passed without another source emission.
//
// It's like DebounceTime, but the time span of emission silence is determined
// by a second Observable.
func (Operators) Debounce(durationSelector func(interface{}) Observable) OperatorFunc {
	return func(source Observable) Observable {
		op := debounceOperator{durationSelector}
		return source.Lift(op.Call)
	}
}
