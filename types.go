package osa

import "io/fs"

// A DirEntry is an entry read from a directory.
type DirEntry = fs.DirEntry

// A FileMode represents a file's mode and permission bits.
type FileMode = fs.FileMode

// A FileInfo describes a file and is returned by Stat and Lstat.
type FileInfo = fs.FileInfo

// PathError records an error and the operation and file path that caused it.
type PathError = fs.PathError
