package vos

import (
	"io/fs"
	"sort"

	os "github.com/echocrow/osa"
)

type dirEntry interface {
	isDir() bool
	isEmpty() bool
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

func (es dirEntries) read() []os.DirEntry {
	res := make([]os.DirEntry, len(es))

	names := make([]string, len(es))
	i := 0
	for n := range es {
		names[i] = n
		i++
	}
	sort.Strings(names)

	for i, n := range names {
		e := es[n]
		res[i] = osDirEntry{n, e.isDir()}
	}

	return res
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

type vFile struct {
	data []byte
}

func (vFile) isDir() bool {
	return false
}

func (f vFile) isEmpty() bool {
	return len(f.data) == 0
}
