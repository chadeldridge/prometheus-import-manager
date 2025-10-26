package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	require := require.New(t)

	got := Version()
	require.Contains(got, appNameShort, "got does not contain the correct short app name")
	require.Contains(got, appNameLong, "got does not contain the correct long app name")
	require.Contains(got, appVersion, "got does not contain the correct app version")
	require.Equal(
		fmt.Sprintf("%s - %s %s", appNameShort, appNameLong, appVersion),
		got,
		"Version() did not return the correct data",
	)
}
