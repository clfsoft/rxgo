package observable

import (
	"context"
)

type congestOperator struct {
	capacity int
}

func (op congestOperator) Call(ctx context.Context, ob Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	done := ctx.Done()
	c := make(chan Notification, op.capacity)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-c:
				switch {
				case t.HasValue:
					t.Observe(ob)
				default:
					t.Observe(ob)
					cancel()
					return
				}
			}
		}
	}()

	source.Subscribe(ctx, func(t Notification) {
		select {
		case <-done:
		case c <- t:
		}
	})

	return ctx, cancel
}

// Congest creates an Observable that mirrors the source Observable, caches
// emissions if the source emits too fast, and congests the source if the cache
// is full.
func (o Observable) Congest(capacity int) Observable {
	if capacity < 1 {
		return o
	}
	op := congestOperator{capacity}
	return o.Lift(op.Call)
}
