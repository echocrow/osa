package vos

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

var (
	errNotDir   = errors.New("not a directory")
	errNotFile  = errors.New("not a file")
	errNotEmpty = errors.New("not empty")
	errOpFailed = errors.New("operation failed")
)

type vfs struct {
	temp string

	home   string
	usrCch string
	usrCfg string

	pwd string

	entries vDir
}

func newVFS() vfs {
	v := &vfs{
		entries: newVDir(),
	}

	sep := string(v.pathSeparator())
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

func (v vfs) Open(name string) (fs.File, error) {
	got, err := v.get(name)
	if err != nil {
		return nil, newPathError("open", name, err)
	}
	_, base := filepath.Split(name)
	return got.toFile(base)
}

func (v vfs) Mkdir(name string, perm fs.FileMode) error {
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

func (v vfs) get(p string) (dirEntry, error) {
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

func (v vfs) getDir(p string) (vDir, error) {
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

func (vfs) pathSeparator() uint8 {
	return filepath.Separator
}

// splitPath splits a given path into a slice of path components.
func (v vfs) splitPath(p string) []string {
	sep := string(v.pathSeparator())
	p = filepath.Clean(p)
	if p == sep {
		return nil
	} else if !strings.HasPrefix(p, sep) {
		panic(errors.New("unexpected relative path"))
	}
	return strings.Split(p, sep)[1:]
}
