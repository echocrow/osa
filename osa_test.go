package osa_test

import (
	"fmt"
	"testing"

	"github.com/echocrow/osa"
	"github.com/echocrow/osa/oos"
	"github.com/echocrow/osa/testos"
	"github.com/stretchr/testify/assert"
)

func TestCurrentDefault(t *testing.T) {
	o := osa.Current()
	testos.AssertOrgOS(t, o)
}

func TestCurrentAndPatch(t *testing.T) {
	tests := []osa.I{
		nil,
		oos.New(),
		osa.Current(),
	}
	for i, o := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			org := osa.Current()
			reset := osa.Patch(o)
			assert.Exactly(t, o, osa.Current(), "sets OS abstraction")
			reset()
			assert.Exactly(t, org, osa.Current(), "resets OS abstraction")
		})
	}
}
