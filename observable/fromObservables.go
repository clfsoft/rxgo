package observable

import (
	"context"
)

type fromObservablesOperator struct {
	Observables []Observable
}

func (op fromObservablesOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	done := ctx.Done()
	for _, obs := range op.Observables {
		select {
		case <-done:
			return canceledCtx, nothingToDo
		default:
		}
		sink.Next(obs)
	}
	sink.Complete()
	return canceledCtx, nothingToDo
}

// FromObservables creates an Observable that emits some Observables
// you specify as arguments, one after the other, and then completes.
func FromObservables(observables ...Observable) Observable {
	if len(observables) == 0 {
		return Empty()
	}
	if len(observables) == 1 {
		return just(observables[0])
	}
	op := fromObservablesOperator{observables}
	return Observable{}.Lift(op.Call)
}
