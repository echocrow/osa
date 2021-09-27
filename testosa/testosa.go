// Package testosa provides tests for OSA implementations.
package testosa

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	osaPkg "github.com/echocrow/osa"
	tos "github.com/echocrow/osa/testos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertOsa tests various internal OSA operations.
func AssertOsa(
	t *testing.T,
	osa osaPkg.I,
	mkTempDir func() string,
	assertExit func(t *testing.T),
	getStdio func() (in io.Writer, out, err io.Reader, reset func()),
) {
	t.Run("OpenDir", func(t *testing.T) {
		tmpDir := mkTempDir()

		dirBasename := "some-dir"
		dirname := tos.Join(tmpDir, dirBasename)
		tos.RequireMkdir(t, osa, dirname)

		tos.RequireMkdir(t, osa, tos.Join(dirname, "aSubdir"))
		tos.RequireMkdir(t, osa, tos.Join(dirname, "someOtherSubdir"))
		tos.RequireEmptyWrite(t, osa, tos.Join(dirname, "file.txt"))

		got, err := osa.Open(dirname)
		assert.NoError(t, err)
		require.NotNil(t, got)

		gotStat, err := got.Stat()
		assert.NoError(t, err)
		assert.True(t, gotStat.IsDir(), "expect dir, not file")
		assert.Equal(t, dirBasename, gotStat.Name())

		gotConts := make([]byte, 1)
		gotLen, err := got.Read(gotConts)
		assert.Zero(t, gotLen)
		assert.Error(t, err)

		gotDir, ok := got.(fs.ReadDirFile)
		assert.True(t, ok)
		require.NotNil(t, gotDir)

		gotEntries, err := gotDir.ReadDir(-1)
		assert.NoError(t, err)
		gotFSEntries := castFsEntries(gotEntries, true)
		assert.Equal(t, []fsEntry{
			{"aSubdir", true},
			{"file.txt", false},
			{"someOtherSubdir", true},
		}, gotFSEntries)
	})
	t.Run("OpenDirPartialRead", func(t *testing.T) {
		tmpDir := mkTempDir()

		tos.RequireMkdir(t, osa, tos.Join(tmpDir, "subdir"))
		tos.RequireMkdir(t, osa, tos.Join(tmpDir, "subdir", "deepDir"))
		tos.RequireMkdir(t, osa, tos.Join(tmpDir, "anotherSubdir"))
		tos.RequireMkdir(t, osa, tos.Join(tmpDir, "zzDir"))
		tos.RequireEmptyWrite(t, osa, tos.Join(tmpDir, "someFile"))
		tos.RequireEmptyWrite(t, osa, tos.Join(tmpDir, "aaFile"))

		gotFile, err := osa.Open(tmpDir)
		require.NoError(t, err)
		gotDir, ok := gotFile.(fs.ReadDirFile)
		require.True(t, ok)

		wantAll := []fsEntry{
			{"aaFile", false},
			{"anotherSubdir", true},
			{"someFile", false},
			{"subdir", true},
			{"zzDir", true},
		}

		got1, err := gotDir.ReadDir(3)
		assert.NoError(t, err)
		assert.Len(t, got1, 3)

		got2, err := gotDir.ReadDir(1)
		assert.NoError(t, err)
		assert.Len(t, got2, 1)

		got3, err := gotDir.ReadDir(2)
		assert.NoError(t, err)
		assert.Len(t, got3, 1)

		got4, err := gotDir.ReadDir(2)
		assert.ErrorIs(t, err, io.EOF)
		assert.Empty(t, got4)

		gotAll := make([]fs.DirEntry, 0, len(wantAll))
		for _, g := range [][]fs.DirEntry{got1, got2, got3, got4} {
			gotAll = append(gotAll, g...)
		}

		assert.Equal(t, wantAll, castFsEntries(gotAll, true))
	})
	t.Run("OpenDirReadErrClosed", func(t *testing.T) {
		tmpDir := mkTempDir()
		tos.RequireEmptyWrite(t, osa, tos.Join(tmpDir, "someDir"))

		gotFile, err := osa.Open(tmpDir)
		require.NoError(t, err)

		gotDir, ok := gotFile.(fs.ReadDirFile)
		require.True(t, ok)

		err = gotDir.Close()
		require.NoError(t, err)

		gotEntries, err := gotDir.ReadDir(-1)
		assert.Empty(t, gotEntries)
		assert.Error(t, err)
	})
	t.Run("OpenFile", func(t *testing.T) {
		tmpDir := mkTempDir()

		fileBasename := "some-file.foo"
		filename := tos.Join(tmpDir, fileBasename)
		data := []byte("some data")
		filelen := len(data)
		tos.RequireWrite(t, osa, filename, string(data))

		got, err := osa.Open(filename)
		assert.NoError(t, err)
		require.NotNil(t, got)

		gotStat, err := got.Stat()
		assert.NoError(t, err)
		assert.False(t, gotStat.IsDir(), "expect file, not dir")
		assert.Equal(t, fileBasename, gotStat.Name())
		assert.Equal(t, int64(filelen), gotStat.Size())

		wantConts := append(data, 0)
		gotConts := make([]byte, filelen+1)
		gotLen, err := got.Read(gotConts)
		assert.NoError(t, err)
		assert.Equal(t, filelen, gotLen)
		assert.Equal(t, wantConts, gotConts)

		err = got.Close()
		assert.NoError(t, err)
		err = got.Close()
		assert.Error(t, err)

		gotLen2, err := got.Read(gotConts)
		assert.Zero(t, gotLen2)
		assert.Error(t, err)
	})
	t.Run("OpenFilePartialRead", func(t *testing.T) {
		tmpDir := mkTempDir()

		filename := tos.Join(tmpDir, "someFile")
		data := []byte("some more data")
		tos.RequireWrite(t, osa, filename, string(data))

		got, err := osa.Open(filename)
		require.NoError(t, err)
		require.NotNil(t, got)

		wantConts1 := data[0:4]
		gotConts1 := make([]byte, len(wantConts1))
		gotLen1, err := got.Read(gotConts1)
		assert.NoError(t, err)
		assert.Equal(t, len(wantConts1), gotLen1)
		assert.Equal(t, wantConts1, gotConts1)

		wantConts2 := data[len(wantConts1):]
		gotConts2 := make([]byte, len(wantConts2))
		gotLen2, err := got.Read(gotConts2)
		assert.NoError(t, err)
		assert.Equal(t, len(wantConts2), gotLen2)
		assert.Equal(t, wantConts2, gotConts2)

		wantConts3 := make([]byte, 1)
		gotConts3 := make([]byte, len(wantConts3))
		gotLen3, err := got.Read(gotConts3)
		assert.ErrorIs(t, err, io.EOF)
		assert.Zero(t, gotLen3)
		assert.Equal(t, wantConts3, gotConts3)

		err = got.Close()
		assert.NoError(t, err)
	})
	t.Run("OpenFileReadErrClosed", func(t *testing.T) {
		tmpDir := mkTempDir()
		path := tos.Join(tmpDir, "testFile")
		tos.RequireWrite(t, osa, path, "some file data")

		got, err := osa.Open(path)
		require.NoError(t, err)

		err = got.Close()
		require.NoError(t, err)

		gotConts := make([]byte, 1)
		gotRead, err := got.Read(gotConts)
		assert.Empty(t, gotRead)
		assert.Equal(t, make([]byte, 1), gotConts)
		assert.Error(t, err)
	})
	t.Run("OpenErrNotExist", func(t *testing.T) {
		tmpDir := mkTempDir()

		missingFilename := tos.Join(tmpDir, "missing")
		tos.RequireNotExists(t, osa, missingFilename)

		_, err := osa.Open(missingFilename)
		assert.Error(t, err)
		assert.True(t, osa.IsNotExist(err), "want not-exist error")
	})

	t.Run("StatDir", func(t *testing.T) {
		tmpDir := mkTempDir()

		testDirName := "someDir"
		testDir := tos.Join(tmpDir, testDirName)
		tos.RequireMkdir(t, osa, testDir)

		stat, err := osa.Stat(testDir)
		require.NotNil(t, stat)
		assert.Equal(t, testDirName, stat.Name())
		assert.True(t, stat.IsDir())
		assert.NoError(t, err)
	})
	t.Run("StatFile", func(t *testing.T) {
		tmpDir := mkTempDir()

		testFileName := "myFile"
		testFile := tos.Join(tmpDir, testFileName)
		tos.RequireEmptyWrite(t, osa, testFile)

		stat, err := osa.Stat(testFile)
		require.NotNil(t, stat)
		assert.Equal(t, testFileName, stat.Name())
		assert.False(t, stat.IsDir())
		assert.NoError(t, err)
	})
	t.Run("StatErr", func(t *testing.T) {
		tmpDir := mkTempDir()
		var err error

		missingDirName := "dirDoesntExist"
		missingDir := tos.Join(tmpDir, missingDirName)
		_, err = osa.Stat(missingDir)
		assert.Error(t, err)

		missingFileName := "fileDoesntExist"
		missingFile := tos.Join(tmpDir, missingFileName)
		_, err = osa.Stat(missingFile)
		assert.Error(t, err)
	})

	t.Run("IsExist", func(t *testing.T) {
		tmpDir := mkTempDir()

		newDir := tos.Join(tmpDir, "new")
		tos.RequireMkdir(t, osa, newDir)

		var err error

		err = osa.Mkdir(newDir, 0700)
		require.Error(t, err)
		assert.True(t, osa.IsExist(err), "expect IsExist err")

		err = osa.Mkdir(tos.Join(tmpDir, "doesnt", "exist"), 0700)
		require.Error(t, err)
		assert.False(t, osa.IsExist(err), "unexpected IsExist err")
	})

	t.Run("IsNotExist", func(t *testing.T) {
		tmpDir := mkTempDir()

		existingDir := tos.Join(tmpDir, "exists")
		tos.RequireMkdir(t, osa, existingDir)

		var err error

		_, err = osa.Stat(existingDir)
		assert.False(t, osa.IsNotExist(err))

		_, err = osa.Stat(tos.Join(tmpDir, "doesnt", "exist"))
		assert.True(t, osa.IsNotExist(err))
	})

	t.Run("IsPathSeparator", func(t *testing.T) {
		sep := osa.PathSeparator()
		notSep := uint8('a')
		if notSep == sep {
			notSep = uint8('b')
		}

		assert.True(t, osa.IsPathSeparator(sep))
		assert.False(t, osa.IsPathSeparator(notSep))
	})

	t.Run("Mkdir", func(t *testing.T) {
		tmpDir := mkTempDir()

		newDir := tos.Join(tmpDir, "new")
		tos.RequireNotExists(t, osa, newDir)

		err := osa.Mkdir(newDir, 0700)
		assert.NoError(t, err)
		tos.AssertExists(t, osa, newDir)
	})
	t.Run("MkdirErrExists", func(t *testing.T) {
		tmpDir := mkTempDir()

		newDir := tos.Join(tmpDir, "new")
		tos.RequireMkdir(t, osa, newDir)

		err := osa.Mkdir(newDir, 0700)
		assert.Error(t, err)
	})

	t.Run("MkdirAllSingle", func(t *testing.T) {
		tmpDir := mkTempDir()

		newDir := tos.Join(tmpDir, "new")
		tos.RequireNotExists(t, osa, newDir)

		err := osa.MkdirAll(newDir, 0700)
		assert.NoError(t, err)
		tos.AssertExists(t, osa, newDir)
	})
	t.Run("MkdirAllNested", func(t *testing.T) {
		tmpDir := mkTempDir()

		parentDir := tos.Join(tmpDir, "parent")
		tos.RequireNotExists(t, osa, parentDir)
		subDir := tos.Join(parentDir, "sub", "dir")
		tos.RequireNotExists(t, osa, subDir)

		err := osa.MkdirAll(subDir, 0700)
		assert.NoError(t, err)
		tos.AssertExists(t, osa, subDir)
	})
	t.Run("MkdirAllErrExists", func(t *testing.T) {
		tmpDir := mkTempDir()

		parent := tos.Join(tmpDir, "parent")
		tos.RequireNotExists(t, osa, parent)
		subDir := tos.Join(parent, "sub", "dir")
		tos.RequireNotExists(t, osa, subDir)

		tos.RequireEmptyWrite(t, osa, parent)

		err := osa.MkdirAll(subDir, 0700)
		assert.Error(t, err)
	})

	t.Run("MkdirTemp", func(t *testing.T) {
		tmpDir := mkTempDir()
		pattern := "myTmpDir"

		myDir, err := osa.MkdirTemp(tmpDir, pattern)
		assert.NoError(t, err)
		tos.AssertExists(t, osa, myDir)

		parent, base := filepath.Split(myDir)
		assert.Equal(t, filepath.Clean(tmpDir), filepath.Clean(parent))
		assert.True(t, strings.HasPrefix(base, pattern))
	})
	t.Run("MkdirTempUnique", func(t *testing.T) {
		tmpDir := mkTempDir()
		pattern := "myTmpDir"

		tests := 4
		prevDirs := make(map[string]struct{}, tests)
		for i := 0; i < tests; i++ {
			myDir, err := osa.MkdirTemp(tmpDir, pattern)
			require.NoError(t, err)
			_, exists := prevDirs[myDir]
			assert.False(t, exists)
			prevDirs[myDir] = struct{}{}
		}
		require.Len(t, prevDirs, tests)
	})
	t.Run("MkdirTempErrNotExist", func(t *testing.T) {
		tmpDir := mkTempDir()
		missingDir := tos.Join(tmpDir, "missingFolder")
		pattern := "myTmpDir"

		_, err := osa.MkdirTemp(missingDir, pattern)
		assert.Error(t, err)
		assert.True(t, osa.IsNotExist(err))
	})
	t.Run("MkdirTempErrInvalidPattern", func(t *testing.T) {
		tmpDir := mkTempDir()
		badPattern := "with/Sep"

		_, err := osa.MkdirTemp(tmpDir, badPattern)
		assert.Error(t, err)
	})

	t.Run("ReadDir", func(t *testing.T) {
		tmpDir := mkTempDir()
		tos.RequireEmptyWrite(t, osa, tos.Join(tmpDir, "zLastFile"))
		subDir := tos.Join(tmpDir, "fooFolder")
		tos.RequireMkdir(t, osa, subDir)
		tos.RequireEmptyWrite(t, osa, tos.Join(subDir, "subFile"))
		tos.RequireEmptyWrite(t, osa, tos.Join(tmpDir, "aFirstFile"))

		want := []fsEntry{
			{"aFirstFile", false},
			{"fooFolder", true},
			{"zLastFile", false},
		}

		gotEntries, err := osa.ReadDir(tmpDir)
		got := castFsEntries(gotEntries, false)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("ReadDirErrMissing", func(t *testing.T) {
		tmpDir := mkTempDir()
		testDir := tos.Join(tmpDir, "missingFolder")
		_, err := osa.ReadDir(testDir)
		assert.Error(t, err)
	})

	t.Run("WriteFile", func(t *testing.T) {
		tmpDir := mkTempDir()
		path := tos.Join(tmpDir, "testFile")
		want := []byte(`some file data`)

		err := osa.WriteFile(path, want, 0600)
		assert.NoError(t, err)

		got, err := osa.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("WriteFileErr", func(t *testing.T) {
		tmpDir := mkTempDir()
		path := tos.Join(tmpDir, "invalid", "path")
		err := osa.WriteFile(path, []byte{}, 0600)
		assert.Error(t, err)
	})

	t.Run("ReadFile", func(t *testing.T) {
		tmpDir := mkTempDir()
		path := tos.Join(tmpDir, "testFile")
		want := []byte(`some file data`)
		err := osa.WriteFile(path, want, 0600)
		require.NoError(t, err)

		got, err := osa.ReadFile(path)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("ReadFileErrMissing", func(t *testing.T) {
		tmpDir := mkTempDir()
		path := tos.Join(tmpDir, "missingFile")
		_, err := osa.ReadFile(path)
		assert.Error(t, err)
	})

	t.Run("RenameFile", func(t *testing.T) {
		tmpDir := mkTempDir()

		file := tos.Join(tmpDir, "testA")
		fileData := "dataA"
		tos.RequireWrite(t, osa, file, fileData)

		subDir := tos.Join(tmpDir, "sub")
		tos.RequireMkdir(t, osa, subDir)

		tests := []string{
			tos.Join(tmpDir, "testB"),
			tos.Join(subDir, "testC"),
		}
		for i, newFile := range tests {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				tos.RequireNotExists(t, osa, newFile)
				err := osa.Rename(file, newFile)
				assert.NoError(t, err)
				tos.AssertNotExists(t, osa, file)
				tos.AssertFileData(t, osa, newFile, fileData)

				file = newFile
			})
		}
	})
	t.Run("RenameDir", func(t *testing.T) {
		tmpDir := mkTempDir()

		oldDir := tos.Join(tmpDir, "oldDir")
		tos.RequireMkdir(t, osa, oldDir)
		oldSubDirName := "mySubDir"
		oldSubDir := tos.Join(oldDir, oldSubDirName)
		tos.RequireMkdir(t, osa, oldSubDir)

		newDir := tos.Join(tmpDir, "newDir")
		tos.RequireNotExists(t, osa, newDir)
		newSubDir := tos.Join(newDir, oldSubDirName)
		tos.RequireNotExists(t, osa, newSubDir)

		err := osa.Rename(oldDir, newDir)
		assert.NoError(t, err)
		tos.AssertNotExists(t, osa, oldDir)
		tos.AssertNotExists(t, osa, oldSubDir)
		tos.AssertExists(t, osa, newDir)
		tos.AssertExists(t, osa, newSubDir)
	})
	t.Run("RenameFileReplace", func(t *testing.T) {
		tmpDir := mkTempDir()

		file := tos.Join(tmpDir, "testA")
		fileData := "dataA"
		tos.RequireWrite(t, osa, file, fileData)

		existingFile := tos.Join(tmpDir, "collision")
		tos.RequireWrite(t, osa, existingFile, "other content")

		err := osa.Rename(file, existingFile)
		assert.NoError(t, err)
		tos.AssertNotExists(t, osa, file)
		tos.AssertFileData(t, osa, existingFile, fileData)
	})
	t.Run("RenameFileErrDirExists", func(t *testing.T) {
		tmpDir := mkTempDir()

		oldFile := tos.Join(tmpDir, "testA")
		tos.RequireEmptyWrite(t, osa, oldFile)

		existingDir := tos.Join(tmpDir, "collision")
		tos.RequireMkdir(t, osa, existingDir)

		err := osa.Rename(oldFile, existingDir)
		assert.Error(t, err)
		assert.True(t, osa.IsExist(err), "expect IsExist err")
	})
	t.Run("RenameDirErrDirExists", func(t *testing.T) {
		tmpDir := mkTempDir()

		oldDir := tos.Join(tmpDir, "testA")
		tos.RequireMkdir(t, osa, oldDir)

		existingDir := tos.Join(tmpDir, "collision")
		tos.RequireMkdir(t, osa, existingDir)

		err := osa.Rename(oldDir, existingDir)
		assert.Error(t, err)
		assert.True(t, osa.IsExist(err), "expect IsExist err")
	})

	t.Run("RemoveFile", func(t *testing.T) {
		tmpDir := mkTempDir()
		testFile := tos.Join(tmpDir, "tmp")
		tos.RequireEmptyWrite(t, osa, testFile)

		err := osa.Remove(testFile)
		tos.AssertNotExists(t, osa, testFile)
		assert.NoError(t, err)
	})
	t.Run("RemoveEmptyDir", func(t *testing.T) {
		tmpDir := mkTempDir()
		testDir := tos.Join(tmpDir, "ripDir")
		tos.RequireMkdir(t, osa, testDir)
		tos.RequireExists(t, osa, tmpDir)

		err := osa.Remove(testDir)
		tos.AssertNotExists(t, osa, testDir)
		assert.NoError(t, err)
	})
	t.Run("RemoveErrNotExist", func(t *testing.T) {
		tmpDir := mkTempDir()
		path := tos.Join(tmpDir, "doesnt", "exist")
		err := osa.Remove(path)
		assert.Error(t, err)
	})
	t.Run("RemoveErrDirWithFile", func(t *testing.T) {
		tmpDir := mkTempDir()
		testDir := tos.Join(tmpDir, "ripDir")
		tos.RequireMkdir(t, osa, testDir)
		tos.RequireExists(t, osa, tmpDir)
		tos.RequireEmptyWrite(t, osa, tos.Join(testDir, "someFile"))

		err := osa.Remove(testDir)
		assert.Error(t, err)
	})
	t.Run("RemoveErrDirWithDir", func(t *testing.T) {
		tmpDir := mkTempDir()
		testDir := tos.Join(tmpDir, "ripDir")
		tos.RequireMkdir(t, osa, testDir)
		tos.RequireExists(t, osa, tmpDir)
		tos.RequireMkdir(t, osa, tos.Join(testDir, "subDir"))

		err := osa.Remove(testDir)
		assert.Error(t, err)
	})

	t.Run("RemoveAllFile", func(t *testing.T) {
		tmpDir := mkTempDir()
		testFile := tos.Join(tmpDir, "tmp")
		tos.RequireEmptyWrite(t, osa, testFile)

		err := osa.RemoveAll(testFile)
		tos.AssertNotExists(t, osa, testFile)
		assert.NoError(t, err)
	})
	t.Run("RemoveAllDir", func(t *testing.T) {
		tmpDir := mkTempDir()
		testDir := tos.Join(tmpDir, "ripDir")
		err := osa.Mkdir(testDir, 0700)
		require.NoError(t, err)
		tos.RequireExists(t, osa, tmpDir)

		err = osa.RemoveAll(testDir)
		tos.AssertNotExists(t, osa, testDir)
		assert.NoError(t, err)
	})
	t.Run("RemoveAllNotExist", func(t *testing.T) {
		tmpDir := mkTempDir()
		path := tos.Join(tmpDir, "doesnt", "exist")
		err := osa.RemoveAll(path)
		assert.NoError(t, err)
	})
	t.Run("RemoveAllDirWithFile", func(t *testing.T) {
		tmpDir := mkTempDir()
		testDir := tos.Join(tmpDir, "ripDir")
		tos.RequireMkdir(t, osa, testDir)
		tos.RequireExists(t, osa, tmpDir)
		tos.RequireEmptyWrite(t, osa, tos.Join(testDir, "someFile"))

		err := osa.RemoveAll(testDir)
		assert.NoError(t, err)
	})
	t.Run("RemoveAllDirWithDir", func(t *testing.T) {
		tmpDir := mkTempDir()
		testDir := tos.Join(tmpDir, "ripDir")
		tos.RequireMkdir(t, osa, testDir)
		tos.RequireExists(t, osa, tmpDir)
		tos.RequireMkdir(t, osa, tos.Join(testDir, "subDir"))

		err := osa.RemoveAll(testDir)
		assert.NoError(t, err)
	})

	t.Run("Getwd", func(t *testing.T) {
		workDir, err := osa.Getwd()
		tos.AssertExists(t, osa, workDir)
		assert.NoError(t, err)
	})
	t.Run("UserCacheDir", func(t *testing.T) {
		cacheDir, err := osa.UserCacheDir()
		tos.AssertExists(t, osa, cacheDir)
		assert.NoError(t, err)
	})
	t.Run("UserConfigDir", func(t *testing.T) {
		cacheDir, err := osa.UserConfigDir()
		tos.AssertExists(t, osa, cacheDir)
		assert.NoError(t, err)
	})
	t.Run("UserHomeDir", func(t *testing.T) {
		homeDir, err := osa.UserHomeDir()
		tos.AssertExists(t, osa, homeDir)
		assert.NoError(t, err)
	})

	t.Run("Exit", assertExit)

	t.Run("Stdio", func(t *testing.T) {
		wIn, rOut, rErr, resetStdio := getStdio()
		if resetStdio != nil {
			defer resetStdio()
		}
		AssertStdio(t, osa, wIn, rOut, rErr)
	})
}

// AssertStdio tests Stdin, Stdout, and Stderr of an OS abstraction.
func AssertStdio(
	t *testing.T,
	osa osaPkg.I,
	wIn io.Writer,
	rOut io.Reader,
	rErr io.Reader,
) bool {
	ok := true
	stdin, stdout, stderr := osa.Stdin(), osa.Stdout(), osa.Stderr()

	ok = ok && t.Run("Stdin", func(t *testing.T) {
		msg := "some input message"
		_, err := io.WriteString(wIn, msg+"\n")
		require.NoError(t, err)

		reader := bufio.NewReader(stdin)
		gotLine, _, err := reader.ReadLine()
		assert.NoError(t, err)
		got := string(gotLine)
		assert.Equal(t, msg, got)
	})

	ok = ok && t.Run("Stdout", func(t *testing.T) {
		msg := "some output message"
		tos.AssertStdWrite(t, stdout, rOut, msg)
	})

	ok = ok && t.Run("Stderr", func(t *testing.T) {
		msg := "some error message"
		tos.AssertStdWrite(t, stderr, rErr, msg)
	})

	return ok
}

// fsEntry represents a simplified fs.DirEntry
type fsEntry struct {
	name  string
	isDir bool
}

// castFsEntries casts fs.DirEntry to basic fsEntry for comparison.
func castFsEntries(entries []fs.DirEntry, doSort bool) []fsEntry {
	es := make([]fsEntry, len(entries))
	for i, e := range entries {
		es[i] = fsEntry{e.Name(), e.IsDir()}
	}
	if doSort {
		sort.Slice(es, func(i, j int) bool { return es[i].name < es[j].name })
	}
	return es
}
