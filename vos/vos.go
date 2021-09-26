// Package vos provides a basic virtual OS abstraction implementation.
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
