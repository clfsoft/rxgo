package observable

import (
	"context"
)

type congestingZipOperator struct {
	Observables []Observable
}

func (op congestingZipOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	done := ctx.Done()

	sink = Finally(sink, cancel)

	length := len(op.Observables)
	channels := make([]chan Notification, length)

	for i := 0; i < length; i++ {
		channels[i] = make(chan Notification)
	}

	go func() {
		for {
			nextValues := make([]interface{}, length)

			for i := 0; i < length; i++ {
				select {
				case <-done:
					return
				case t := <-channels[i]:
					switch {
					case t.HasValue:
						nextValues[i] = t.Value
					default:
						sink(t)
						return
					}
				}
			}

			sink.Next(nextValues)
		}
	}()

	for index, obs := range op.Observables {
		c := channels[index]
		go obs.Subscribe(ctx, func(t Notification) {
			select {
			case <-done:
			case c <- t:
			}
		})
	}

	return ctx, cancel
}

// CongestingZip combines multiple Observables to create an Observable that
// emits the values of each of its input Observables as an slice.
//
// It's like Zip, but it congests subscribed Observables.
func CongestingZip(observables ...Observable) Observable {
	if len(observables) == 0 {
		return Empty()
	}
	op := congestingZipOperator{observables}
	return Observable{}.Lift(op.Call)
}

// CongestingZipAll converts a higher-order Observable into a first-order
// Observable by waiting for the outer Observable to complete, then applying
// CongestingZip.
//
// CongestingZipAll flattens an Observable-of-Observables by applying
// CongestingZip when the Observable-of-Observables completes.
//
// It's like ZipAll, but it congests subscribed Observables.
func (Operators) CongestingZipAll() OperatorFunc {
	return func(source Observable) Observable {
		op := toObservablesOperator{CongestingZip}
		return source.Lift(op.Call)
	}
}
