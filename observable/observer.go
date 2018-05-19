package observable

// An Observer is a consumer of notifications delivered by an Observable.
type Observer func(Notification)

// Next delivers a Next notification to this Observer.
func (sink Observer) Next(val interface{}) {
	sink(Notification{Value: val, HasValue: true})
}

// Error delivers an Error notification to this Observer.
func (sink Observer) Error(err error) {
	sink(Notification{Value: err, HasError: true})
}

// Complete delivers a Complete notification to this Observer.
func (sink Observer) Complete() {
	sink(Notification{})
}

// Notify delivers a notification to this Observer.
func (sink *Observer) Notify(t Notification) {
	(*sink)(t)
}

// NopObserver is an Observer that does nothing.
var NopObserver Observer = func(Notification) {}

func withFinalizer(sink Observer, finalize func()) Observer {
	return func(t Notification) {
		sink(t)
		if t.HasValue {
			return
		}
		finalize()
	}
}
