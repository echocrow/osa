package oos_test

import (
	"testing"

	"github.com/echocrow/osa/oos"
	"github.com/echocrow/osa/testosa"
)

func TestOOS(t *testing.T) {
	o := oos.New()
	testosa.AssertOrgOS(t, o)
}
