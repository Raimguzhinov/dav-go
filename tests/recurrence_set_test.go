package tests

import (
	"testing"

	"github.com/Raimguhinov/dav-go/tests/suite"
)

func TestRecurrence_EveryDay(t *testing.T) {
	ctx, st := suite.New(t, true)
	suite.CompareContentsByTestName(ctx, t, st)
}

func TestRecurrence_ByWorkdaysUntilNewYear(t *testing.T) {
	ctx, st := suite.New(t, true)
	suite.CompareContentsByTestName(ctx, t, st)
}
