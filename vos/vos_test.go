package vos_test

import (
	"fmt"
	"testing"

	"github.com/echocrow/osa"
	"github.com/echocrow/osa/testos"
	"github.com/echocrow/osa/vos"
	"github.com/stretchr/testify/assert"
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

	testos.AssertOsa(t, v, mkTempDir, assertExit)
}
