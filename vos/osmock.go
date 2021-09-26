package vos

import (
	"io/fs"
	"time"
)

type osDirEntry struct {
	name  string
	isDir bool
}

func (e osDirEntry) Name() string {
	return e.name
}

func (e osDirEntry) IsDir() bool {
	return e.isDir
}

func (e osDirEntry) Type() fs.FileMode {
	return 0
}

func (e osDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

type osFileInfo struct {
	name  string
	isDir bool
	size  int64
}

func (fi osFileInfo) Name() string {
	return fi.name
}

func (fi osFileInfo) Size() int64 {
	return fi.size
}

func (fi osFileInfo) Mode() fs.FileMode {
	return 0
}

func (osFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fi osFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi osFileInfo) Sys() interface{} {
	return nil
}
