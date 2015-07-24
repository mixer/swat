package profile

import (
	"os"
	"os/signal"
)

// Signaller is an embedded struct used to trigger actions when
// a syscall is sent. It should not be used directly.
type Signaler struct {
	fn      func()
	signals []os.Signal
	closer  chan bool
}

// Used to run an action when an OS signal is received.
func (s *Signaler) OnSignal(signals ...os.Signal) {
	s.signals = signals
}

func (s *Signaler) end() {
	select {
	case <-s.closer:
	case s.closer <- true:
		<-s.closer
	}
}

func (s *Signaler) start() {
	defer func() {
		s.closer <- true
	}()

	if len(s.signals) == 0 {
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, s.signals...)

	for {
		select {
		case <-s.closer:
			return
		case <-ch:
			s.fn()
		}
	}
}
