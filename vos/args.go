package vos

import "os"

// PatchArgs allows temporary overwriting of OS package args.
func PatchArgs() (set func(args []string), reset func()) {
	org := os.Args

	set = func(args []string) { os.Args = args }
	reset = func() { os.Args = org }
	return
}
