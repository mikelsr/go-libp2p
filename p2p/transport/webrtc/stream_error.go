package libp2pwebrtc

import "net"

type Error struct {
	message            string
	temporary, timeout bool
}

var _ net.Error = &Error{}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Temporary() bool {
	return e.temporary
}

func (e *Error) Timeout() bool {
	return e.timeout
}

// ErrTimeout is used when an i/o deadline is reached
var ErrTimeout = &Error{message: "i/o deadline reached", timeout: true, temporary: true}
