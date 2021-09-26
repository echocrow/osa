package vos

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	os "github.com/echocrow/osa"
)

var (
	errNotDir   = errors.New("not a directory")
	errNotFile  = errors.New("not a file")
	errNotEmpty = errors.New("not empty")
	errOpFailed = errors.New("operation failed")
)

type vosFS struct {
	temp string

	home   string
	usrCch string
	usrCfg string

	pwd string

	entries vDir
}

func newFS() vosFS {
	v := &vosFS{
		entries: newVDir(),
	}

	sep := string(v.PathSeparator())
	mkdir := func(p string) {
		if err := v.Mkdir(p, 0700); err != nil {
			panic(err)
		}
	}

	v.temp = sep + "temp"
	mkdir(v.temp)

	v.home = sep + "home"
	mkdir(v.home)
	v.usrCch = filepath.Join(v.home, ".cache")
	mkdir(v.usrCch)
	v.usrCfg = filepath.Join(v.home, ".config")
	mkdir(v.usrCfg)

	v.pwd = v.home

	return *v
}

func (vosFS) PathSeparator() uint8 {
	return filepath.Separator
}

func (v vosFS) Mkdir(name string, perm fs.FileMode) error {
	parent, base := filepath.Split(name)
	parDir, err := v.getDir(parent)
	if err != nil {
		return newPathError("mkdir", parent, err)
	}
	if err := parDir.add(base, newVDir()); err != nil {
		return newPathError("mkdir", name, err)
	}
	return nil
}

func (v vosFS) get(p string) (dirEntry, error) {
	dir := v.entries
	var got dirEntry = dir
	var err error
	for _, name := range v.splitPath(p) {
		got, err = dir.get(name)
		if err != nil {
			return nil, err
		}
		dir, _ = got.(vDir)
	}
	return got, nil
}

func (v vosFS) getDir(p string) (vDir, error) {
	e, err := v.get(p)
	if err != nil {
		return vDir{}, err
	}
	dir, ok := e.(vDir)
	if !ok {
		return vDir{}, errNotDir
	}
	return dir, nil
}

func (v vosFS) getFile(p string) (vFile, error) {
	e, err := v.get(p)
	if err != nil {
		return vFile{}, err
	}
	if file, ok := e.(vFile); !ok {
		return vFile{}, errNotFile
	} else {
		return file, nil
	}
}

// splitPath splits a given path into a slice of path components.
func (v vosFS) splitPath(p string) []string {
	sep := string(v.PathSeparator())
	p = filepath.Clean(p)
	if p == sep {
		return nil
	} else if !strings.HasPrefix(p, sep) {
		panic(errors.New("unexpected relative path"))
	}
	return strings.Split(p, sep)[1:]
}

func (v vosFS) Open(name string) (fs.File, error) {
	got, err := v.get(name)
	if err != nil {
		return nil, newPathError("open", name, err)
	}
	_, base := filepath.Split(name)
	var data []byte
	if file, ok := got.(vFile); ok {
		data = file.data
	}
	return &fsFile{
		name:   base,
		data:   data,
		isOpen: true,
	}, nil
}

func (v vosFS) Stat(name string) (os.FileInfo, error) {
	e, err := v.get(name)
	if err != nil {
		return nil, newPathError("stat", name, err)
	}
	_, base := filepath.Split(name)
	return osFileInfo{base, e.isDir(), 0}, nil
}

func (vosFS) IsExist(err error) bool {
	return underlyingError(err) == fs.ErrExist
}

func (vosFS) IsNotExist(err error) bool {
	return underlyingError(err) == fs.ErrNotExist
}

func (v vosFS) IsPathSeparator(c uint8) bool {
	return c == v.PathSeparator()
}

func (v vosFS) MkdirAll(name string, perm os.FileMode) error {
	dir := v.entries
	for _, n := range v.splitPath(name) {
		got := dir.tryGet(n)
		if got == nil {
			d := newVDir()
			if err := dir.add(n, d); err != nil {
				return newPathError("mkdir", name, err)
			}
			dir = d
		} else {
			var ok bool
			if dir, ok = got.(vDir); !ok {
				return newPathError("mkdir", name, errNotDir)
			}
		}
	}
	return nil
}

func (v vosFS) MkdirTemp(dir, pattern string) (string, error) {
	if dir == "" {
		dir = v.temp
	}
	var tmpDirSfx uint = 1
	for {
		base := pattern + fmt.Sprint(tmpDirSfx)
		path := filepath.Join(dir, base)
		if tmpDirSfx == 0 {
			return "", newPathError("mkdir", path, errOpFailed)
		}
		if err := v.Mkdir(path, 0700); !v.IsExist(err) {
			return path, err
		}
		tmpDirSfx++
	}
}

func (v vosFS) ReadDir(name string) ([]os.DirEntry, error) {
	dir, err := v.getDir(name)
	if err != nil {
		return nil, newPathError("open", name, err)
	}
	return dir.read(), nil
}

func (v vosFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	parent, base := filepath.Split(name)
	parDir, err := v.getDir(parent)
	if err != nil {
		return newPathError("open", parent, err)
	}
	if err := parDir.update(base, vFile{data}); err != nil {
		return newPathError("write", name, err)
	}
	return nil
}

func (v vosFS) ReadFile(name string) ([]byte, error) {
	f, err := v.getFile(name)
	if err != nil {
		return nil, newPathError("open", name, err)
	}
	return f.data, nil
}

func (v vosFS) Rename(oldpath, newpath string) error {
	oldE, err := v.get(oldpath)
	if err != nil {
		return newPathError("rename", oldpath, err)
	}

	oParent, oBase := filepath.Split(oldpath)
	oParDir, err := v.getDir(oParent)
	if err != nil {
		return newPathError("rename", oldpath, err)
	}
	if !oParDir.has(oBase) {
		return newPathError("rename", oldpath, fs.ErrNotExist)
	}

	nParent, nBase := filepath.Split(newpath)
	nParDir, err := v.getDir(nParent)
	if err != nil {
		return newPathError("rename", newpath, err)
	}
	if collE, err := nParDir.get(nBase); err == nil && collE.isDir() {
		return newPathError("rename", newpath, fs.ErrExist)
	}
	if err := nParDir.update(nBase, oldE); err != nil {
		return newPathError("rename", newpath, err)
	}
	oParDir.delete(oBase)
	return nil
}

func (v vosFS) Remove(name string) error {
	parent, base := filepath.Split(name)
	parDir, err := v.getDir(parent)
	if err != nil {
		return err
	}
	e, err := parDir.get(base)
	if err != nil {
		return newPathError("open", parent, err)
	}
	if e.isDir() && !e.isEmpty() {
		return newPathError("remove", name, errNotEmpty)
	}
	parDir.delete(base)
	return nil
}

func (v vosFS) RemoveAll(name string) error {
	parent, base := filepath.Split(name)
	if parDir, err := v.getDir(parent); err == nil {
		parDir.delete(base)
	}
	return nil
}

func (v vosFS) Getwd() (dir string, err error) {
	return v.pwd, nil
}

func (v vosFS) UserCacheDir() (string, error) {
	return v.usrCch, nil
}

func (v vosFS) UserConfigDir() (string, error) {
	return v.usrCfg, nil
}

func (v vosFS) UserHomeDir() (string, error) {
	return v.home, nil
}

// MkTempDir creates a new pseudo-temporary directory.
func MkTempDir(v vos) string {
	path, err := v.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	return path
}

// underlyingError returns the underlying error for known os error types.
func underlyingError(err error) error {
	switch err := err.(type) {
	case *os.PathError:
		return err.Err
	}
	return err
}

func newPathError(op, path string, err error) *os.PathError {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}
