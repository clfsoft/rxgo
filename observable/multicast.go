package observable

import (
	"context"
)

type multicastOperator struct {
	source         Operator
	subjectFactory func() SubjectLike
	selector       func(context.Context, SubjectLike) Observable
}

func (op multicastOperator) Call(ctx context.Context, ob Observer) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	subject := op.subjectFactory()
	obsv := op.selector(ctx, subject)
	obsv.Subscribe(ctx, withFinalizer(ob, cancel))
	select {
	case <-ctx.Done():
	default:
		op.source.Call(ctx, subject)
	}
	return ctx, cancel
}

// Multicast creates an Observable that calls subjectFactory function to create
// a Subject, passes this Subject to selector function, where you have a chance
// to make more subscriptions to that Subject, then subscribes the source
// Observable to emit values to that Subject. Moreover, the selector function
// returns an Observable which is subscribed to the Observer, if it completes
// or emits an error, all subscriptions shall be canceled.
func (o Observable) Multicast(subjectFactory func() SubjectLike, selector func(context.Context, SubjectLike) Observable) Observable {
	op := multicastOperator{
		source:         o.Op,
		subjectFactory: subjectFactory,
		selector:       selector,
	}
	return Observable{op}
}