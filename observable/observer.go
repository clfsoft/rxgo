package observable

// An Observer is a consumer of notifications delivered by an Observable.
type Observer func(Notification)

// Next delivers a Next notification to this Observer.
func (ob Observer) Next(val interface{}) {
	ob(Notification{Value: val, HasValue: true})
}

// Error delivers an Error notification to this Observer.
func (ob Observer) Error(err error) {
	ob(Notification{Value: err, HasError: true})
}

// Complete delivers a Complete notification to this Observer.
func (ob Observer) Complete() {
	ob(Notification{})
}

// NopObserver is an Observer that does nothing.
var NopObserver Observer = func(Notification) {}

func withFinalizer(ob Observer, finalize func()) Observer {
	return func(t Notification) {
		t.Observe(ob)
		if t.HasValue {
			return
		}
		finalize()
	}
}
