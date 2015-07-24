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
type scheduler struct {
	fn     func()
	at     time.Time
	after  time.Duration
	every  time.Duration
	length time.Duration
	until  time.Time
	closer chan bool
}

func newScheduler(fn func()) *scheduler {
	return &scheduler{fn: fn, closer: make(chan bool, 1)}
}

// `after` starts something at the given duration after the current time.
func (s *scheduler) After(after time.Duration) *scheduler {
	s.after = after
	return s
}

// `every` runs something at an interval after its start time.
// Omitting it runs something just once.
func (s *scheduler) Every(every time.Duration) *scheduler {
	s.every = every
	return s
}

// `at` start something at a given time
func (s *scheduler) At(at time.Time) *scheduler {
	s.at = at
	return s
}

// `for` specifies how long "every" runs. Omitting it runs it
// for infinite time.
func (s *scheduler) For(length time.Duration) *scheduler {
	s.length = length
	return s
}

// `for` specifies how long "every" runs. Omitting it runs it
// for infinite time.
func (s *scheduler) Until(until time.Time) *scheduler {
	s.until = until
	return s
}

func (s *scheduler) validate() error {
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

func (s *scheduler) end() {
	select {
	case <-s.closer:
	case s.closer <- true:
		<-s.closer
	}
}

// gets the initial sleep time before starting calling the function.
func (s *scheduler) resolveSleep() time.Duration {
	if s.after > 0 {
		return s.after
	} else if !s.at.IsZero() {
		return s.at.Sub(time.Now())
	}

	return 0
}

// Returns whether there's enough data to qualify the scheduler
// as being activated.
func (s *scheduler) isActivated() bool {
	return s.after > 0 ||
		!s.at.IsZero() ||
		s.every > 0
}

// Returns the time that the scheduler should run until.
func (s *scheduler) getUntil() time.Time {
	if s.length > 0 {
		return time.Now().Add(s.length)
	} else if !s.until.IsZero() {
		return s.until
	}

	// humanity will be extinct long before this time, but if
	// needed this can be refactored when we transcend space-time
	return time.Unix(1<<62, 0)
}

func (s *scheduler) start() {
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
