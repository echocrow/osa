package osa_test

import (
	"testing"

	"github.com/echocrow/osa"
	"github.com/echocrow/osa/testos"
)

type gbl struct{}

func (gbl) Stat(name string) (osa.FileInfo, error) { return osa.Stat(name) }

func (gbl) IsExist(err error) bool { return osa.IsExist(err) }

func (gbl) IsNotExist(err error) bool { return osa.IsNotExist(err) }

func (gbl) PathSeparator() uint8 { return osa.PathSeparator() }

func (gbl) IsPathSeparator(c uint8) bool { return osa.IsPathSeparator(c) }

func (gbl) Mkdir(name string, perm osa.FileMode) error {
	return osa.Mkdir(name, perm)
}

func (gbl) MkdirAll(name string, perm osa.FileMode) error {
	return osa.MkdirAll(name, perm)
}

func (gbl) MkdirTemp(dir, pattern string) (string, error) {
	return osa.MkdirTemp(dir, pattern)
}

func (gbl) ReadDir(name string) ([]osa.DirEntry, error) {
	return osa.ReadDir(name)
}

func (gbl) WriteFile(name string, data []byte, perm osa.FileMode) error {
	return osa.WriteFile(name, data, perm)
}

func (gbl) ReadFile(name string) ([]byte, error) { return osa.ReadFile(name) }

func (gbl) Rename(oldpath, newpath string) error {
	return osa.Rename(oldpath, newpath)
}

func (gbl) Remove(name string) error { return osa.Remove(name) }

func (gbl) RemoveAll(path string) error { return osa.RemoveAll(path) }

func (gbl) Getwd() (dir string, err error) { return osa.Getwd() }

func (gbl) UserCacheDir() (string, error) { return osa.UserCacheDir() }

func (gbl) UserConfigDir() (string, error) { return osa.UserConfigDir() }

func (gbl) UserHomeDir() (string, error) { return osa.UserHomeDir() }

func (gbl) Exit(code int) { osa.Exit(code) }

func TestGlobals(t *testing.T) {
	g := gbl{}
	testos.AssertOrgOS(t, g)
}
