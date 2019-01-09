package observable

import (
	"container/list"
	"context"
)

type congestOperator struct {
	capacity int
}

func (op congestOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	if op.capacity < 65 {
		return op.forSmallCapacity(ctx, sink, source)
	}
	return op.forLargeCapacity(ctx, sink, source)
}

func (op congestOperator) forSmallCapacity(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
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
					sink(t)
				default:
					sink(t)
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

func (op congestOperator) forLargeCapacity(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	done := ctx.Done()

	c := make(chan *list.Element)
	go func() {
		for {
			select {
			case <-done:
				return
			case e := <-c:
				switch t := e.Value.(Notification); {
				case t.HasValue:
					sink(t)
				default:
					sink(t)
					cancel()
					return
				}
			}
		}
	}()

	q := make(chan Notification)
	go func() {
		var queue list.List
		for {
			var (
				in  <-chan Notification
				out chan<- *list.Element
			)
			length := queue.Len()
			if length < op.capacity {
				in = q
			}
			if length > 0 {
				out = c
			}
			select {
			case <-done:
				return
			case t := <-in:
				queue.PushBack(t)
			case out <- queue.Front():
				queue.Remove(queue.Front())
			}
		}
	}()

	source.Subscribe(ctx, func(t Notification) {
		select {
		case <-done:
		case q <- t:
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
