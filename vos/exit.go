package vos

func (vos) Exit(code int) {
	panic(exitCode(code))
}

// CatchExit allows recovering from vos.Exit() and calls catch with the denoted
// exit code.
func CatchExit(catch func(code int)) {
	if r := recover(); r != nil {
		if code, ok := r.(exitCode); ok {
			catch(int(code))
		} else {
			panic(r)
		}
	}
}

type exitCode int
