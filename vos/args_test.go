package vos_test

import (
	"os"
	"testing"

	"github.com/echocrow/osa/vos"
	"github.com/stretchr/testify/assert"
)

func TestPatchArgs(t *testing.T) {
	org := os.Args

	set, reset := vos.PatchArgs()

	assert.Equal(t, org, os.Args)

	args1 := []string{}
	set(args1)
	assert.Equal(t, args1, os.Args)

	args2 := []string{"foo", "bar"}
	set(args2)
	assert.Equal(t, args2, os.Args)

	reset()
	assert.Equal(t, org, os.Args)
}
