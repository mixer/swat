package profile

import (
	"errors"
	"io"
	"runtime/pprof"
)

// Returns an action that dumps a pprof lookup, with the
// given name and debug constant.
func DumpPProfLookup(name string, debug int) *BaseAction {
	return NewAction(func(w io.Writer) error {
		pp := pprof.Lookup(name)
		if pp == nil {
			return errors.New("unknown pprof " + name)
		}

		if err := pp.WriteTo(w, debug); err != nil {
			return errors.New("error writing pprof: " + err.Error())
		}

		return nil
	})
}

// Returns an action that dumps all running goroutines,
// like you'd get from a panic.
func DumpGoroutine() *BaseAction {
	return DumpPProfLookup("goroutine", 2)
}

// Returns an action that dumps a sample of all head allocations.
func DumpHeap() *BaseAction {
	return DumpPProfLookup("heap", 1)
}

// Returns an action that dumps stack traces that led to
// blocking on synchronization primitives.
func DumpBlocking() *BaseAction {
	return DumpPProfLookup("block", 1)
}

// Returns an action that dumps stack traces that led
// to the creation of new OS threads.
func DumpThreadCreate() *BaseAction {
	return DumpPProfLookup("threadcreate", 1)
}
