package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	mockConfig = &Config{
		Flags:             make(map[string]string),
		Debug:             true,
		ConfigFile:        "/tmp/pim.yml",
		RawExportTypes:    []string{DefaultExportType},
		ExportTypes:       map[string]bool{DefaultExportType: true},
		Sources:           "/tmp/sources",
		TargetsDir:        "/tmp/targets",
		TargetsFileExt:    ".yaml",
		TargetsFileSuffix: "_sd_targets",
		APIHost:           DefaultAPIHost,
		APIPort:           DefaultAPIPort,
		ShutdownTimeout:   5,
	}
	mockConfigValues = map[string]string{
		"debug":                 "true",
		"config_file":           "/tmp/pim.yml",
		"export_types":          "file_sd",
		"targets_file_ext":      ".yaml",
		"sources":               "/tmp/sources",
		"targets_dir":           "/tmp/targets",
		"targets_file_suffix":   "_sd_targets",
		"http_api_host":         DefaultAPIHost,
		"http_api_port":         DefaultAPIPort,
		"http_shutdown_timeout": "5",
	}
)

func newEmptyConfig() *Config {
	return &Config{
		Flags:          make(map[string]string),
		RawExportTypes: make([]string, 0),
		ExportTypes:    make(map[string]bool),
	}
}

func newDefaultConfig() *Config {
	return &Config{
		Flags:             make(map[string]string),
		RawExportTypes:    make([]string, 0),
		ExportTypes:       make(map[string]bool),
		ConfigFile:        DefaultConfigFile,
		Sources:           DefaultSources,
		TargetsDir:        DefaultTargetsDir,
		TargetsFileExt:    DefaultTargetsFileExt,
		TargetsFileSuffix: DefaultTargetsFileSuffix,
		APIHost:           DefaultAPIHost,
		APIPort:           DefaultAPIPort,
		ShutdownTimeout:   5,
	}
}

func TestDefaultConfig(t *testing.T) {
	require := require.New(t)

	got := DefaultConfig()
	require.Equal(newDefaultConfig(), got, "default config did not have expected values.")
}

func TestConfigValidateTargetsFileExt(t *testing.T) {
	require := require.New(t)

	for _, i := range validTargetsExtensions {
		t.Run("Match_"+i, func(t *testing.T) {
			err := validateTargetsFileExt(i)
			require.NoError(err, "validateTargetsFileExt() returned an unexpected error.")
		})
	}

	t.Run("NoMatch", func(t *testing.T) {
		err := validateTargetsFileExt(".txt")
		require.Error(err, "validateTargetsFileExt() did not return an error for '.txt'.")
	})

	t.Run("EmptyValue", func(t *testing.T) {
		err := validateTargetsFileExt("")
		require.Error(err, "validateTargetsFileExt() did not return an error for '.txt'.")
	})
}

func TestConfigParseEnvVars(t *testing.T) {
	require := require.New(t)
	envPrefix = "TEST_TEST_"

	t.Run("NoVars", func(t *testing.T) {
		flags := make(Flags)
		env := make(map[string]string)
		parseEnvVars(flags, env)
		require.Empty(flags, "parssEnvVars() returned a non-empty value")
	})

	t.Run("Vars", func(t *testing.T) {
		flags := make(Flags)
		env := map[string]string{"TEST_TEST_VAR": "test"}
		parseEnvVars(flags, env)
		require.NotEmpty(flags, "parssEnvVars() returned an empty value")
		require.Contains(flags, "var", "vars does not contain expected key")
		require.Equal(flags["var"], "test", "vars[test_var] returned unexpected value")
	})
}

func requireValue(t *testing.T, c *Config, k, v string) {
	require := require.New(t)

	switch k {
	// Do nothing with config_file since we've already handled it.
	case "config_file":
		require.Empty(c.ConfigFile, fmt.Sprintf("%s is not empty", k))
	case "debug":
		if v == "true" {
			require.True(c.Debug, fmt.Sprintf("%s is not true", k))
			return
		}

		require.False(c.Debug, fmt.Sprintf("%s is not false", k))
	case "export_types":
		for _, et := range strings.Split(v, ",") {
			require.Contains(c.RawExportTypes, et, fmt.Sprintf("%s not found RawExportTypes", k))
		}
	case "targets_dir":
		require.Equal(v, c.TargetsDir, fmt.Sprintf("%s did not match", k))
	case "targets_file_ext":
		require.Equal(v, c.TargetsFileExt, fmt.Sprintf("%s did not match", k))
	case "targets_file_suffix":
		require.Equal(v, c.TargetsFileSuffix, fmt.Sprintf("%s did not match", k))
	case "sources":
		require.Equal(v, c.Sources, fmt.Sprintf("%s did not match", k))
	}
}

func TestConfigSetConfigValue(t *testing.T) {
	require := require.New(t)

	t.Run("EmptyKey", func(t *testing.T) {
		config := &Config{}
		err := config.setConfigValue("", "")
		require.Error(err, "did not return an error")
		require.Empty(config, "Config struct not empty")
	})

	t.Run("EmptyValue", func(t *testing.T) {
		// Use newEmptyConfig so they map values won't be nil which is different from make() empty.
		config := newEmptyConfig()
		exp := newEmptyConfig()
		for k := range mockConfigValues {
			t.Run("EmptyValue_"+k, func(t *testing.T) {
				err := config.setConfigValue(k, "")
				switch k {
				case "debug":
					require.Error(err, "setConfigValue did not return error")
					require.ErrorIs(err, os.ErrInvalid, "setConfigValue returned wrong error")
				case "export_types", "targets_file_ext":
					require.Error(err, "setConfigValue() did not return error")
					require.Equal(exp, config, "configs did not match")
				case "http_shutdown_timeout":
					require.Error(err, "setConfigValue did not return an error")
				default:
					require.NoError(err, "setConfigValue returned an unexpected error")
				}

				require.Equal(exp, config, "configs did not match")
			})
		}
	})

	t.Run("InvalidKey", func(t *testing.T) {
		config := &Config{}
		err := config.setConfigValue("no_match", "value")
		require.Error(err, "did not return an error")
		require.Empty(config, "Config struct not empty")
	})

	for k, v := range mockConfigValues {
		// Use newEmptyConfig so they map values won't be nil which is different from make() empty.
		config := newEmptyConfig()
		// fmt.Printf("k: %s, vType: %T v: %s", k, v, v)
		t.Run("ValidValue_"+k, func(t *testing.T) {
			err := config.setConfigValue(k, v)
			require.NoError(err, "setConfigValue() returned an unexpected error")
			require.NotEmpty(config, "Config struct is empty")
			requireValue(t, config, k, v)
		})
	}
}

func requireConfigValue(t *testing.T, c *Config, k, v string, checkExportTypes bool) {
	require := require.New(t)

	switch k {
	// Skip config_file before calling this function as it might be in various states depending
	// on what you are testing. If you need to test config_file, test it directly.
	// case "config_file":
	case "debug":
		if v == "true" {
			require.True(c.Debug, fmt.Sprintf("%s is not true", k))
			return
		}

		require.False(c.Debug, fmt.Sprintf("%s is not false", k))
	case "export_types":
		for _, et := range strings.Split(v, ",") {
			require.Contains(c.RawExportTypes, et, fmt.Sprintf("%s '%s' not found in RawExportTypes", k, v))
			if checkExportTypes {
				require.Contains(c.ExportTypes, et, fmt.Sprintf("%s '%s' not found in ExportTypes", k, v))
			}
		}
	case "targets_dir":
		require.Equal(v, c.TargetsDir, fmt.Sprintf("%s did not match", k))
	case "targets_file_ext":
		require.Equal(v, c.TargetsFileExt, fmt.Sprintf("%s did not match", k))
	case "targets_file_suffix":
		require.Equal(v, c.TargetsFileSuffix, fmt.Sprintf("%s did not match", k))
	case "sources":
		require.Equal(v, c.Sources, fmt.Sprintf("%s did not match", k))
	}
}

func TestConfigParseConfigFile(t *testing.T) {
	require := require.New(t)
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	t.Run("EmptyFilename", func(t *testing.T) {
		config := &Config{}
		err := config.parseConfigFile("")
		require.Error(err, "did not return an error")
		require.Empty(config, "Config struct not empty")
	})

	t.Run("MissingFile", func(t *testing.T) {
		config := &Config{}
		name := uuid.New().String()
		file := filepath.Join(tempDir, name)

		err := config.parseConfigFile(file)
		require.Error(err, "did not return an error")
		require.Empty(config, "Config struct not empty")
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		config := &Config{}
		invalidConfig := "this is not a valid config"
		f := filepath.Join(tempDir, "invalid.yml")
		err := WriteFile(f, []byte(invalidConfig), PermStdRead)
		require.NoError(err, "failed to write invalid config file")

		err = config.parseConfigFile(f)
		require.Error(err, "did not return an error")
		require.Empty(config, "Config struct not empty")
	})

	// Use the test config from mock_file.go to test loading the config.
	t.Run("ValidYAML", func(t *testing.T) {
		config := newEmptyConfig()

		// Write a valid config file to read in.
		f := filepath.Join(tempDir, "pim.yml")
		err := WriteFile(f, []byte(MockTestConfigYAML), PermStdRead)
		require.NoError(err, "failed to write invalid config file")

		err = config.parseConfigFile(f)
		require.NoError(err, "did not return an error")
		require.NotEmpty(config, "Config struct not empty")
		for k, v := range mockConfigValues {
			// Skip values we know are not set.
			if k == "config_file" {
				continue
			}

			t.Run("ValidValue_"+k, func(t *testing.T) {
				requireConfigValue(t, config, k, v, false)
			})
		}
	})

	t.Run("ValidJSON", func(t *testing.T) {
		config := newEmptyConfig()

		// Write a valid config file to read in.
		f := filepath.Join(tempDir, "pim.json")
		err := WriteFile(f, []byte(MockTestConfigJSON), PermStdRead)
		require.NoError(err, "failed to write invalid config file")

		err = config.parseConfigFile(f)
		require.NoError(err, "returned an error")
		require.NotEmpty(config, "Config struct not empty")
		fmt.Printf("RawExportTypes: %s\n", strings.Join(config.RawExportTypes, ", "))
		for k, v := range mockConfigValues {
			// Skip values we know are not set.
			if k == "config_file" {
				continue
			}

			t.Run("ValidValue_"+k, func(t *testing.T) {
				requireConfigValue(t, config, k, v, false)
			})
		}
	})
}

func TestConfigProcessExportTypes(t *testing.T) {
	require := require.New(t)

	t.Run("Empty", func(t *testing.T) {
		config := newEmptyConfig()
		config.processExportTypes()
		require.Len(config.ExportTypes, 1, "unexpected array length")
		require.Contains(config.ExportTypes, DefaultExportType, "did not contain default value")
	})

	t.Run("SingleValue", func(t *testing.T) {
		config := newEmptyConfig()
		config.RawExportTypes = []string{DefaultExportType}
		config.processExportTypes()
		require.Len(config.ExportTypes, 1, "unexpected array length")
		require.Contains(config.ExportTypes, DefaultExportType, "did not contain default value")
	})

	t.Run("MultipleValues", func(t *testing.T) {
		values := []string{"file_sd", "http"}
		config := newEmptyConfig()
		config.RawExportTypes = values
		config.processExportTypes()
		require.Len(config.ExportTypes, len(values), "unexpected array length")
		for _, v := range values {
			require.Contains(config.ExportTypes, v, "did not contain default value")
		}
	})
}

func TestConfigNewConfig(t *testing.T) {
	require := require.New(t)
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	t.Run("MissingConfig", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, "pim: ", log.LstdFlags, true)

		f := filepath.Join(tempDir, "pim.yml")
		err := os.Remove(f)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			require.NoError(err, "unexpected error we ensuring %s does not exist", f)
		}

		flags := Flags{"debug": "true"}
		expect := newDefaultConfig()
		expect.RawExportTypes = nil
		expect.ExportTypes = map[string]bool{DefaultExportType: true}
		expect.ConfigFile = DefaultConfigFile
		expect.Debug = true
		expect.Flags = flags

		config, err := NewConfig(l, flags, map[string]string{})
		require.NoError(err, "NewConfig returned unexpected error")
		require.Equal(expect, config, "config had expected values")
		require.Contains(
			buf.String(),
			fmt.Sprintf("[DEBUG] cound not find config file at %s", DefaultConfigFile),
			"debug line not found in log",
		)
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, "pim: ", log.LstdFlags, true)

		f := filepath.Join(tempDir, "pim.yml")
		err := os.Remove(f)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			require.NoError(err, "unexpected error we ensuring %s does not exist", f)
		}

		flags := Flags{"config_file": f, "debug": "true"}
		expect := newDefaultConfig()
		expect.ConfigFile = f
		expect.Debug = true

		// Write a valid config file to read in.
		err = WriteFile(f, []byte("this is not a valid config"), PermStdRead)
		require.NoError(err, "failed to write invalid config file")

		config, err := NewConfig(l, flags, map[string]string{})
		require.Error(err, "NewConfig did not return an error")
		require.Equal(expect, config, "config had expected values")
		require.Empty(buf, "log buffer not empty")
	})

	t.Run("ValidConfig", func(t *testing.T) {
		var buf bytes.Buffer
		l := NewLogger(&buf, "pim: ", log.LstdFlags, true)

		f := filepath.Join(tempDir, "pim.yml")
		err := os.Remove(f)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			require.NoError(err, "unexpected error we ensuring %s does not exist", f)
		}

		flags := Flags{"config_file": f, "debug": "true"}
		expect := *mockConfig
		expect.RawExportTypes = nil
		expect.ConfigFile = f
		expect.Flags = flags

		// Write a valid config file to read in.
		err = WriteFile(f, []byte(MockTestConfigYAML), PermStdRead)
		require.NoError(err, "failed to write invalid config file")

		config, err := NewConfig(l, flags, map[string]string{})
		require.NoError(err, "NewConfig returned unexpected error")
		require.Equal(&expect, config, "config had expected values")
		require.Empty(buf, "log buffer not empty")
	})

	/*
		t.Run("MissingConfig", func(t *testing.T) {
			// Write a valid config file to read in.
			f := filepath.Join(tempDir, "pim.yml")
			err := WriteFile(f, []byte(MockTestConfigYAML), PermStdRead)
			require.NoError(err, "failed to write invalid config file")
		})
	*/
}
