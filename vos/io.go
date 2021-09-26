package vos

import (
	"bytes"
	"io"
)

type vosIO struct {
	stdin  *bytes.Buffer
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func newIO() vosIO {
	return vosIO{
		stdin:  new(bytes.Buffer),
		stdout: new(bytes.Buffer),
		stderr: new(bytes.Buffer),
	}
}

func (v vosIO) Stdin() io.Reader {
	return v.stdin
}
func (v vosIO) Stdout() io.Writer {
	return v.stdout
}
func (v vosIO) Stderr() io.Writer {
	return v.stderr
}

// GetStdio returns IO read-writers for Stdin, Stdout, and Stderr of the
// vos instance.
func GetStdio(v vos) (stdin, stdout, stderr io.ReadWriter) {
	return v.stdin, v.stdout, v.stderr
}

// ClearStdio clears IO read-writers for Stdin, Stdout, and Stderr of the
// vos instance.
func ClearStdio(v vos) {
	v.stdin.Reset()
	v.stdout.Reset()
	v.stderr.Reset()
}
