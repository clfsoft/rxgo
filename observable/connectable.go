package observable

import (
	"context"
	"sync"
)

// A ConnectableObservable is an Observable that only subscribes the source
// Observable by calling its Connect method. Calling its Subscribe method
// will not subscribe the source, instead, it subscribes a local Subject,
// which means that its can be called many times with different Observers.
type ConnectableObservable struct {
	*connectableNoCopy
}

type connectableNoCopy struct {
	mu             sync.Mutex
	source         Observable
	subjectFactory func() Subject
	connection     context.Context
	disconnect     context.CancelFunc
	subject        Subject
	refCount       int
}

func (o *connectableNoCopy) getSubject() Subject {
	if o.subject == nil {
		o.subject = o.subjectFactory()
	}
	return o.subject
}

func (o *connectableNoCopy) doConnect(addRef bool) (context.Context, context.CancelFunc) {
	var try *cancellableLocker

	o.mu.Lock()

	defer func() {
		if try != nil {
			try.Lock()
		}

		o.mu.Unlock()

		if try != nil {
			try.CancelAndUnlock()
		}
	}()

	connection := o.connection

	if connection == nil {
		try = &cancellableLocker{}

		subject := o.getSubject()

		ctx, cancel := o.source.Subscribe(context.Background(), ObserverFunc(func(t Notification) {
			if t.HasValue {
				subject.Next(t.Value)
				return
			}

			tryLocked := try.Lock()

			if !tryLocked {
				o.mu.Lock()
			}

			if connection == o.connection {
				o.connection = nil
				o.disconnect = nil
				o.subject = nil
				o.refCount = 0
			}

			if tryLocked {
				try.Unlock()
			} else {
				o.mu.Unlock()
			}

			t.Observe(subject)
		}))

		select {
		case <-ctx.Done():
			return ctx, cancel
		default:
		}

		connection = ctx
		o.connection = ctx
		o.disconnect = cancel
	}

	if addRef {
		o.refCount++

		return connection, func() {
			o.mu.Lock()
			defer o.mu.Unlock()

			if connection != o.connection {
				return
			}
			if o.refCount == 0 {
				return
			}

			o.refCount--

			if o.refCount == 0 {
				o.disconnect()
				o.connection = nil
				o.disconnect = nil
				o.subject = nil
			}
		}
	}

	return connection, func() {
		o.mu.Lock()
		defer o.mu.Unlock()

		if connection != o.connection {
			return
		}

		o.disconnect()
		o.connection = nil
		o.disconnect = nil
		o.subject = nil
		o.refCount = 0
	}
}

func (o *connectableNoCopy) connectAddRef() (context.Context, context.CancelFunc) {
	return o.doConnect(true)
}

// Connect invokes an execution of an ConnectableObservable.
func (o ConnectableObservable) Connect() (context.Context, context.CancelFunc) {
	return o.doConnect(false)
}

// Subscribe subscribes a local Subject, which is used to multicast to many Observers.
func (o ConnectableObservable) Subscribe(ctx context.Context, ob Observer) (context.Context, context.CancelFunc) {
	o.mu.Lock()
	subject := o.getSubject()
	o.mu.Unlock()
	return subject.Subscribe(ctx, ob)
}

type refCountOperator struct {
	connectable ConnectableObservable
}

func (op refCountOperator) Call(ctx context.Context, ob Observer) (context.Context, context.CancelFunc) {
	ctx, cancel := op.connectable.Subscribe(ctx, ob)
	_, releaseRef := op.connectable.connectAddRef()

	go func() {
		<-ctx.Done()
		releaseRef()
	}()

	return ctx, cancel
}

// RefCount creates an Observable that keeps track of how many subscribers
// it has. When the number of subscribers increases from 0 to 1, it will call
// Connect() for us, which starts the shared execution. Only when the number
// of subscribers decreases from 1 to 0 will it be fully unsubscribed, stopping
// further execution.
func (o ConnectableObservable) RefCount() Observable {
	op := refCountOperator{o}
	return Observable{op}
}

// Publish returns a ConnectableObservable, which is a variety of Observable
// that waits until its Connect method is called before it begins emitting
// items to those Observers that have subscribed to it.
func (o Observable) Publish() ConnectableObservable {
	subject := NewSubject()
	return ConnectableObservable{&connectableNoCopy{
		source:         o,
		subjectFactory: func() Subject { return subject },
	}}
}

// PublishBehavior is like Publish, but it uses a BehaviorSubject instead.
func (o Observable) PublishBehavior(val interface{}) ConnectableObservable {
	subject := NewBehaviorSubject(val)
	return ConnectableObservable{&connectableNoCopy{
		source:         o,
		subjectFactory: func() Subject { return subject },
	}}
}

// Share returns a shared Observable that, when subscribed multiple times, it
// won't subscribe the source Observable twice before the previous subscription
// finishes.
func (o Observable) Share() Observable {
	connectable := ConnectableObservable{&connectableNoCopy{
		source:         o,
		subjectFactory: NewSubject,
	}}
	return connectable.RefCount()
}
