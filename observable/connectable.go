package observable

import (
	"context"
	"sync"
	"time"
)

// A ConnectableObservable is an Observable that only subscribes to the source
// Observable by calling its Connect method. Calling its Subscribe method will
// not subscribe the source, instead, it subscribes to a local Subject, which
// means that it can be called many times with different Observers.
type ConnectableObservable struct {
	*connectableObservable
}

type connectableObservable struct {
	Observable
	mutex          sync.Mutex
	source         Observable
	subjectFactory func() Subject
	connection     context.Context
	disconnect     context.CancelFunc
	subject        Subject
	refCount       int
}

func newConnectableObservable(source Observable, subjectFactory func() Subject) *connectableObservable {
	connectable := connectableObservable{
		source:         source,
		subjectFactory: subjectFactory,
	}
	connectable.Observable = Observable{}.Lift(
		func(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
			return connectable.getSubject().Subscribe(ctx, sink)
		},
	)
	return &connectable
}

func (o *connectableObservable) getSubjectLocked() Subject {
	if o.subject.Observer == nil {
		o.subject = o.subjectFactory()
	}
	return o.subject
}

func (o *connectableObservable) getSubject() Subject {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.getSubjectLocked()
}

func (o *connectableObservable) connect(addRef bool) (context.Context, context.CancelFunc) {
	o.mutex.Lock()

	var try *cancellableLocker
	defer func() {
		if try != nil && try.Lock() {
			defer try.CancelAndUnlock()
		}
		o.mutex.Unlock()
	}()

	connection := o.connection

	if connection == nil {
		try = &cancellableLocker{}

		subject := o.getSubjectLocked()

		ctx, cancel := o.source.Subscribe(context.Background(), func(t Notification) {
			if t.HasValue {
				subject.Next(t.Value)
				return
			}

			tryLocked := try.Lock()

			if !tryLocked {
				o.mutex.Lock()
			}

			if connection == o.connection {
				o.connection = nil
				o.disconnect = nil
				o.subject = Subject{}
				o.refCount = 0
			}

			if tryLocked {
				try.Unlock()
			} else {
				o.mutex.Unlock()
			}

			t.Observe(subject.Observer)
		})

		if isDone(ctx) {
			return Done()
		}

		connection = ctx
		o.connection = ctx
		o.disconnect = cancel
	}

	if addRef {
		o.refCount++

		return connection, func() {
			o.mutex.Lock()
			defer o.mutex.Unlock()

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
				o.subject = Subject{}
			}
		}
	}

	return connection, func() {
		o.mutex.Lock()
		defer o.mutex.Unlock()

		if connection != o.connection {
			return
		}

		o.disconnect()
		o.connection = nil
		o.disconnect = nil
		o.subject = Subject{}
		o.refCount = 0
	}
}

func (o *connectableObservable) connectAddRef() (context.Context, context.CancelFunc) {
	return o.connect(true)
}

// Connect invokes an execution of an ConnectableObservable.
func (o ConnectableObservable) Connect() (context.Context, context.CancelFunc) {
	return o.connect(false)
}

type refCountOperator struct {
	Connectable ConnectableObservable
}

func (op refCountOperator) Call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	ctx, cancel := op.Connectable.Subscribe(ctx, sink)
	_, releaseRef := op.Connectable.connectAddRef()

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
	return Observable{}.Lift(op.Call)
}

// Multicast returns a ConnectableObservable, which is a variety of Observable
// that waits until its Connect method is called before it begins emitting
// items to those Observers that have subscribed to it.
func (o Observable) Multicast(subjectFactory func() Subject) ConnectableObservable {
	return ConnectableObservable{newConnectableObservable(o, subjectFactory)}
}

// Publish is like Multicast, but it uses only one subject.
func (o Observable) Publish() ConnectableObservable {
	subject := NewSubject()
	return o.Multicast(func() Subject { return subject })
}

// PublishBehavior is like Publish, but it uses a BehaviorSubject instead.
func (o Observable) PublishBehavior(val interface{}) ConnectableObservable {
	subject := NewBehaviorSubject(val)
	return o.Multicast(func() Subject { return subject.Subject })
}

// PublishReplay is like Publish, but it uses a ReplaySubject instead.
func (o Observable) PublishReplay(bufferSize int, windowTime time.Duration) ConnectableObservable {
	subject := NewReplaySubject(bufferSize, windowTime)
	return o.Multicast(func() Subject { return subject.Subject })
}

// Share returns a new Observable that multicasts (shares) the original
// Observable. When subscribed multiple times, it guarantees that only one
// subscription is made to the source Observable at the same time. When all
// subscribers have unsubscribed it will unsubscribe from the source Observable.
func (Operators) Share() OperatorFunc {
	return func(source Observable) Observable {
		return source.Multicast(NewSubject).RefCount()
	}
}

// ShareReplay is like Share, but it uses a ReplaySubject instead.
func (Operators) ShareReplay(bufferSize int, windowTime time.Duration) OperatorFunc {
	return func(source Observable) Observable {
		newSubject := func() Subject {
			return NewReplaySubject(bufferSize, windowTime).Subject
		}
		return source.Multicast(newSubject).RefCount()
	}
}
