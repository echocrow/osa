// Package vos provides the original OS abstraction implementation.
package oos

import (
	"os"
)

type oos struct{}

func New() oos {
	return oos{}
}

func (oos) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (oos) IsExist(err error) bool {
	return os.IsExist(err)
}

func (oos) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (oos) PathSeparator() uint8 {
	return os.PathSeparator
}

func (oos) IsPathSeparator(c uint8) bool {
	return os.IsPathSeparator(c)
}

func (oos) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (oos) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func (oos) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

func (oos) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func (oos) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (oos) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (oos) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (oos) Remove(name string) error {
	return os.Remove(name)
}

func (oos) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (oos) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (oos) UserCacheDir() (string, error) {
	return os.UserCacheDir()
}

func (oos) UserConfigDir() (string, error) {
	return os.UserConfigDir()
}

func (oos) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (oos) Exit(code int) {
	os.Exit(code)
}
