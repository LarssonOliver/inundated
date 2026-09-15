package e2e_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSettings_GetCreatesDefaultsThenUpdatePersists(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	// GET lazily creates a default row in userless mode.
	getResp, err := client.GetSettingsWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())
	require.Equal(t, Monday, getResp.JSON200.WeekStartDay)
	require.Equal(t, "UTC", getResp.JSON200.Timezone)
	require.Equal(t, Long, getResp.JSON200.DurationFormat)
	require.Equal(t, N24h, getResp.JSON200.TimeFormat)
	require.Equal(t, Iso, getResp.JSON200.DateFormat)

	// PATCH partially updates, leaving other fields untouched.
	sunday := Sunday
	usDate := Us
	updateResp, err := client.UpdateSettingsWithResponse(ctx, UpdateSettingsJSONRequestBody{
		WeekStartDay: &sunday,
		DateFormat:   &usDate,
	})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())
	require.Equal(t, Sunday, updateResp.JSON200.WeekStartDay)
	require.Equal(t, "UTC", updateResp.JSON200.Timezone)
	require.Equal(t, Us, updateResp.JSON200.DateFormat)

	// The update is durable.
	getAfter, err := client.GetSettingsWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, 200, getAfter.StatusCode())
	require.Equal(t, Sunday, getAfter.JSON200.WeekStartDay)
}

func TestSettings_Update_InvalidTimezoneIsRejected(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	bad := "Not/AZone"
	resp, err := client.UpdateSettingsWithResponse(ctx, UpdateSettingsJSONRequestBody{
		Timezone: &bad,
	})
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode())
}

func TestSettings_Update_BrowserTimezoneSentinelIsAccepted(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	browser := "browser"
	resp, err := client.UpdateSettingsWithResponse(ctx, UpdateSettingsJSONRequestBody{
		Timezone: &browser,
	})
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
	require.Equal(t, "browser", resp.JSON200.Timezone)
}
