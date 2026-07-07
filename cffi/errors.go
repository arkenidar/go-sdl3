package main

import "sync"

var (
	errMu   sync.Mutex
	lastErr string
)

func setLastError(err error) {
	errMu.Lock()
	defer errMu.Unlock()
	if err != nil {
		lastErr = err.Error()
	} else {
		lastErr = ""
	}
}

func getLastError() string {
	errMu.Lock()
	defer errMu.Unlock()
	return lastErr
}

// guard recovers a panic inside an exported function and reports it as an
// error, so a Go panic never unwinds across the cgo boundary. f should set
// any out-params via captured pointers before returning.
func guard(f func() error) (status int32) {
	defer func() {
		if r := recover(); r != nil {
			setLastError(&panicError{r})
			status = -1
		}
	}()
	if err := f(); err != nil {
		setLastError(err)
		return -1
	}
	setLastError(nil)
	return 0
}

type panicError struct{ v any }

func (p *panicError) Error() string {
	if e, ok := p.v.(error); ok {
		return "panic: " + e.Error()
	}
	return "panic: recovered"
}
