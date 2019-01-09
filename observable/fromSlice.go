package observable

import (
	"context"
)

type fromSliceOperator struct {
	Slice []interface{}
}

func (op fromSliceOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	done := ctx.Done()
	for _, val := range op.Slice {
		select {
		case <-done:
			return canceledCtx, doNothing
		default:
		}
		sink.Next(val)
	}
	sink.Complete()
	return canceledCtx, doNothing
}

// FromSlice creates an Observable that emits values from an slice, one after
// the other, and then completes.
func FromSlice(slice []interface{}) Observable {
	if len(slice) == 0 {
		return Empty()
	}
	op := fromSliceOperator{slice}
	return Observable{}.Lift(op.Call)
}

// Just creates an Observable that emits some values you specify as arguments,
// one after the other, and then completes.
func Just(slice ...interface{}) Observable {
	return FromSlice(slice)
}
