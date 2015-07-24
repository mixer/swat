package profile

import (
	"io"
	"log"
	"os"
	"time"
)

// The base action is used to generate all the actions in Swat.
type BaseAction struct {
	*Scheduler
	*Targeter
	*Signaler
	fn      func(io.Writer) error
	lastErr error
}

var _ Action = &BaseAction{}

func newBaseAction(fn func(io.Writer) error) *BaseAction {
	return &BaseAction{
		Scheduler: new(Scheduler),
		Targeter:  new(Targeter),
		Signaler:  new(Signaler),
		fn:        fn,
	}
}

// `after` starts something at the given duration after the current time.
func (b *BaseAction) After(after time.Duration) *BaseAction {
	b.Scheduler.After(after)
	return b
}

func (b *BaseAction) Every(every time.Duration) *BaseAction {
	b.Scheduler.Every(every)
	return b
}

// `at` start something at a given time
func (b *BaseAction) At(at time.Time) *BaseAction {
	b.Scheduler.At(at)
	return b
}

// `for` specifies how long "every" runs. Omitting it runs it
// for infinite time.
func (b *BaseAction) For(length time.Duration) *BaseAction {
	b.Scheduler.For(length)
	return b
}

// `for` specifies how long "every" runs. Omitting it runs it
// for infinite time.
func (b *BaseAction) Until(until time.Time) *BaseAction {
	b.Scheduler.Until(until)
	return b
}

// Used to run an action when an OS signal is received.
func (b *BaseAction) OnSignal(signals ...os.Signal) *BaseAction {
	b.Signaler.OnSignal(signals...)
	return b
}

// Writes the output of the action to the writer.
func (b *BaseAction) ToWriter(w io.Writer) *BaseAction {
	b.Targeter.ToWriter(w)
	return b
}

// Writes the output of the action to the writer.
func (b *BaseAction) ToFile(f string) *BaseAction {
	if b.lastErr == nil {
		b.lastErr = b.Targeter.ToFile(f)
	}

	return b
}

// Implements Action.Start
func (b *BaseAction) Start() error {
	if b.lastErr != nil {
		return b.lastErr
	}

	if err := b.Scheduler.validate(); err != nil {
		return err
	}

	fn := func() {
		if err := b.fn(b.writer); err != nil {
			log.Printf("Swat Error: %s", err)
		}
	}

	b.Scheduler.fn = fn
	b.Signaler.fn = fn

	go b.Scheduler.start()
	go b.Signaler.start()

	return nil
}

// Implements Action.End
func (b *BaseAction) End() {
	parallel(
		b.Signaler.end,
		b.Scheduler.end,
		b.Targeter.end,
	)
}
