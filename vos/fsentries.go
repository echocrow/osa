package vos

import (
	"io/fs"
)

type dirEntry interface {
	isDir() bool
	size() int
	isEmpty() bool
	toFile(name string) (fs.File, error)
}

type dirEntries map[string]dirEntry

func (es dirEntries) has(name string) bool {
	_, ok := es[name]
	return ok
}

func (es dirEntries) tryGet(name string) dirEntry {
	got := es[name]
	return got
}

func (es dirEntries) get(name string) (dirEntry, error) {
	if got := es.tryGet(name); got == nil {
		return nil, fs.ErrNotExist
	} else {
		return got, nil
	}
}

func (es dirEntries) add(name string, e dirEntry) error {
	if es.has(name) {
		return fs.ErrExist
	}
	es[name] = e
	return nil
}

func (es dirEntries) update(name string, e dirEntry) error {
	if coll := es.tryGet(name); coll != nil {
		wantDir := e.isDir()
		if wantDir != coll.isDir() {
			if wantDir {
				return errNotDir
			}
			return errNotFile
		}
		if coll.isDir() && !coll.isEmpty() {
			return errNotEmpty
		}
	}
	es[name] = e
	return nil
}

func (es dirEntries) delete(name string) {
	delete(es, name)
}

func (es dirEntries) list() []fsFileInfo {
	contents := make([]fsFileInfo, len(es))
	i := 0
	for n, e := range es {
		contents[i] = fsFileInfo{n, e.isDir(), int64(e.size())}
		i++
	}
	return contents
}

func (es dirEntries) size() int {
	return len(es)
}

func (es dirEntries) isEmpty() bool {
	return es.size() == 0
}

type vDir struct {
	dirEntries
}

func newVDir() vDir {
	return vDir{make(dirEntries)}
}

func (vDir) isDir() bool {
	return true
}

func (d vDir) toFile(name string) (fs.File, error) {
	return &fsDir{
		fsFile: &fsFile{
			name: name,
			data: nil,
		},
		contents: d.list(),
	}, nil
}

type vFile struct {
	data []byte
}

func newVFile(data []byte) vFile {
	if data == nil {
		data = make([]byte, 0)
	}
	return vFile{data}
}

func (vFile) isDir() bool {
	return false
}

func (f vFile) size() int {
	return len(f.data)
}
func (f vFile) isEmpty() bool {
	return f.size() == 0
}

func (f vFile) toFile(name string) (fs.File, error) {
	return &fsFile{
		name: name,
		data: f.data,
	}, nil
}
