// Package osa provides abstractions over various OS-related operations.
//
// Here is a simple example of how to use this package:
//
//	import (
//		os "github.com/echocrow/osa"
//	)
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
	"github.com/echocrow/osa/oos"
)

// Interface I describes available OS methods on the OS abstraction.
type I interface {
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
}

// Default returns the standard OS abstraction implementation.
func Default() I { return oos.New() }

var osa I = Default()

// Current returns the current OS abstraction implementation.
func Current() I { return osa }

// Patch monkey-patches the OS abstraction and returns a restore function.
// TODO Make thread-safe.
// TODO Make chain-patching-safe.
func Patch(o I) func() {
	org := osa
	osa = o
	return func() { osa = org }
}

// Lstat returns a FileInfo describing the named file.
func Stat(name string) (FileInfo, error) { return osa.Stat(name) }

// IsExist returns a boolean indicating whether the error is known to report
// that a file or directory already exists.
func IsExist(err error) bool { return osa.IsExist(err) }

// IsNotExist returns a boolean indicating whether the error is known to
// report that a file or directory does not exist.
func IsNotExist(err error) bool { return osa.IsNotExist(err) }

// PathSeparator returns the directory separator character.
func PathSeparator() uint8 { return osa.PathSeparator() }

// IsPathSeparator reports whether c is a directory separator character.
func IsPathSeparator(c uint8) bool { return osa.IsPathSeparator(c) }

// Mkdir creates a new directory.
func Mkdir(name string, perm FileMode) error { return osa.Mkdir(name, perm) }

// MkdirAll creates a directory named path, along with any necessary parents.
func MkdirAll(name string, perm FileMode) error {
	return osa.MkdirAll(name, perm)
}

// MkdirTemp creates a new temporary directory in the directory dir and
// returns the pathname of the new directory.
func MkdirTemp(dir, pattern string) (string, error) {
	return osa.MkdirTemp(dir, pattern)
}

// ReadDir reads the named directory and returns all its directory entries
// sorted by filename.
func ReadDir(name string) ([]DirEntry, error) { return osa.ReadDir(name) }

// WriteFile writes data to the named file, creating it if necessary.
func WriteFile(name string, data []byte, perm FileMode) error {
	return osa.WriteFile(name, data, perm)
}

// ReadFile reads the named file and returns the contents.
func ReadFile(name string) ([]byte, error) { return osa.ReadFile(name) }

// Rename renames (moves) oldpath to newpath.
func Rename(oldpath, newpath string) error {
	return osa.Rename(oldpath, newpath)
}

// Remove removes the named file or empty directory.
func Remove(name string) error { return osa.Remove(name) }

// RemoveAll removes path and any children it contains
func RemoveAll(path string) error { return osa.RemoveAll(path) }

// Getwd returns a rooted path name corresponding to the current directory.
func Getwd() (dir string, err error) { return osa.Getwd() }

// UserCacheDir returns the default directory to use for cached data.
func UserCacheDir() (string, error) { return osa.UserCacheDir() }

// UserConfigDir returns the default directory to use for configuration data.
func UserConfigDir() (string, error) { return osa.UserConfigDir() }

// UserHomeDir returns the current user's home directory.
func UserHomeDir() (string, error) { return osa.UserHomeDir() }

// Exit causes the current program to exit with the given status code.
func Exit(code int) { osa.Exit(code) }
