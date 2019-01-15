package observable

import (
	"context"
)

// BufferCountConfig is the configuration type for BufferCount.
type BufferCountConfig struct {
	BufferSize       int
	StartBufferEvery int
}

// MakeFunc creates an OperatorFunc from this type.
func (conf BufferCountConfig) MakeFunc() OperatorFunc {
	return MakeFunc(bufferCountOperator(conf).Call)
}

type bufferCountOperator BufferCountConfig

func (op bufferCountOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	var (
		buffer    = make([]interface{}, 0, op.BufferSize)
		skipCount int
	)
	return source.Subscribe(ctx, func(t Notification) {
		switch {
		case t.HasValue:
			if skipCount > 0 {
				skipCount--
				break
			}
			buffer = append(buffer, t.Value)
			if len(buffer) < op.BufferSize {
				break
			}
			newBuffer := make([]interface{}, 0, op.BufferSize)
			if op.StartBufferEvery < op.BufferSize {
				newBuffer = append(newBuffer, buffer[op.StartBufferEvery:]...)
			} else {
				skipCount = op.StartBufferEvery - op.BufferSize
			}
			sink.Next(buffer)
			buffer = newBuffer
		case t.HasError:
			sink(t)
		default:
			if len(buffer) > 0 {
				for op.StartBufferEvery < len(buffer) {
					newBuffer := append([]interface{}(nil), buffer[op.StartBufferEvery:]...)
					sink.Next(buffer)
					buffer = newBuffer
				}
				sink.Next(buffer)
			}
			sink(t)
		}
	})
}

// BufferCount buffers the source Observable values until the size hits the
// maximum bufferSize given.
//
// BufferCount collects values from the past as an slice, and emits that slice
// only when its size reaches bufferSize.
func (Operators) BufferCount(bufferSize int) OperatorFunc {
	return func(source Observable) Observable {
		op := bufferCountOperator{bufferSize, bufferSize}
		return source.Lift(op.Call)
	}
}
