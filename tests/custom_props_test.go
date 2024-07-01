package tests

import (
	"testing"

	"github.com/Raimguzhinov/dav-go/tests/suite"
)

func TestCustomProps_WeakTypes(t *testing.T) {
	ctx, st := suite.New(t, true)
	suite.CompareContentsByTestName(ctx, t, st)
}

func TestCustomProps_StrongTypes(t *testing.T) {
	ctx, st := suite.New(t, true)
	suite.CompareContentsByTestName(ctx, t, st)
}
