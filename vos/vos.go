// Package vos provides a basic virtual OS abstraction implementation.
//
// This package mimicks "os" package features in-memory, so no real files are
// created, read, updated, or deleted. The package provides a "Patch()"
// function to inject this implementation e.g. during testing. Only test
// packages typically need to use this.
package vos

import (
	os "github.com/echocrow/osa"
)

func Patch() (vos, func()) {
	o := New()
	restore := os.Patch(o)
	return o, restore
}

type vos struct {
	vosFS
	vosIO
}

func New() vos {
	return vos{
		vosFS: newFS(),
		vosIO: newIO(),
	}
}
