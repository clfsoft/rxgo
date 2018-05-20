package observable

import (
	"errors"
)

var (
	ErrEmpty           = errors.New("empty")
	ErrNotNotification = errors.New("not notification")
	ErrNotObservable   = errors.New("not observable")
	ErrOutOfRange      = errors.New("out of range")
	ErrTimeout         = errors.New("timeout")
	ErrTooMany         = errors.New("too many")
)
