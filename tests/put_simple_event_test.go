package tests

import (
	"path"
	"testing"

	"github.com/Raimguhinov/dav-go/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutSimpleEvent_HappyPath(t *testing.T) {
	ctx, st := suite.New(t, true)
	suite.CompareContentsByTestName(ctx, t, st)
}

func TestPutSimpleEvent_WithoutProdID(t *testing.T) {
	ctx, st := suite.New(t, true)
	testCalPath := suite.GetCalendars(ctx, t, st)
	calIn, uid := suite.GetCalendarObjectFromFile(t, suite.InputExt)

	objPath := path.Join(testCalPath, uid+suite.IcsExt)
	reqObj, err := st.Client.PutCalendarObject(ctx, objPath, calIn)
	require.Error(t, err)
	assert.Nil(t, reqObj)

	respObj, err := st.Client.GetCalendarObject(ctx, objPath)
	require.Error(t, err)
	assert.Nil(t, respObj)
}

func TestPutSimpleEvent_WithoutVersion(t *testing.T) {
	ctx, st := suite.New(t, true)
	testCalPath := suite.GetCalendars(ctx, t, st)
	calIn, uid := suite.GetCalendarObjectFromFile(t, suite.InputExt)

	objPath := path.Join(testCalPath, uid+suite.IcsExt)
	reqObj, err := st.Client.PutCalendarObject(ctx, objPath, calIn)
	require.Error(t, err)
	assert.Nil(t, reqObj)

	respObj, err := st.Client.GetCalendarObject(ctx, objPath)
	require.Error(t, err)
	assert.Nil(t, respObj)
}
