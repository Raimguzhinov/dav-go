package tests

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/Raimguhinov/dav-go/tests/suite"
	"github.com/emersion/go-ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	EnvTestCaseDir = "TESTCASES_DIR"
	InputExt       = ".in.ics"
	OutputExt      = ".out.ics"
	IcsExt         = ".ics"
)

func getCalendars(ctx context.Context, t *testing.T, st *suite.Suite) string {
	principal, err := st.Client.FindCurrentUserPrincipal(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, principal)

	calendarHomeSet, err := st.Client.FindCalendarHomeSet(ctx, principal)
	require.NoError(t, err)
	assert.NotEmpty(t, calendarHomeSet)

	st.CreateTestCalendar(ctx, calendarHomeSet)

	calendars, err := st.Client.FindCalendars(ctx, calendarHomeSet)
	require.NoError(t, err)
	assert.NotEmpty(t, calendars)

	var testCalPath string
	for _, calendar := range calendars {
		assert.NotEmpty(t, calendar.Name)
		assert.Contains(t, calendar.SupportedComponentSet, ical.CompEvent)
		assert.Contains(t, calendar.Path, calendarHomeSet)
		assert.Greater(t, calendar.MaxResourceSize, int64(10))

		if calendar.Name == st.TestCalendarData["name"].(string) {
			testCalPath = calendar.Path
		}
	}
	return testCalPath
}

func getCalendarObjectFromFile(t *testing.T, ext string) (*ical.Calendar, string) {
	testCaseDir := os.Getenv(EnvTestCaseDir)
	require.NotEmpty(t, testCaseDir)

	var testPath string
	switch ext {
	case InputExt:
		testPath = path.Join(testCaseDir, t.Name()+InputExt)
	case OutputExt:
		testPath = path.Join(testCaseDir, t.Name()+OutputExt)
	default:
	}

	data, err := os.Open(testPath)
	require.NoError(t, err)
	defer data.Close()

	reader := bufio.NewReader(data)
	dec := ical.NewDecoder(reader)
	cal, err := dec.Decode()
	require.NoError(t, err)
	assert.NotEmpty(t, cal)

	uid, err := cal.Events()[0].Props.Text(ical.PropUID)
	require.NoError(t, err)
	assert.NotEmpty(t, uid)

	return cal, uid
}

func readBytesData(t *testing.T, buf *bytes.Buffer, cal *ical.Calendar) {
	f := bufio.NewWriter(buf)
	enc := ical.NewEncoder(f)
	err := enc.Encode(cal)
	require.NoError(t, err)
	_ = f.Flush()
}

func TestPutSimpleEvent_HappyPath(t *testing.T) {
	ctx, st := suite.New(t, true)
	testCalPath := getCalendars(ctx, t, st)
	calIn, uid := getCalendarObjectFromFile(t, InputExt)

	objPath := path.Join(testCalPath, uid+IcsExt)
	reqObj, err := st.Client.PutCalendarObject(ctx, objPath, calIn)
	require.NoError(t, err)
	assert.NotEmpty(t, reqObj)

	respObj, err := st.Client.GetCalendarObject(ctx, objPath)
	require.NoError(t, err)
	assert.NotEmpty(t, respObj)

	var respBuf, fileOutBuf bytes.Buffer
	readBytesData(t, &respBuf, respObj.Data)

	calOut, _ := getCalendarObjectFromFile(t, OutputExt)
	readBytesData(t, &fileOutBuf, calOut)

	assert.Equal(t, fileOutBuf.String(), respBuf.String())
}
