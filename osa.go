// Package osa provides abstractions over various OS-related operations.
//
// Here is a simple example of how to use this package:
//
//	import (
//		os "github.com/echocrow/osa"
//	)
//
//	data, err := os.ReadFile("file.go") // Use "os" as usual.
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Note: This package only supports a subset of functions and variables of the
// actual os standard library package.
//
package osa

import (
	"io/fs"

	"github.com/echocrow/osa/oos"
)

// Interface I describes available OS methods on the OS abstraction.
type I interface {
	// Open opens the named file.
	Open(name string) (fs.File, error)
	// Lstat returns a FileInfo describing the named file.
	Stat(name string) (FileInfo, error)
	// IsExist returns a boolean indicating whether the error is known to report
	// that a file or directory already exists.
	IsExist(err error) bool
	// IsNotExist returns a boolean indicating whether the error is known to
	// report that a file or directory does not exist.
	IsNotExist(err error) bool
	// PathSeparator returns the directory separator character.
	PathSeparator() uint8
	// IsPathSeparator reports whether c is a directory separator character.
	IsPathSeparator(c uint8) bool
	// Mkdir creates a new directory.
	Mkdir(name string, perm FileMode) error
	// MkdirAll creates a directory named path, along with any necessary parents.
	MkdirAll(name string, perm FileMode) error
	// MkdirTemp creates a new temporary directory in the directory dir and
	// returns the pathname of the new directory.
	MkdirTemp(dir, pattern string) (string, error)
	// ReadDir reads the named directory and returns all its directory entries
	// sorted by filename.
	ReadDir(name string) ([]DirEntry, error)
	// WriteFile writes data to the named file, creating it if necessary.
	WriteFile(name string, data []byte, perm FileMode) error
	// ReadFile reads the named file and returns the contents.
	ReadFile(name string) ([]byte, error)
	// Rename renames (moves) oldpath to newpath.
	Rename(oldpath, newpath string) error
	// Remove removes the named file or empty directory.
	Remove(name string) error
	// RemoveAll removes path and any children it contains
	RemoveAll(path string) error
	// Getwd returns a rooted path name corresponding to the current directory.
	Getwd() (dir string, err error)
	// UserCacheDir returns the default directory to use for cached data.
	UserCacheDir() (string, error)
	// UserConfigDir returns the default directory to use for configuration data.
	UserConfigDir() (string, error)
	// UserHomeDir returns the current user's home directory.
	UserHomeDir() (string, error)
	// Exit causes the current program to exit with the given status code.
	Exit(code int)
	// Stdio returns IO readers and writers for Stdin, Stdout, and Stderr.
	Stdio
}

// Default returns the standard OS abstraction implementation.
func Default() I { return oos.New() }

var osa I = Default()

// Current returns the current OS abstraction implementation.
func Current() I { return osa }

// Patch monkey-patches the OS abstraction and returns a restore function.
//
// BUG(echocrow): Patch is not yet thread-safe.
//
// BUG(echocrow): Calling Patch multiple times before resetting may result in
// incomplete resets when reset funcs are not invoked in reverse order.
func Patch(o I) func() {
	org := osa
	osa = o
	return func() { osa = org }
}

//go:generate go run ./gen -call=osa
