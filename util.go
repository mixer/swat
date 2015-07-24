package profile

import (
	"sync"
)

// Does several functions in parallel and waits for their completion.
func parallel(fns ...func()) {
	wg := new(sync.WaitGroup)
	for _, fn := range fns {
		wg.Add(1)
		go fn()
	}

	wg.Wait()
}
