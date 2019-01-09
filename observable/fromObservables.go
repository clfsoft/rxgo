package observable

import (
	"context"
)

type fromObservablesOperator struct {
	Observables []Observable
}

func (op fromObservablesOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	done := ctx.Done()
	for _, obsv := range op.Observables {
		select {
		case <-done:
			return canceledCtx, nothingToDo
		default:
		}
		sink.Next(obsv)
	}
	sink.Complete()
	return canceledCtx, nothingToDo
}

// FromObservables creates an Observable that emits Observables from an slice,
// one after the other, and then completes.
func FromObservables(observables []Observable) Observable {
	if len(observables) == 0 {
		return Empty()
	}
	op := fromObservablesOperator{observables}
	return Observable{}.Lift(op.Call)
}
