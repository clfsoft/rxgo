package observable

import (
	"container/list"
	"context"
	"time"
)

// A ReplaySubject buffers a set number of values and will emit those values
// immediately to any new subscribers in addition to emitting new values to
// existing subscribers.
type ReplaySubject struct {
	Subject
	try        cancellableLocker
	observers  []*Observer
	err        error
	buffer     list.List
	bufferSize int
	windowTime time.Duration
}

type replaySubjectValue struct {
	Deadline time.Time
	Value    interface{}
}

func (s *ReplaySubject) trimBuffer() {
	if s.bufferSize > 0 {
		for s.buffer.Len() > s.bufferSize {
			s.buffer.Remove(s.buffer.Front())
		}
	}
	if s.windowTime > 0 && s.buffer.Len() > 0 {
		now := time.Now()
		for {
			e := s.buffer.Front()
			if e == nil {
				break
			}
			if e.Value.(replaySubjectValue).Deadline.After(now) {
				break
			}
			s.buffer.Remove(e)
		}
	}
}

func (s *ReplaySubject) notify(t Notification) {
	if s.try.Lock() {
		switch {
		case t.HasValue:
			var deadline time.Time
			if s.windowTime > 0 {
				deadline = time.Now().Add(s.windowTime)
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

func (s *ReplaySubject) call(ctx context.Context, sink Observer, source Observable) (context.Context, context.CancelFunc) {
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
						s.observers[len(s.observers)-1] = nil
						s.observers = s.observers[:len(s.observers)-1]
						break
					}
				}
				s.try.Unlock()
			}
		}()

		s.trimBuffer()

		for e := s.buffer.Front(); e != nil; e = e.Next() {
			if isDone(ctx) {
				break
			}
			sink.Next(e.Value.(replaySubjectValue).Value)
		}

		s.try.Unlock()
		return ctx, cancel
	}

	if s.err != nil {
		sink.Error(s.err)
		return canceledCtx, nothingToDo
	}

	s.trimBuffer()

	for e := s.buffer.Front(); e != nil; e = e.Next() {
		if isDone(ctx) {
			return canceledCtx, nothingToDo
		}
		sink.Next(e.Value.(replaySubjectValue).Value)
	}
	sink.Complete()
	return canceledCtx, nothingToDo
}

// NewReplaySubject returns a new ReplaySubject.
func NewReplaySubject(bufferSize int, windowTime time.Duration) *ReplaySubject {
	s := &ReplaySubject{
		bufferSize: bufferSize,
		windowTime: windowTime,
	}
	s.Observer = s.notify
	s.Observable = s.Observable.Lift(s.call)
	return s
}
