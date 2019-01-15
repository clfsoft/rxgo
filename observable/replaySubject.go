package observable

import (
	"context"
	"time"

	"github.com/b97tsk/rxgo/queue"
)

// A ReplaySubject buffers a set number of values and will emit those values
// immediately to any new subscribers in addition to emitting new values to
// existing subscribers.
type ReplaySubject struct {
	*replaySubject
}

// NewReplaySubject creates a new ReplaySubject.
func NewReplaySubject(bufferSize int, windowTime time.Duration) ReplaySubject {
	s := &replaySubject{
		BufferSize: bufferSize,
		WindowTime: windowTime,
	}
	s.Subject = Subject{
		Observable: Observable{}.Lift(s.call),
		Observer:   s.notify,
	}
	return ReplaySubject{s}
}

type replaySubject struct {
	Subject
	try        cancellableLocker
	observers  []*Observer
	err        error
	buffer     queue.Queue
	BufferSize int
	WindowTime time.Duration
}

type replaySubjectValue struct {
	Deadline time.Time
	Value    interface{}
}

func (s *replaySubject) trimBuffer() {
	if s.BufferSize > 0 {
		for s.buffer.Len() > s.BufferSize {
			s.buffer.PopFront()
		}
	}
	if s.WindowTime > 0 {
		now := time.Now()
		for s.buffer.Len() > 0 {
			if s.buffer.Front().(replaySubjectValue).Deadline.After(now) {
				break
			}
			s.buffer.PopFront()
		}
	}
}

func (s *replaySubject) notify(t Notification) {
	if s.try.Lock() {
		switch {
		case t.HasValue:
			var deadline time.Time
			if s.WindowTime > 0 {
				deadline = time.Now().Add(s.WindowTime)
			}
			s.buffer.PushBack(replaySubjectValue{deadline, t.Value})
			s.trimBuffer()

			for _, sink := range s.observers {
				sink.Notify(t)
			}

			s.try.Unlock()

		case t.HasError:
			observers := s.observers
			s.observers = nil
			s.err = t.Value.(error)

			s.try.CancelAndUnlock()

			for _, sink := range observers {
				sink.Notify(t)
			}

		default:
			observers := s.observers
			s.observers = nil

			s.try.CancelAndUnlock()

			for _, sink := range observers {
				sink.Notify(t)
			}
		}
	}
}

func (s *replaySubject) call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
	if s.try.Lock() {
		ctx, cancel := context.WithCancel(ctx)

		observer := Finally(sink, cancel)
		s.observers = append(s.observers, &observer)

		go func() {
			<-ctx.Done()
			if s.try.Lock() {
				for i, sink := range s.observers {
					if sink == &observer {
						copy(s.observers[i:], s.observers[i+1:])
						n := len(s.observers)
						s.observers[n-1] = nil
						s.observers = s.observers[:n-1]
						break
					}
				}
				s.try.Unlock()
			}
		}()

		s.trimBuffer()

		for i, j := 0, s.buffer.Len(); i < j; i++ {
			if isDone(ctx) {
				break
			}
			sink.Next(s.buffer.At(i).(replaySubjectValue).Value)
		}

		s.try.Unlock()
		return ctx, cancel
	}

	if s.err != nil {
		sink.Error(s.err)
		return Done()
	}

	s.trimBuffer()

	for i, j := 0, s.buffer.Len(); i < j; i++ {
		if isDone(ctx) {
			return Done()
		}
		sink.Next(s.buffer.At(i).(replaySubjectValue).Value)
	}
	sink.Complete()
	return Done()
}
