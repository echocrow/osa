package vos

import (
	"fmt"
	"io/fs"
	"path/filepath"

	os "github.com/echocrow/osa"
)

type vosFS struct {
	vfs
}

func newFS() vosFS {
	return vosFS{
		vfs: newVFS(),
	}
}

func (vosFS) PathSeparator() uint8 {
	return filepath.Separator
}

func (v vosFS) Stat(name string) (os.FileInfo, error) {
	return fs.Stat(v.vfs, name)
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

func (v vosFS) MkdirAll(name string, perm fs.FileMode) error {
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

func (v vosFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(v.vfs, name)
}

func (v vosFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	parent, base := filepath.Split(name)
	parDir, err := v.getDir(parent)
	if err != nil {
		return newPathError("open", parent, err)
	}
	if err := parDir.update(base, newVFile(data)); err != nil {
		return newPathError("write", name, err)
	}
	return nil
}

func (v vosFS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(v.vfs, name)
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
