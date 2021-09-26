package osa

import "io"

// Stdio describes a set of functions that return IO readers or writers for
// Stdin, Stdout, and Stderr.
type Stdio interface {
	// Stdin returns IO reader for Stdin.
	Stdin() io.Reader
	// Stdout returns IO writer for Stdout.
	Stdout() io.Writer
	// Stderr returns IO writer for Stderr.
	Stderr() io.Writer
}

// Stdin, Stdout, and Stderr are readers and writers for the standard input,
// standard output, and standard error streams.
var (
	Stdin  io.Reader = stdin{}
	Stdout io.Writer = stdout{}
	Stderr io.Writer = stderr{}
)

type stdin struct{}
type stdout struct{}
type stderr struct{}

func (stdin) Read(p []byte) (int, error)   { return osa.Stdin().Read(p) }
func (stdout) Write(p []byte) (int, error) { return osa.Stdout().Write(p) }
func (stderr) Write(p []byte) (int, error) { return osa.Stderr().Write(p) }
