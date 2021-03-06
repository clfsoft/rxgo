package observable_test

import (
	"context"
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestObservable_Publish(t *testing.T) {
	obs := Interval(step(1)).Publish()
	ctx, _ := Zip(
		obs.Pipe(operators.Take(4)),
		obs.Pipe(operators.Skip(4), operators.Take(4)),
	).Pipe(
		operators.Map(
			func(val interface{}, idx int) interface{} {
				vals := val.([]interface{})
				return vals[0].(int) * vals[1].(int)
			},
		),
		operators.ToSlice(),
		toString,
	).Subscribe(context.Background(), func(x Notification) {
		switch {
		case x.HasValue:
			if x.Value != "[0 5 12 21]" {
				t.Fail()
			}
		case x.HasError:
			t.Error(x.Value)
		}
	})
	select {
	case <-ctx.Done():
		t.Fail()
		return
	default:
	}
	_, disconnect := obs.Connect()
	defer disconnect()
	<-ctx.Done()
}

func TestObservable_PublishBehavior(t *testing.T) {
	obs := Interval(step(1)).PublishBehavior(-1)
	ctx, _ := Zip(
		obs.Pipe(operators.Take(4)),
		obs.Pipe(operators.Skip(4), operators.Take(4)),
	).Pipe(
		operators.Map(
			func(val interface{}, idx int) interface{} {
				vals := val.([]interface{})
				return vals[0].(int) * vals[1].(int)
			},
		),
		operators.ToSlice(),
		toString,
	).Subscribe(context.Background(), func(x Notification) {
		switch {
		case x.HasValue:
			if x.Value != "[-3 0 5 12]" {
				t.Fail()
			}
		case x.HasError:
			t.Error(x.Value)
		}
	})
	select {
	case <-ctx.Done():
		t.Fail()
		return
	default:
	}
	_, disconnect := obs.Connect()
	defer disconnect()
	<-ctx.Done()
}

func TestObservable_PublishReplay(t *testing.T) {
	obs := Interval(step(2)).PublishReplay(2, 0)
	ctx, _ := Zip(
		obs.Pipe(operators.Take(4)),
		obs.Pipe(operators.Skip(4), operators.Take(4), delaySubscription(7)),
	).Pipe(
		operators.Map(
			func(val interface{}, idx int) interface{} {
				vals := val.([]interface{})
				return vals[0].(int) * vals[1].(int)
			},
		),
		operators.ToSlice(),
		toString,
	).Subscribe(context.Background(), func(x Notification) {
		switch {
		case x.HasValue:
			if x.Value != "[0 6 14 24]" {
				t.Fail()
			}
		case x.HasError:
			t.Error(x.Value)
		}
	})
	select {
	case <-ctx.Done():
		t.Fail()
		return
	default:
	}
	_, disconnect := obs.Connect()
	defer disconnect()
	<-ctx.Done()
}

func TestOperators_Share(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		obs := Interval(step(3)).Pipe(
			operators.Take(4),
			operators.Share(),
		)
		subscribe(
			t,
			[]Observable{
				Merge(
					obs,
					obs.Pipe(delaySubscription(4)),
					obs.Pipe(delaySubscription(8)),
					obs.Pipe(delaySubscription(13)),
				),
			},
			0, 1, 1, 2, 2, 2, 3, 3, 3, 0, 1, 2, 3, xComplete,
		)
	})
	t.Run("#2", func(t *testing.T) {
		obs := Interval(step(3)).Pipe(
			operators.Share(),
			operators.Take(4),
		)
		subscribe(
			t,
			[]Observable{
				Merge(
					obs,
					obs.Pipe(delaySubscription(4)),
					obs.Pipe(delaySubscription(8)),
					obs.Pipe(delaySubscription(19)),
				),
			},
			0, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 5, 0, 1, 2, 3, xComplete,
		)
	})
}

func TestOperators_ShareReplay(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		obs := Interval(step(3)).Pipe(
			operators.Take(4),
			operators.ShareReplay(1, 0),
		)
		subscribe(
			t,
			[]Observable{
				Merge(
					obs,
					obs.Pipe(delaySubscription(4)),
					obs.Pipe(delaySubscription(8)),
					obs.Pipe(delaySubscription(13)),
				),
			},
			0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 0, 1, 2, 3, xComplete,
		)
	})
	t.Run("#2", func(t *testing.T) {
		obs := Interval(step(3)).Pipe(
			operators.ShareReplay(1, 0),
			operators.Take(4),
		)
		subscribe(
			t,
			[]Observable{
				Merge(
					obs,
					obs.Pipe(delaySubscription(4)),
					obs.Pipe(delaySubscription(8)),
					obs.Pipe(delaySubscription(16)),
				),
			},
			0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 0, 1, 2, 3, xComplete,
		)
	})
}
