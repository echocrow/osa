package vos_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/echocrow/osa"
	"github.com/echocrow/osa/testos"
	"github.com/echocrow/osa/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatch(t *testing.T) {
	org := osa.Current()
	v, reset := vos.Patch()
	assert.Exactly(t, v, osa.Current())
	reset()
	assert.Exactly(t, org, osa.Current())
}

func TestVos(t *testing.T) {
	v := vos.New()

	mkTempDir := func() string { return vos.MkTempDir(v) }

	assertExit := func(t *testing.T) {
		tests := []int{0, 1, 34}
		for _, code := range tests {
			want := code
			t.Run(fmt.Sprint(code), func(t *testing.T) {
				defer vos.CatchExit(func(got int) {
					assert.Equal(t, want, got)
				})
				v.Exit(code)
				t.Fatal("shoul have exited")
			})
		}
	}

	getStdio := func() (in io.Writer, out, err io.Reader, reset func()) {
		in, out, err = vos.GetStdio(v)
		return
	}

	testos.AssertOsa(t, v, mkTempDir, assertExit, getStdio)
}

func TestGetStdio(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	stdin, stdout, stderr := vos.GetStdio(v)

	tests := []struct {
		n string
		w io.Writer
		r io.Reader
	}{
		{"stdin", stdin, v.Stdin()},
		{"stdout", v.Stdout(), stdout},
		{"stderr", v.Stderr(), stderr},
	}
	for _, tc := range tests {
		t.Run(tc.n, func(t *testing.T) {
			assertWriteReadPipe(t, tc.w, tc.r)
		})
	}
}

func TestClearStdio(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	stdin, stdout, stderr := vos.GetStdio(v)

	pipes := []struct {
		n string
		w io.Writer
		r io.Reader
	}{
		{"stdin", stdin, v.Stdin()},
		{"stdout", v.Stdout(), stdout},
		{"stderr", v.Stderr(), stderr},
	}

	data := []byte("some text\n")
	for _, p := range pipes {
		p.w.Write(data)
	}

	vos.ClearStdio(v)

	for _, p := range pipes {
		t.Run(p.n, func(t *testing.T) {
			assertEmptyReader(t, p.r)
		})
	}
}

func assertWriteReadPipe(t *testing.T, w io.Writer, r io.Reader) {
	data := []byte("some text\n")
	dataLen := len(data)

	prevLen, prevErr := r.Read(make([]byte, 1))
	require.Empty(t, prevLen, "expected reader to be empty")
	require.Equal(t, io.EOF, prevErr, "expected reader to be empty")

	wLen, wErr := w.Write(data)
	require.Equal(t, dataLen, wLen, "expected write to succeed")
	require.NoError(t, wErr, "expected write to succeed")

	read := make([]byte, dataLen)
	rLen, rErr := r.Read(read)
	assert.Equal(t, data, read, "expected to read entire data")
	require.Equal(t, dataLen, rLen, "expected to read entire data")
	require.NoError(t, rErr, "expected to read entire data")

	assertEmptyReader(t, r)
}

func assertEmptyReader(t *testing.T, r io.Reader) {
	len, err := r.Read(make([]byte, 1))
	require.Empty(t, len, "expected reader to be empty")
	require.Equal(t, io.EOF, err, "expected reader to be empty")
}
