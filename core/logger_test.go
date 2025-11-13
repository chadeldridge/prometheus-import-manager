package core

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoggerNewLogger(t *testing.T) {
	require := require.New(t)
	var buf bytes.Buffer

	l := NewLogger(&buf, "app: ", log.LstdFlags, false)
	require.Equal(&buf, l.Writer(), "writer did not match")
	require.Equal("app: ", l.Prefix(), "prefix did not match")
	require.Equal(log.LstdFlags, l.Flags(), "flags did not match")
	require.False(l.DebugMode, "debug mode did not match")
}

func TestLoggerDebug(t *testing.T) {
	require := require.New(t)

	t.Run("debug on", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, "app: ", log.LstdFlags, true)
		l.Debug("test")
		require.Contains(buf.String(), "[DEBUG] test\n", "debug message did not match")
	})

	t.Run("debug off", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, "app: ", log.LstdFlags, false)
		l.Debug("test")
		require.Empty(buf, "buffer was not empty")
	})
}

func TestLoggerDebugf(t *testing.T) {
	require := require.New(t)

	t.Run("debug on", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, "app: ", log.LstdFlags, true)
		l.Debugf("test %s", "message")
		require.Contains(buf.String(), "[DEBUG] test message\n", "debug message did not match")
	})

	t.Run("debug off", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, "app: ", log.LstdFlags, false)
		l.Debugf("test %s", "message")
		require.Empty(buf, "buffer was not empty")
	})
}

// TODO: Write additional tests.
