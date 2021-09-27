package vos

import (
	"errors"
	"io"
	"io/fs"
	"time"

	"github.com/echocrow/osa"
)

type fsFile struct {
	name     string
	data     []byte
	isClosed bool
	read     int
}

func (f *fsFile) Stat() (fs.FileInfo, error) {
	return fsFileInfo{
		name:  f.name,
		isDir: f.data == nil,
		size:  int64(len(f.data)),
	}, nil
}

func (f *fsFile) Read(to []byte) (int, error) {
	if f.data == nil {
		return 0, errNotFile
	}
	if f.isClosed {
		return 0, errors.New("not open")
	}
	start := f.read
	end := len(f.data)
	if start >= end {
		return 0, io.EOF
	}
	l := end - start
	if len(to) < l {
		l = len(to)
	}
	copy(to, f.data[start:])
	f.read += l
	return l, nil
}

func (f *fsFile) Close() error {
	if f.isClosed {
		return errors.New("not open")
	}
	f.isClosed = true
	return nil
}

// fsFileInfo represents a fs.FileInfo and fs.DirEntry
type fsFileInfo struct {
	name  string
	isDir bool
	size  int64
}

func (f fsFileInfo) Name() string {
	return f.name
}

func (f fsFileInfo) Size() int64 {
	return f.size
}

func (f fsFileInfo) Mode() osa.FileMode {
	return 0
}
func (f fsFileInfo) Type() osa.FileMode {
	return 0
}

func (f fsFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (f fsFileInfo) IsDir() bool {
	return f.isDir
}

func (f fsFileInfo) Sys() interface{} {
	return nil
}

func (f fsFileInfo) Info() (fs.FileInfo, error) {
	return f, nil
}

type fsDir struct {
	*fsFile
	contents []fsFileInfo
}

// ReadDir reads the contents of the directory and returns a slice of up to n
// DirEntry values in directory order.
//
// See fs.ReadDirFile
func (d *fsDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.isClosed {
		return nil, errors.New("not open")
	}
	start := d.read
	l := len(d.contents) - start
	if l <= 0 {
		return nil, io.EOF
	}
	if 0 < n && n < l {
		l = n
	}
	end := start + l
	entries := make([]fs.DirEntry, l)
	for i, c := range d.contents[start:end] {
		entries[i] = c
	}
	d.read = end
	return entries, nil
}
