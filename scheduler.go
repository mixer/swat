package profile

import (
	"errors"
	"time"
)

// Scheduler is meant to be embedded in actions to allow interval
// and timeouts to be created for them. It should not be used
// directly.
//
// Usage of this is a bit complex, but *should* be reasonably
// natural to use.
//  - `at` start something at a given time
//  - `after` starts something at the given duration after the
//    current time.
//  - Omitting both `at` and `after` sets the start time at the
//    current time.
//  - `every` runs something at an interval after its start time
//  - Omitting `every` runs something just once.
//  - `for` specifies how long "every" runs
//  - `until` specifies a time for "every" to stop running at
type Scheduler struct {
	fn     func()
	at     time.Time
	after  time.Duration
	every  time.Duration
	length time.Duration
	until  time.Time
	closer chan bool
}

func newScheduler(fn func()) *Scheduler {
	return &Scheduler{fn: fn, closer: make(chan bool, 1)}
}

// `after` starts something at the given duration after the current time.
func (s *Scheduler) After(after time.Duration) *Scheduler {
	s.after = after
	return s
}

// `every` runs something at an interval after its start time.
// Omitting it runs something just once.
func (s *Scheduler) Every(every time.Duration) *Scheduler {
	s.every = every
	return s
}

// `at` start something at a given time
func (s *Scheduler) At(at time.Time) *Scheduler {
	s.at = at
	return s
}

// `for` specifies how long "every" runs. Omitting it runs it
// for infinite time.
func (s *Scheduler) For(length time.Duration) *Scheduler {
	s.length = length
	return s
}

// `for` specifies how long "every" runs. Omitting it runs it
// for infinite time.
func (s *Scheduler) Until(until time.Time) *Scheduler {
	s.until = until
	return s
}

func (s *Scheduler) validate() error {
	if !s.at.IsZero() && s.after > 0 {
		return errors.New("Swat Error: Using both 'At' and 'After' will lead to unexepected results.")
	}

	if !s.until.IsZero() && s.length > 0 {
		return errors.New("Swat Error: Using both 'Until' and 'For' will lead to unexepected results.")
	}

	if (s.length > 0 || !s.until.IsZero()) && s.every == 0 {
		return errors.New("Swat Error: 'Every' is required when using 'Until' or 'For'.")
	}

	return nil
}

func (s *Scheduler) end() {
	select {
	case <-s.closer:
	case s.closer <- true:
		<-s.closer
	}
}

// gets the initial sleep time before starting calling the function.
func (s *Scheduler) resolveSleep() time.Duration {
	if s.after > 0 {
		return s.after
	} else if !s.at.IsZero() {
		return s.at.Sub(time.Now())
	}

	return 0
}

// Returns whether there's enough data to qualify the scheduler
// as being activated.
func (s *Scheduler) isActivated() bool {
	return s.after > 0 ||
		!s.at.IsZero() ||
		s.every > 0
}

// Returns the time that the scheduler should run until.
func (s *Scheduler) getUntil() time.Time {
	if s.length > 0 {
		return time.Now().Add(s.length)
	} else if !s.until.IsZero() {
		return s.until
	}

	// humanity will be extinct long before this time, but if
	// needed this can be refactored when we transcend space-time
	return time.Unix(1<<62, 0)
}

func (s *Scheduler) start() {
	defer func() {
		s.closer <- true
	}()

	// Deactivate the scheduler if nothing useful was passed.

	if !s.isActivated() {
		return
	}

	select {
	case <-s.closer:
		return
	case <-time.After(s.resolveSleep()):
	}

	until := s.getUntil()
	for time.Now().Before(until) {
		s.fn()

		if s.every == 0 {
			return
		}

		select {
		case <-s.closer:
			return
		case <-time.After(s.every):
		}
	}
}
