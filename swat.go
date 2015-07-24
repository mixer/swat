package profile

import (
	"sync"
)

// An "action" is the basic unit of Swat. It is started and should
// listen for events, until End is called.
type Action interface {
	// Start should verify conditions, returning an error if
	// necessary, then start listening asynchronously.
	Start() error
	// End should signal the action to stop, and block until it does.
	End()
}

type Swat struct {
	actions []Action
}

// Creates a Swat with the given actions, and boots them
// all automatically.
func Start(actions ...Action) (*Swat, error) {
	s := new(Swat)
	return s, s.Boot(actions)
}

// Starts all associated actions. If an action's Start method returns
// an error, then no actions are run.
func (s *Swat) Boot(actions []Action) error {
	for _, action := range actions {
		if err := action.Start(); err != nil {
			s.End()
			return err
		}

		s.actions = append(s.actions, action)
	}

	return nil
}

// Closes and waits for all actions to end.
func (s *Swat) End() {
	wg := new(sync.WaitGroup)
	for _, action := range s.actions {
		wg.Add(1)
		go func(action Action) {
			defer wg.Done()
			action.End()
		}(action)
	}

	wg.Wait()
}
