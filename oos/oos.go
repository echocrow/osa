// Package oos provides the original OS abstraction implementation.
//
// This package simply wraps and calls the default "os" functions of the
// standard library. This is the default "osa" implementation, so typically code
// does not need to import or directly interact with this package.
package oos

import "os"

//go:generate go run ../gen -pkg=.. -name=oos -call=os -import=os
func New() oos {
	return oos{}
}

type File = *os.File
type FileInfo = os.FileInfo
type FileMode = os.FileMode
type DirEntry = os.DirEntry

func PathSeparator() uint8 { return os.PathSeparator }
