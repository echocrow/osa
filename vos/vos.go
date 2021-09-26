// Package vos provides a basic virtual OS abstraction implementation.
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

func Patch() (vos, func()) {
	o := New()
	restore := os.Patch(o)
	return o, restore
}

type vos struct {
	temp string

	home   string
	usrCch string
	usrCfg string

	pwd string

	entries vDir
}

func New() vos {
	v := &vos{
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

func (vos vos) Stat(name string) (os.FileInfo, error) {
	e, err := vos.get(name)
	if err != nil {
		return nil, newPathError("stat", name, err)
	}
	_, base := filepath.Split(name)
	return osFileInfo{base, e.isDir(), 0}, nil
}

func (vos vos) IsExist(err error) bool {
	return underlyingError(err) == fs.ErrExist
}

func (vos vos) IsNotExist(err error) bool {
	return underlyingError(err) == fs.ErrNotExist
}

func (vos) PathSeparator() uint8 {
	return filepath.Separator
}

func (vos vos) IsPathSeparator(c uint8) bool {
	return c == vos.PathSeparator()
}

func (vos vos) Mkdir(name string, perm os.FileMode) error {
	parent, base := filepath.Split(name)
	parDir, err := vos.getDir(parent)
	if err != nil {
		return newPathError("mkdir", parent, err)
	}
	if err := parDir.add(base, newVDir()); err != nil {
		return newPathError("mkdir", name, err)
	}
	return nil
}

func (vos vos) MkdirAll(name string, perm os.FileMode) error {
	dir := vos.entries
	for _, n := range vos.splitPath(name) {
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

func (vos vos) MkdirTemp(dir, pattern string) (string, error) {
	if dir == "" {
		dir = vos.temp
	}
	var tmpDirSfx uint = 1
	for {
		base := pattern + fmt.Sprint(tmpDirSfx)
		path := filepath.Join(dir, base)
		if tmpDirSfx == 0 {
			return "", newPathError("mkdir", path, errOpFailed)
		}
		if err := vos.Mkdir(path, 0700); !vos.IsExist(err) {
			return path, err
		}
		tmpDirSfx++
	}
}

func (vos vos) ReadDir(name string) ([]os.DirEntry, error) {
	dir, err := vos.getDir(name)
	if err != nil {
		return nil, newPathError("open", name, err)
	}
	return dir.read(), nil
}

func (vos vos) WriteFile(name string, data []byte, perm os.FileMode) error {
	parent, base := filepath.Split(name)
	parDir, err := vos.getDir(parent)
	if err != nil {
		return newPathError("open", parent, err)
	}
	if err := parDir.update(base, vFile{data}); err != nil {
		return newPathError("write", name, err)
	}
	return nil
}

func (vos vos) ReadFile(name string) ([]byte, error) {
	f, err := vos.getFile(name)
	if err != nil {
		return nil, newPathError("open", name, err)
	}
	return f.data, nil
}

func (vos vos) Rename(oldpath, newpath string) error {
	oldE, err := vos.get(oldpath)
	if err != nil {
		return newPathError("rename", oldpath, err)
	}

	oParent, oBase := filepath.Split(oldpath)
	oParDir, err := vos.getDir(oParent)
	if err != nil {
		return newPathError("rename", oldpath, err)
	}
	if !oParDir.has(oBase) {
		return newPathError("rename", oldpath, fs.ErrNotExist)
	}

	nParent, nBase := filepath.Split(newpath)
	nParDir, err := vos.getDir(nParent)
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

func (vos vos) Remove(name string) error {
	parent, base := filepath.Split(name)
	parDir, err := vos.getDir(parent)
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

func (vos vos) RemoveAll(name string) error {
	parent, base := filepath.Split(name)
	if parDir, err := vos.getDir(parent); err == nil {
		parDir.delete(base)
	}
	return nil
}

func (vos vos) Getwd() (dir string, err error) {
	return vos.pwd, nil
}

func (vos vos) UserCacheDir() (string, error) {
	return vos.usrCch, nil
}

func (vos vos) UserConfigDir() (string, error) {
	return vos.usrCfg, nil
}

func (vos vos) UserHomeDir() (string, error) {
	return vos.home, nil
}

func (vos) Exit(code int) {
	panic(exitCode(code))
}

func (vos vos) get(p string) (dirEntry, error) {
	dir := vos.entries
	var got dirEntry = dir
	var err error
	for _, name := range vos.splitPath(p) {
		got, err = dir.get(name)
		if err != nil {
			return nil, err
		}
		dir, _ = got.(vDir)
	}
	return got, nil
}

func (vos vos) getDir(p string) (vDir, error) {
	e, err := vos.get(p)
	if err != nil {
		return vDir{}, err
	}
	dir, ok := e.(vDir)
	if !ok {
		return vDir{}, errNotDir
	}
	return dir, nil
}

func (vos vos) getFile(p string) (vFile, error) {
	e, err := vos.get(p)
	if err != nil {
		return vFile{}, err
	}
	if file, ok := e.(vFile); !ok {
		return vFile{}, errNotFile
	} else {
		return file, nil
	}
}

// MkTempDir creates a new pseudo-temporary directory.
func MkTempDir(vos vos) string {
	path, err := vos.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	return path
}

// CatchExit allows recovering from vos.Exit() and calls catch with the denoted
// exit code.
func CatchExit(catch func(code int)) {
	if r := recover(); r != nil {
		if code, ok := r.(exitCode); ok {
			catch(int(code))
		} else {
			panic(r)
		}
	}
}

// splitPath splits a given path into a slice of path components.
func (vos vos) splitPath(p string) []string {
	sep := string(vos.PathSeparator())
	p = filepath.Clean(p)
	if p == sep {
		return nil
	} else if !strings.HasPrefix(p, sep) {
		panic(errors.New("unexpected relative path"))
	}
	return strings.Split(p, sep)[1:]
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

type exitCode int