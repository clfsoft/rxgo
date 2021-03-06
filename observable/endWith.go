package observable

// EndWith creates an Observable that emits the items you specify as arguments
// after it finishes emitting items emitted by the source Observable.
func (Operators) EndWith(values ...interface{}) OperatorFunc {
	return func(source Observable) Observable {
		if len(values) == 0 {
			return source
		}
		return Concat(source, FromSlice(values))
	}
}
