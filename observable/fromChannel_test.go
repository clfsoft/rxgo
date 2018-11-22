package observable_test

import (
	"testing"

	. "github.com/b97tsk/rxgo/observable"
)

func TestFromChannel(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		c := make(chan interface{})
		go func() {
			c <- "A"
			c <- "B"
			c <- "C"
			close(c)
		}()
		subscribe(t, []Observable{FromChannel(c)}, "A", "B", "C", xComplete)
	})
	t.Run("#2", func(t *testing.T) {
		c := make(chan interface{})
		go func() {
			c <- "A"
			c <- "B"
			c <- "C"
			close(c)
		}()
		subscribe(
			t,
			[]Observable{
				FromChannel(c).Pipe(operators.Take(1)),
				FromChannel(c).Pipe(operators.Take(2)),
			},
			"A", xComplete,
			"B", "C", xComplete,
		)
	})
}
