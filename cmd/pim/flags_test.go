package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chadeldridge/prometheus-import-manager/core"
	"github.com/stretchr/testify/require"
)

type testFlags struct {
	appname   string
	flags     core.Flags
	wantFlags core.Flags
	args      []string
}

var flagTests = []testFlags{
	{
		appname: "app",
		flags: map[string]string{
			"-v": "",
			"-c": "/tmp/pim.yml",
			"-e": "file_sd",
			"-s": "/tmp/sources",
		},
		wantFlags: map[string]string{
			"debug":        "true",
			"config_file":  "/tmp/pim.yml",
			"export_types": "file_sd",
			"sources":      "/tmp/sources",
			"command":      "export",
		},
		args: []string{"export"},
	},
	{
		appname: "app",
		flags: core.Flags{
			"--verbose":        "",
			"--config-file":    "/tmp/pim.yml",
			"--export-types":   "file_sd",
			"--targets-ext":    ".yaml",
			"--sources":        "/tmp/sources",
			"--targets":        "/tmp/targets",
			"--targets-suffix": "_sd_targets",
		},
		wantFlags: core.Flags{
			"debug":               "true",
			"config_file":         "/tmp/pim.yml",
			"export_types":        "file_sd",
			"targets_file_ext":    ".yaml",
			"sources":             "/tmp/sources",
			"targets_dir":         "/tmp/targets",
			"targets_file_suffix": "_sd_targets",
			"command":             "export",
		},
		args: []string{"export"},
	},
	{
		appname: "app",
		flags: core.Flags{
			"--verbose":        "",
			"--config-file":    "/tmp/pim.yml",
			"--export-types":   "file_sd",
			"--targets-ext":    ".yaml",
			"--sources":        "/tmp/source.file",
			"--targets":        "/tmp/targets",
			"--targets-suffix": "_sd_targets",
		},
		wantFlags: core.Flags{
			"debug":               "true",
			"config_file":         "/tmp/pim.yml",
			"export_types":        "file_sd",
			"targets_file_ext":    ".yaml",
			"sources":             "/tmp/source.file",
			"targets_dir":         "/tmp/targets",
			"targets_file_suffix": "_sd_targets",
			"command":             "export",
		},
		args: []string{"export"},
	},
}

func (f testFlags) toArgs() []string {
	a := []string{f.appname}
	for k, v := range f.flags {
		if v == "" {
			a = append(a, k)
			continue
		}

		a = append(a, k, v)
	}

	return append(a, f.args...)
}

func TestFlagsSetSources(t *testing.T) {
	require := require.New(t)

	t.Run("MissingFile", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		flags := make(core.Flags)
		value := filepath.Join(tempDir, "targets.yml")
		err := setSources(flags, value)
		require.Error(err, "setSources did not return an error")
		require.ErrorIs(err, os.ErrNotExist)
		require.Empty(flags)
	})

	t.Run("SourcesDir", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		want := core.Flags{"sources": tempDir}
		flags := make(core.Flags)
		err := setSources(flags, tempDir)
		require.NoError(err, "setSources did not return an error")
		require.Equal(want, flags, "flags returned with unexpected value")
	})

	t.Run("SourcesFile", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)
		value := filepath.Join(tempDir, "targets.yml")
		err := core.WriteFile(value, []byte("test"), 0o644)
		require.NoError(err, "WriteFile returned unexpected error")

		want := core.Flags{"sources": value}
		flags := make(core.Flags)
		err = setSources(flags, value)
		require.NoError(err, "setSources did not return an error")
		require.Equal(want, flags, "flags returned with unexpected value")
	})
}

func TestFlagsParseExportCommand(t *testing.T) {
	require := require.New(t)

	t.Run("NoArgs", func(t *testing.T) {
		flags := make(core.Flags)
		r, err := parseExportCommand(flags, []string{})

		require.Error(err, "setSources did not return an error")
		require.ErrorIs(err, os.ErrInvalid)
		require.Empty(flags)
		require.Nil(r)
	})

	t.Run("WithSources", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		want := core.Flags{"sources": tempDir, "command": "export"}
		flags := make(core.Flags)
		args := []string{"export", tempDir}
		r, err := parseExportCommand(flags, args)
		require.NoError(err, "parseExportCommand return an unexpected error")
		require.Equal(want, flags, "flags did not match")
		require.Empty(r, "remainder not empty")
	})

	t.Run("WithTargetsDir", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		want := core.Flags{"sources": tempDir, "command": "export", "targets_dir": tempDir}
		flags := make(core.Flags)
		args := []string{"export", tempDir, tempDir}
		r, err := parseExportCommand(flags, args)
		require.NoError(err, "parseExportCommand return an unexpected error")
		require.Equal(want, flags, "flags did not match")
		require.Empty(r, "remainder not empty")
	})
}

func TestFlagsParseArgs(t *testing.T) {
	require := require.New(t)

	t.Run("UnknownFlag", func(t *testing.T) {
		flags, err := parseArgs([]string{"app", "--invalid"})
		require.Error(err, "did not return an error")
		require.Nil(flags, "flags did not match")
	})

	core.MockWriteFile("/tmp/sources", core.MockTestConfigYAML, true, nil)
	t.Run("Flags", func(t *testing.T) {
		for _, ft := range flagTests {
			// log.Printf("args: %v\n", tf.toArgs())
			flags, err := parseArgs(ft.toArgs())
			// log.Printf("flags: %v\n", flags)
			require.NoError(err, "returned unexpected error")
			require.Equal(ft.wantFlags, flags, "flags did not match")
		}
	})

	t.Run("Args", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		want := core.Flags{"sources": tempDir, "command": "export"}
		flags, err := parseArgs([]string{"app", "export", tempDir})
		require.NoError(err, "returned unexpected error")
		require.NotEmpty(flags, "flags is empty")
		require.Equal(want, flags, "flags did not match")
	})

	t.Run("Help", func(t *testing.T) {
		flags, err := parseArgs([]string{"app", "--help"})
		require.NoError(err, "returned unexpected error")
		require.Contains(flags, "exit_0", "flags missing key")
		require.Contains(flags, "print", "flags missing key")
		require.Equal(printHelp(), flags["print"], "print did not match")
		// require.Nil(flags, "flags did not match")
	})

	t.Run("Version", func(t *testing.T) {
		flags, err := parseArgs([]string{"app", "--version"})
		require.NoError(err, "returned unexpected error")
		require.Contains(flags, "exit_0", "flags missing key")
		require.Contains(flags, "print", "flags missing key")
		require.Contains(flags["print"], "Manager v", "print did not match")
		// require.Nil(flags, "flags did not match")
	})
}
