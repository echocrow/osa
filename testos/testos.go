// Package testos provides utilities for common OS assert/require test
// operations.
package testos

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	osaPkg "github.com/echocrow/osa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RequireEmptyWrite requires that writing an empty file succeeds.
func RequireEmptyWrite(t *testing.T, osa osaPkg.I, path string) {
	RequireWrite(t, osa, path, "")
}

// RequireWrite requires that writing to a file succeeds.
func RequireWrite(t *testing.T, osa osaPkg.I, path string, data string) {
	err := osa.WriteFile(path, []byte(data), 0600)
	require.NoError(t, err)
	RequireFileData(t, osa, path, data)
}

// RequireMkdir requires that creating a directory succeeds.
func RequireMkdir(t *testing.T, osa osaPkg.I, path string) {
	err := osa.Mkdir(path, 0700)
	require.NoError(t, err)
	RequireExists(t, osa, path)
}

// RequireMkdirAll requires that creating a directory (and possible parent
// directories) succeeds.
func RequireMkdirAll(t *testing.T, osa osaPkg.I, path string) {
	err := osa.MkdirAll(path, 0700)
	require.NoError(t, err)
	RequireExists(t, osa, path)
}

// AssertExists asserts that a file or directory exists.
func AssertExists(t *testing.T, osa osaPkg.I, path string) bool {
	return assert.Truef(t, exists(osa, path), "expected %s to exist", path)
}

// AssertNotExists asserts that a file or directory does not exist.
func AssertNotExists(t *testing.T, osa osaPkg.I, path string) bool {
	return assert.Falsef(t, exists(osa, path), "expected %s to not exist", path)
}

// AssertExistsIsDir asserts that an entry exists and is either file or directory.
func AssertExistsIsDir(t *testing.T, osa osaPkg.I, path string, wantDir bool) bool {
	ok := true
	gotDir, err := existsIsDir(osa, path)
	wantName := "file"
	if wantDir {
		wantName = "directory"
	}
	ok = ok && assert.NoError(t, err, "expected %s to exist", path)
	ok = ok && assert.Equal(t, wantDir, gotDir, "expected %s to be a %s", path, wantName)
	return ok
}

// RequireExists requires that a file or directory exists.
func RequireExists(t *testing.T, osa osaPkg.I, path string) {
	require.Truef(t, exists(osa, path), "expected %s to exist", path)
}

// RequireNotExists requires that a file or directory does not exist.
func RequireNotExists(t *testing.T, osa osaPkg.I, path string) {
	require.Falsef(t, exists(osa, path), "expected %s to not exist", path)
}

// RequireExistsIsDir requires that an entry exists and is either file or directory.
func RequireExistsIsDir(t *testing.T, osa osaPkg.I, path string, wantDir bool) {
	gotDir, err := existsIsDir(osa, path)
	wantName := "file"
	if wantDir {
		wantName = "directory"
	}
	require.NoError(t, err, "expected %s to exist", path)
	require.Equal(t, wantDir, gotDir, "expected %s to be a %s", path, wantName)
}

// AssertFileData requires that a file exists and has a certain data
func AssertFileData(t *testing.T, osa osaPkg.I, path string, want string) bool {
	ok := true
	got, err := osa.ReadFile(path)
	ok = ok && assert.NoError(t, err)
	ok = ok && assert.Equal(t, want, string(got))
	return ok
}

// RequireFileData requires that a file exists and has a certain data
func RequireFileData(t *testing.T, osa osaPkg.I, path string, want string) {
	got, err := osa.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, want, string(got))
}

// AssertIsEmpty asserts that a file or directory is empty.
func AssertIsEmpty(t *testing.T, osa osaPkg.I, path string) bool {
	ok := true
	empty, err := isEmpty(osa, path)
	ok = ok && assert.NoError(t, err, "expected %s to exist and be readable", path)
	ok = ok && assert.True(t, empty, "expected %s to be empty", path)
	return ok
}

// RequireIsEmpty requires that a file or directory is empty.
func RequireIsEmpty(t *testing.T, osa osaPkg.I, path string) {
	empty, err := isEmpty(osa, path)
	require.NoError(t, err, "expected %s to exist and be readable", path)
	require.True(t, empty, "expected %s to be empty", path)
}

// RequireTempDir requires that a temp dir was created.
func RequireTempDir(t *testing.T, osa osaPkg.I) string {
	tmpDir, err := osa.MkdirTemp("", "")
	require.NoError(t, err)
	return tmpDir
}

// Join joins any number of path elements into a single path.
func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// AssertStdWrite tests a writable standard stream (i.e. Stdout or Stderr).
func AssertStdWrite(
	t *testing.T,
	stdWrite io.Writer,
	stdRead io.Reader,
	message string,
) bool {
	ok := true

	_, err := fmt.Fprintln(stdWrite, message)
	ok = ok && assert.NoError(t, err)

	reader := bufio.NewReader(stdRead)
	gotLine, _, err := reader.ReadLine()
	ok = ok && assert.NoError(t, err)
	got := string(gotLine)
	ok = ok && assert.Equal(t, message, got)

	return ok
}

// exists returns a boolean indicating whether a file or directory exists.
func exists(osa osaPkg.I, path string) bool {
	_, err := osa.Stat(path)
	return !osa.IsNotExist(err)
}

// existsIsDir returns a boolean indicating whether an entry is a file or
// directory, or returns an error if the entry does not exists.
func existsIsDir(osa osaPkg.I, path string) (isDir bool, err error) {
	fi, err := osa.Stat(path)
	isDir = fi != nil && fi.IsDir()
	return
}

// isEmpty returns true when a given file or directry is empty.
func isEmpty(osa osaPkg.I, path string) (isEmpty bool, err error) {
	fi, err := osa.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		es, err := osa.ReadDir(path)
		if err == io.EOF {
			err = nil
		}
		return len(es) == 0, err
	}
	return fi.Size() == 0, nil
}
