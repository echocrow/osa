// Package vos provides the original OS abstraction implementation.
package oos

import "os"

//go:generate go run ../gen -pkg=.. -name=oos -call=os -import=os
func New() oos {
	return oos{}
}

type FileInfo = os.FileInfo
type FileMode = os.FileMode
type DirEntry = os.DirEntry

func PathSeparator() uint8 { return os.PathSeparator }
