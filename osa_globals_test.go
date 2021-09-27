package osa_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/echocrow/osa"
	"github.com/echocrow/osa/testos"
	"github.com/echocrow/osa/testosa"
)

//go:generate go run ./gen -name=gbl -call=osa -import=github.com/echocrow/osa
type FileInfo = osa.FileInfo
type FileMode = osa.FileMode
type DirEntry = osa.DirEntry

func TestGlobals(t *testing.T) {
	g := gbl{}
	testosa.AssertOrgOS(t, g)
}

type custStdioGbl struct {
	gbl
	stdio *bytes.Buffer
}

func (c custStdioGbl) Stdin() io.Reader  { return c.stdio }
func (c custStdioGbl) Stdout() io.Writer { return c.stdio }
func (c custStdioGbl) Stderr() io.Writer { return c.stdio }

func TestGlobalStdio(t *testing.T) {
	getStdout := func() io.Writer { return osa.Stdout }

	stdio := new(bytes.Buffer)
	cg := custStdioGbl{stdio: stdio}
	reset := osa.Patch(cg)
	defer reset()

	// Require that stdio works as expected before the actual assertion.
	testosa.AssertStdio(t, cg, stdio, stdio, stdio)

	stdout := getStdout()
	msg := "some custom stdio message"
	testos.AssertStdWrite(t, stdout, stdio, msg)
}
