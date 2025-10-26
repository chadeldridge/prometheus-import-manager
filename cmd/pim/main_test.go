package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chadeldridge/prometheus-import-manager/core"
	"github.com/stretchr/testify/require"
)

var (
	/*
	   	jsonSource = `[
	     {
	       "jobs": ["blackbox_icmp"],
	       "labels": {
	         "environment": "prod",
	         "datacenter": "us-east-1",
	         "role": "monitoring"
	       },
	       "targets": ["prom.example.com", "grafana.example.com"]
	     },
	     {
	       "jobs": ["node-exporter", "mysql-exporter"],
	       "labels": {
	         "environment": "stg",
	         "datacenter": "us-east-1"
	       },
	       "targets": ["node1.example.com", "node2.example.com"]
	     }
	   ]`
	*/
	yamlSource = `- jobs:
    - blackbox_icmp
  labels:
    environment: prod
    datacenter: us-east-1
    role: monitoring
  targets:
    - prom.example.com
    - grafana.example.com
- jobs:
    - node-exporter
    - mysql-exporter
  labels:
    environment: stg
    datacenter: us-east-1
  targets:
    - node1.example.com
    - node2.example.com
`
	targetsFilesJSON = map[string]string{
		"blackbox_icmp_targets.json": `[
  {
    "jobs": [
      "blackbox_icmp"
    ],
    "labels": {
      "datacenter": "us-east-1",
      "environment": "prod",
      "role": "monitoring"
    },
    "targets": [
      "prom.example.com",
      "grafana.example.com"
    ]
  }
]`,
		"mysql-exporter_targets.json": `[
  {
    "jobs": [
      "mysql-exporter"
    ],
    "labels": {
      "datacenter": "us-east-1",
      "environment": "stg"
    },
    "targets": [
      "node1.example.com",
      "node2.example.com"
    ]
  }
]`,
		"node-exporter_targets.json": `[
  {
    "jobs": [
      "node-exporter"
    ],
    "labels": {
      "datacenter": "us-east-1",
      "environment": "stg"
    },
    "targets": [
      "node1.example.com",
      "node2.example.com"
    ]
  }
]`,
	}
	targetsFilesYAML = map[string]string{
		"blackbox_icmp_targets.yml": `- jobs:
    - blackbox_icmp
  labels:
    datacenter: us-east-1
    environment: prod
    role: monitoring
  targets:
    - prom.example.com
    - grafana.example.com
`,
		"mysql-exporter_targets.yml": `- jobs:
    - mysql-exporter
  labels:
    datacenter: us-east-1
    environment: stg
  targets:
    - node1.example.com
    - node2.example.com
`,
		"node-exporter_targets.yml": `- jobs:
    - node-exporter
  labels:
    datacenter: us-east-1
    environment: stg
  targets:
    - node1.example.com
    - node2.example.com
`,
	}
)

func createTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	return tempDir
}

func unsetAllTestEnvVars() {
	for _, e := range os.Environ() {
		k := strings.SplitN(e, "=", 1)[0]
		if strings.HasPrefix(k, "TEST_PIM_") {
			os.Unsetenv(k)
		}
	}
}

func newBaseConfig() *core.Config {
	return &core.Config{
		Debug:             true,
		Flags:             core.Flags{"debug": "true", "targets_file_ext": ".yml"},
		RawExportTypes:    nil,
		ExportTypes:       map[string]bool{"file_sd": true},
		ConfigFile:        core.DefaultConfigFile,
		Sources:           core.DefaultSources,
		TargetsDir:        core.DefaultTargetsDir,
		TargetsFileExt:    core.DefaultTargetsFileExt,
		TargetsFileSuffix: core.DefaultTargetsFileSuffix,
		APIHost:           core.DefaultAPIHost,
		APIPort:           core.DefaultAPIPort,
		ShutdownTimeout:   core.DefaultShutdownTimeout,
	}
}

func TestMainGetEnv(t *testing.T) {
	require := require.New(t)

	// Unset all TEST_PIM env vars left behind by other tests.
	unsetAllTestEnvVars()

	expect := []string{"TEST_PIM_config_file", "/tmp/pim.yml"}
	err := os.Setenv(expect[0], expect[1])
	require.NoError(err, "os.Setenv returned an unexpected error")

	env := getEnv()
	require.GreaterOrEqual(len(env), 1, "getEnv() returned the wrong number of variables")
	require.Contains(env, expect[0], "variable not found")
	require.Equal(expect[1], env[expect[0]], "value did not match")
}

func TestMainPrep(t *testing.T) {
	require := require.New(t)

	t.Run("InvalidFlag", func(t *testing.T) {
		var buf bytes.Buffer
		args := []string{"app", "--invalid"}
		env := map[string]string{}

		logger, config, err := prep(&buf, args, env)
		require.Nil(logger, "not nil")
		require.Nil(config, "not nil")
		require.Error(err, "prep did not return an error")
		require.ErrorIs(err, os.ErrInvalid, "prep returned the wrong error")
	})

	t.Run("NoArgs", func(t *testing.T) {
		var buf bytes.Buffer
		args := []string{"app"}
		env := map[string]string{}

		logger, config, err := prep(&buf, args, env)
		require.Nil(logger, "not nil")
		require.Nil(config, "not nil")
		require.Error(err, "prep did not return an error")
		require.ErrorIs(err, os.ErrInvalid, "prep returned the wrong error")
	})

	t.Run("exit_0", func(t *testing.T) {
		var buf bytes.Buffer
		// Version gets printed to Stdout so be sure to set core.Stdout.
		core.Stdout = &buf
		args := []string{"app", "--version", "export"}
		env := map[string]string{}

		logger, config, err := prep(&buf, args, env)
		require.NotNil(logger, "logger not nil")
		require.Nil(config, "config not nil")
		require.NoError(err, "prep returned an unexpected error")
		require.Contains(buf.String(), appVersion, "buffer did not contain version info")
	})

	// Make NewConfig error.
	t.Run("FailNewConfig", func(t *testing.T) {
		var buf bytes.Buffer
		// Version gets printed to Stdout so be sure to set core.Stdout.
		core.Stdout = &buf
		// Set --targets-ext to a non accepted value to force an error.
		args := []string{"app", "--targets-ext", "txt", "export"}
		env := map[string]string{}

		logger, config, err := prep(&buf, args, env)
		require.NotNil(logger, "logger not nil")
		require.Nil(config, "config not nil")
		require.Error(err, "prep did not return an error")
		require.ErrorIs(err, os.ErrInvalid, "prep returned the wrong error")
		require.Contains(err.Error(), "targets_file_ext", "returned error did not contain expected text")
	})

	// Valid
	// 	Set Debug and check for buf output.
	t.Run("Success", func(t *testing.T) {
		var buf bytes.Buffer
		core.Stdout = &buf
		core.Stderr = &buf

		// Get a copy of successConfig and update the value of TargetsFileExt to match
		// the expected value for this test.
		bc := newBaseConfig()
		bc.TargetsFileExt = core.DefaultYAMLFileExt
		bc.Flags["command"] = "export"

		// Set --targets-ext to a non accepted value to force an error.
		args := []string{"app", "--targets-ext", ".yml", "export"}
		env := map[string]string{"PIM_DEBUG": "true"}
		err := core.SetEnvPrefix(appNameShort)
		require.NoError(err, "SetEnvPrefix retuned an unexpected error")

		logger, config, err := prep(&buf, args, env)

		// Test Logger
		require.NotNil(logger, "logger not nil")
		logger.Println("test message")
		require.Contains(buf.String(), "test message", "buffer did not contain expected log message")
		logger.Println("stdout test message")
		require.Contains(buf.String(), "stdout test message", "stdout buffer did not contain expected log message")
		logger.Println("stderr test message")
		require.Contains(buf.String(), "stderr test message", "stderr buffer did not contain expected log message")

		// Test Config
		require.NotNil(config, "config not nil")
		require.Contains(config.Flags, "targets_file_ext", "config.Flags missing item: targets_file_ext")
		require.Equal(config.Flags["command"], "export", "config.Args missing item: export")
		require.Equal(config.Sources, core.DefaultSources, "source dir did not match default")
		require.Equal(bc, config, "config had expected values")

		// Test Debug
		require.True(config.Debug, "debug not true in config")
		require.Contains(buf.String(), "Debug: on", "buffer did not contain expected debug message")
		require.Contains(buf.String(), "Config: ", "buffer did not contain expected debug message")

		// Test Error
		require.NoError(err, "prep returned an unexpected error")
	})
}

func TestMainExport(t *testing.T) {
	require := require.New(t)

	t.Run("SourcesFileNotFound", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		var buf bytes.Buffer
		core.Stdout = &buf
		core.Stderr = &buf
		logger := core.NewLogger(&buf, "pim: ", log.LstdFlags, false)

		flags := core.Flags{"command": "export"}
		config := newBaseConfig()
		config.Flags = flags
		config.Sources = tempDir

		err := export(logger, config)
		require.Error(err, "export did not return an error")
		require.ErrorIs(err, os.ErrNotExist, "export returned the wrong error")
		require.Contains(err.Error(), "error loading source:", "export returned the wrong error")
	})

	t.Run("TargetsDirNotFound", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		var buf bytes.Buffer
		logger := core.NewLogger(&buf, "pim: ", log.LstdFlags, false)

		sourcesDir := filepath.Join(tempDir, "sources")
		err := os.MkdirAll(sourcesDir, 0o755)
		require.NoError(err, "failed to create temp sources dir")
		/*
			info, err := os.Stat(sourcesDir)
			require.NoError(err, "failed to stat temp sources dir")
			fmt.Println(info.Mode().Perm().String())
		*/

		flags := core.Flags{"command": "export"}
		config := newBaseConfig()
		config.Flags = flags
		config.Sources = sourcesDir
		config.TargetsDir = filepath.Join(tempDir, "invalidDir")

		sfile := filepath.Join(sourcesDir, "targets.yml")
		err = core.WriteFile(sfile, []byte(yamlSource), 0o644)
		require.NoError(err, "failed to write sources file to %s", sfile)

		err = export(logger, config)
		require.Error(err, "export did not return an error")
		require.ErrorIs(err, os.ErrNotExist, "export returned the wrong error")
		require.Contains(err.Error(), "error exporting targets:", "export returned the wrong error")
	})

	t.Run("YAML", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		var buf bytes.Buffer
		logger := core.NewLogger(&buf, "pim: ", log.LstdFlags, false)

		// Create the sources directory and file to load.
		sourcesDir := filepath.Join(tempDir, "sources")
		err := os.MkdirAll(sourcesDir, 0o755)
		require.NoError(err, "failed to create temp sources dir")

		sfile := filepath.Join(sourcesDir, "targets.yml")
		err = core.WriteFile(sfile, []byte(yamlSource), 0o644)
		require.NoError(err, "failed to write sources file to %s", sfile)

		// Create the targets directory where the targets file will be written.
		targetsDir := filepath.Join(tempDir, "targets")
		err = os.MkdirAll(targetsDir, 0o755)
		require.NoError(err, "failed to create temp targets dir")

		flags := core.Flags{"command": "export"}
		config := newBaseConfig()
		config.Flags = flags
		config.Sources = sourcesDir
		config.TargetsDir = targetsDir
		// Specify the target files should be YAMl since JSON is the default.
		config.TargetsFileExt = core.DefaultYAMLFileExt

		err = export(logger, config)
		require.NoError(err, "export returned an unexpected error")

		// Read in the targets.yml file and make sure it matches the expected output.
		// e, err := os.ReadDir(targetsDir)
		// require.NoError(err)
		//	f := filepath.Join(targetsDir, i.Name())
		for fn, d := range targetsFilesYAML {
			f := filepath.Join(config.TargetsDir, fn)
			_, err := os.Stat(f)
			require.NoError(err, "stat targets file faild with an unexpected error")
			got, err := core.ReadFile(f)
			// got = got[:len(got)-1]
			require.NoError(err, "ReadFile returned unexpected error")
			require.Equal(d, string(got), "contents of targets file did not match")
		}
	})

	t.Run("JSON", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		var buf bytes.Buffer
		logger := core.NewLogger(&buf, "pim: ", log.LstdFlags, false)

		// Create the sources directory and file to load.
		sourcesDir := filepath.Join(tempDir, "sources")
		err := os.MkdirAll(sourcesDir, 0o755)
		require.NoError(err, "failed to create temp sources dir")

		sfile := filepath.Join(sourcesDir, "targets.yml")
		err = core.WriteFile(sfile, []byte(yamlSource), 0o644)
		require.NoError(err, "failed to write sources file to %s", sfile)

		// Create the targets directory where the targets file will be written.
		targetsDir := filepath.Join(tempDir, "targets")
		err = os.MkdirAll(targetsDir, 0o755)
		require.NoError(err, "failed to create temp targets dir")

		flags := core.Flags{"command": "export"}
		config := newBaseConfig()
		config.Flags = flags
		config.Sources = sourcesDir
		config.TargetsDir = targetsDir

		err = export(logger, config)
		require.NoError(err, "export returned an unexpected error")

		// Read in the targets.yml file and make sure it matches the expected output.
		// e, err := os.ReadDir(targetsDir)
		// require.NoError(err)
		// for _, i := range e {
		//	fn := i.Name()
		for fn, d := range targetsFilesJSON {
			f := filepath.Join(config.TargetsDir, fn)
			_, err := os.Stat(f)
			require.NoError(err, "stat targets file faild with an unexpected error")
			got, err := core.ReadFile(f)
			// got = got[:len(got)-1]
			// fmt.Printf(`"%s": `, fn)
			// fmt.Printf("`%s`,\n", got)
			require.NoError(err, "ReadFile returned unexpected error")
			require.Equal(d, string(got), "contents of targets file did not match")
		}
	})
}

func testRunServer(ctx context.Context, logger *core.Logger, config *core.Config) error {
	// Capture the interrupt signal to gracefully shutdown the server.
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)

	go func() {
		err := run(ctx, logger, config)
		ch <- err
	}()
	cancel()

	err := <-ch
	return err
}

func TestMainRun(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	var out bytes.Buffer
	logger := core.NewLogger(&out, "pim: ", log.LstdFlags, false)
	core.Stdout = &out
	core.Stderr = &out

	config := newBaseConfig()
	config.Debug = true

	// Unemplemented
	err := testRunServer(ctx, logger, config)
	require.NoError(err, "run returned an unexpected error")
}

func testHandlerRun(ctx context.Context, logger *core.Logger, config *core.Config) error {
	// Capture the interrupt signal to gracefully shutdown the server.
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)

	go func() {
		err := handler(ctx, logger, config)
		ch <- err
	}()
	cancel()

	err := <-ch
	return err
}

func TestMainHandler(t *testing.T) {
	require := require.New(t)

	t.Run("export", func(t *testing.T) {
		tempDir := createTempDir(t)
		defer os.RemoveAll(tempDir)

		var buf bytes.Buffer
		logger := core.NewLogger(&buf, "pim: ", log.LstdFlags, false)

		// Create the sources directory and file to load.
		sourcesDir := filepath.Join(tempDir, "sources")
		err := os.MkdirAll(sourcesDir, 0o755)
		require.NoError(err, "failed to create temp sources dir")

		sfile := filepath.Join(sourcesDir, "targets.yml")
		err = core.WriteFile(sfile, []byte(yamlSource), 0o644)
		require.NoError(err, "failed to write sources file to %s", sfile)

		// Create the targets directory where the targets file will be written.
		targetsDir := filepath.Join(tempDir, "targets")
		err = os.MkdirAll(targetsDir, 0o755)
		require.NoError(err, "failed to create temp targets dir")

		flags := core.Flags{"command": "export"}
		config := newBaseConfig()
		config.Flags = flags
		config.Sources = sourcesDir
		config.TargetsDir = targetsDir

		ctx := context.Background()
		err = handler(ctx, logger, config)
		require.NoError(err, "export returned an unexpected error")

		// Read in the targets.yml file and make sure it matches the expected output.
		// e, err := os.ReadDir(targetsDir)
		// require.NoError(err)
		// for _, i := range e {
		//	fn := i.Name()
		for fn, d := range targetsFilesJSON {
			f := filepath.Join(config.TargetsDir, fn)
			_, err := os.Stat(f)
			require.NoError(err, "stat targets file faild with an unexpected error")
			got, err := core.ReadFile(f)
			// got = got[:len(got)-1]
			// fmt.Printf(`"%s": `, fn)
			// fmt.Printf("`%s`,\n", got)
			require.NoError(err, "ReadFile returned unexpected error")
			require.Equal(d, string(got), "contents of targets file did not match")
		}
	})

	t.Run("run", func(t *testing.T) {
		var buf bytes.Buffer
		core.Stdout = &buf
		core.Stderr = &buf
		logger := core.NewLogger(&buf, "pim: ", log.LstdFlags, false)

		flags := core.Flags{"command": "run"}
		config := newBaseConfig()
		config.Flags = flags
		config.ShutdownTimeout = 5

		ctx := context.Background()
		err := testHandlerRun(ctx, logger, config)
		require.NoError(err, "handler returned an unexepcted error")
		require.Contains(
			buf.String(),
			"http server listening on 0.0.0.0:9900",
			"handler run output did not contain expected string",
		)
	})
}
