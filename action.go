package profile

import (
	"io"
	"log"
	"os"
	"time"
)

// The base action is used to generate all the actions in Swat.
type BaseAction struct {
	*scheduler
	*targeter
	*signaler
	fn      func(io.Writer) error
	lastErr error
}

var _ Action = &BaseAction{}

// Creates and returns a new generic action. Currently there
// are two ways of triggering actions: by using a Scheduler
// and/or Signaller.
//
// The Scheduler is activated by calling After, Every, At, For,
// or Until, and is used for running tasks after durations or
// on intervals.
//
// The Signaller is activated by calling OnSignal, and will trigger
// an action to be run when a process gets a POSIX signal.
//
// The output of the action can be sent to a writer. You can specify
// a writer using ToWriter, and there's a shortcut for specifying
// a file output using ToFile.
func NewAction(fn func(io.Writer) error) *BaseAction {
	return &BaseAction{
		scheduler: new(scheduler),
		targeter:  new(targeter),
		signaler:  new(signaler),
		fn:        fn,
	}
}

// `After` starts something after a given duration. Cannot be used
// with `At`. Omitting `After` and `At` cause the scheduler to
// start the task immediately
func (b *BaseAction) After(after time.Duration) *BaseAction {
	b.scheduler.After(after)
	return b
}

// `At` starts something at a given time. Cannot be used with
// `After`. Omitting `After` and `At` cause the scheduler to
// start the task immediately.
func (b *BaseAction) At(at time.Time) *BaseAction {
	b.scheduler.At(at)
	return b
}

// Specifies the time between runs, after the start time (specified
// by `After` or `At`) has passed. Omitting `Every` causes the
// action to be run just once.
func (b *BaseAction) Every(every time.Duration) *BaseAction {
	b.scheduler.Every(every)
	return b
}

// `For` specifies how long `Every` runs. annot be
// used with `Until`. Omitting both `For` and `Every` cause
// the event to run for an infinite time.
func (b *BaseAction) For(length time.Duration) *BaseAction {
	b.scheduler.For(length)
	return b
}

// `Unil` specifies a time at which `Every` stops. Cannot be
// used with `For`. Omitting both `For` and `Every` cause the
// event to run for an infinite time.
func (b *BaseAction) Until(until time.Time) *BaseAction {
	b.scheduler.Until(until)
	return b
}

// Used to run an action when an OS signal is received.
func (b *BaseAction) OnSignal(signals ...os.Signal) *BaseAction {
	b.signaler.OnSignal(signals...)
	return b
}

// Writes the output of the action to the writer.
func (b *BaseAction) ToWriter(w io.Writer) *BaseAction {
	b.targeter.ToWriter(w)
	return b
}

// Writes the output of the action to the writer.
func (b *BaseAction) ToFile(f string) *BaseAction {
	if b.lastErr == nil {
		b.lastErr = b.targeter.ToFile(f)
	}

	return b
}

// Implements Action.Start
func (b *BaseAction) Start() error {
	if b.lastErr != nil {
		return b.lastErr
	}

	if err := b.scheduler.validate(); err != nil {
		return err
	}

	fn := func() {
		if err := b.fn(b.writer); err != nil {
			log.Printf("Swat Error: %s", err)
		}
	}

	b.scheduler.fn = fn
	b.signaler.fn = fn

	go b.scheduler.start()
	go b.signaler.start()

	return nil
}

// Implements Action.End
func (b *BaseAction) End() {
	parallel(
		b.signaler.end,
		b.scheduler.end,
		b.targeter.end,
	)
}
