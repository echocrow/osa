package vos_test

import (
	"io"
	"testing"

	"github.com/echocrow/osa/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
