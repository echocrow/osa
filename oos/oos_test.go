package oos_test

import (
	"testing"

	"github.com/echocrow/osa/oos"
	"github.com/echocrow/osa/testos"
)

func TestOOS(t *testing.T) {
	o := oos.New()
	testos.AssertOrgOS(t, o)
}
