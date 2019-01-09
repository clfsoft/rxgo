package observable

import (
	"context"
)

type retryWhenOperator struct {
	source   Operator
	notifier func(Observable) Observable
}

func (op retryWhenOperator) Call(ctx context.Context, ob Observer) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	sourceCtx, sourceCancel := canceledCtx, noopFunc
	subject := Subject(nil)
	ob = Normalize(ob)

	var observer Observer

	observer = ObserverFunc(func(t Notification) {
		switch {
		case t.HasValue:
			ob.Next(t.Value)
		case t.HasError:
			if subject == nil {
				subject = NewSubject()
				obsv := op.notifier(subject.AsObservable())
				obsv.Subscribe(ctx, ObserverFunc(func(t Notification) {
					switch {
					case t.HasValue:
						sourceCancel()

						sourceCtx, sourceCancel = context.WithCancel(ctx)
						op.source.Call(sourceCtx, observer)

					case t.HasError:
						ob.Error(t.Value.(error))
						cancel()

					default:
						ob.Complete()
						cancel()
					}
				}))
			}
			subject.Next(t.Value.(error))
		default:
			ob.Complete()
			cancel()
		}
	})

	sourceCtx, sourceCancel = context.WithCancel(ctx)
	op.source.Call(sourceCtx, observer)

	return ctx, cancel
}

// RetryWhen creates an Observable that mirrors the source Observable with the
// exception of an Error. If the source Observable calls Error, this method
// will emit the Throwable that caused the Error to the Observable returned
// from notifier. If that Observable calls Complete or Error then this method
// will call Complete or Error on the child subscription. Otherwise this method
// will resubscribe to the source Observable.
func (o Observable) RetryWhen(notifier func(Observable) Observable) Observable {
	op := retryWhenOperator{
		source:   o.Op,
		notifier: notifier,
	}
	return Observable{op}
}
