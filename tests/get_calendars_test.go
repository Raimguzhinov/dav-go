package tests

import (
	"testing"

	"github.com/Raimguhinov/dav-go/tests/suite"
	"github.com/emersion/go-ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCalendars_HappyPath(t *testing.T) {
	ctx, st := suite.New(t, true)
	principal, err := st.Client.FindCurrentUserPrincipal(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, principal)

	calendarHomeSet, err := st.Client.FindCalendarHomeSet(ctx, principal)
	require.NoError(t, err)
	assert.NotEmpty(t, calendarHomeSet)

	calendars, err := st.Client.FindCalendars(ctx, calendarHomeSet)
	require.NoError(t, err)

	for _, calendar := range calendars {
		assert.NotEmpty(t, calendar.Name)
		assert.Contains(t, calendar.SupportedComponentSet, ical.CompEvent)
		assert.Contains(t, calendar.Path, calendarHomeSet)
		assert.Greater(t, calendar.MaxResourceSize, int64(10))
	}
}

func TestGetCalendars_IncorrectPrincipal(t *testing.T) {
	ctx, st := suite.New(t, false)
	principal, err := st.Client.FindCurrentUserPrincipal(ctx)
	require.Error(t, err)
	assert.Empty(t, principal)

	calendarHomeSet, err := st.Client.FindCalendarHomeSet(ctx, principal)
	require.Error(t, err)
	require.Empty(t, calendarHomeSet)
}

func TestGetCalendars_IncorrectHomeSetPath(t *testing.T) {
	ctx, st := suite.New(t, true)
	principal, err := st.Client.FindCurrentUserPrincipal(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, principal)

	calendarHomeSet, err := st.Client.FindCalendarHomeSet(ctx, principal)
	require.NoError(t, err)
	assert.NotEmpty(t, calendarHomeSet)

	calendarHomeSet += "wrong_path"
	calendars, err := st.Client.FindCalendars(ctx, calendarHomeSet)
	require.Error(t, err)
	require.Empty(t, calendars)
}
