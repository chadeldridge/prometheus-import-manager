package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	DefaultExportFirst       = false
	DefaultExportType        = "file_sd"
	DefaultConfigFile        = "/etc/pim/pim.yml"
	DefaultSources           = "/etc/pim/sources"
	DefaultTargetsDir        = "/etc/prometheus/file_sd"
	DefaultTargetsFileSuffix = "_targets"
	DefaultJSONFileExt       = ".json"
	DefaultYAMLFileExt       = ".yml"
	DefaultTargetsFileExt    = DefaultJSONFileExt

	DefaultAPIHost         = "0.0.0.0"
	DefaultAPIPort         = "9900"
	DefaultShutdownTimeout = 5
)

var (
	// command                string
	envPrefix              = "PIM_"
	validExportTypes       = []string{"file_sd"}
	validConfigExtensions  = []string{".yml", ".yaml", ".json"}
	validTargetsExtensions = []string{".yml", ".yaml", ".json"}
)

type Flags map[string]string

type Config struct {
	Debug bool
	Flags Flags

	ExportFirst    bool     `json:"export_first,omitempty" yaml:"export_first,omitempty"`
	RawExportTypes []string `json:"export_types,omitempty" yaml:"export_types,omitempty"`
	ExportTypes    map[string]bool

	// The path to the config file.
	ConfigFile string
	// The path to the pim sources file or directory.
	Sources string `json:"sources,omitempty" yaml:"sources,omitempty"`
	// The path to the directory to write the targets files.
	TargetsDir string `json:"targets_dir,omitempty" yaml:"targets_dir,omitempty"`

	// ".json" would create $job_targets.json
	TargetsFileExt string `json:"targets_file_ext,omitempty" yaml:"targets_file_ext,omitempty"`
	// "_targets" would create $job_targets.yml
	TargetsFileSuffix string `json:"targets_file_suffix,omitempty" yaml:"targets_file_suffix,omitempty"`

	// How to to name the target files.
	/*
		target_split:
		  - job
		  - datacenter
		  - application
		Output files: ${job}_${datacenter}_${application}_targets.json
		Might create the following target files.
		  blackbox_icmp_atl_webapp_targets.json
		  blackbox_http_atl_webapp_targets.json
		  blackbox_icmp_jfk_mysql_targets.json
	*/
	//TargetSplit []string `json:"target_split,omitempty" yaml:"target_split,omitempty"`

	// HTTP Endpont
	APIHost     string `json:"http_api_host,omitempty" yaml:"http_api_host,omitempty"`
	APIPort     string `json:"http_api_port,omitempty" yaml:"http_api_port,omitempty"`
	TLSCertFile string `json:"http_tls_cert_file,omitempty" yaml:"http_tls_cert_file,omitempty"`
	TLSKeyFile  string `json:"http_tls_key_file,omitempty" yaml:"http_tls_key_file,omitempty"`
	// Server shutdown timeout in seconds.
	ShutdownTimeout int `default:"5" json:"http_shutdown_timeout,omitempty" yaml:"http_shutdown_timeout,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Flags:             make(map[string]string),
		RawExportTypes:    make([]string, 0),
		ExportTypes:       make(map[string]bool),
		ExportFirst:       DefaultExportFirst,
		ConfigFile:        DefaultConfigFile,
		Sources:           DefaultSources,
		TargetsDir:        DefaultTargetsDir,
		TargetsFileSuffix: DefaultTargetsFileSuffix,
		TargetsFileExt:    DefaultTargetsFileExt,
		APIHost:           DefaultAPIHost,
		APIPort:           DefaultAPIPort,
		ShutdownTimeout:   DefaultShutdownTimeout,
		// TargetSplit:       make([]string, 0),
	}
}

// SetEnvPrefix takes name and sets the envPrefix variable to "NAME_".
func SetEnvPrefix(name string) error {
	if name == "" {
		return fmt.Errorf("failed to set env variable prefix: %w: '%s'", os.ErrInvalid, name)
	}

	envPrefix = fmt.Sprintf("%s_", strings.ToUpper(name))
	return nil
}

func NewConfig(
	logger *Logger,
	flags Flags,
	env map[string]string,
) (*Config, error) {
	c := DefaultConfig()

	// Parse the environment variables into a normalized format and merge with flags so we are
	// working from a single list of set variables.
	parseEnvVars(flags, env)

	// If there's a config file, parse it and set the values in the config.
	if v, ok := flags["config_file"]; ok {
		c.ConfigFile = v
	}

	if v, ok := flags["debug"]; ok {
		logger.DebugMode = true
		logger.Debug("Debug: on")
		c.setConfigValue("debug", v)
	}

	// Read in the config file. If no config is found, continue with the default values.
	logger.Debugf("parsing config file: %s", c.ConfigFile)
	if err := c.parseConfigFile(c.ConfigFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return c, err
		}

		// We will continue with the default values if no config file is found, so only
		// log an error if Debug is enabled.
		logger.Debugf("cound not find config file at %s", c.ConfigFile)
	}

	// Try to find each supported variable passed in by flags or env. If found, overwrite the
	// value in config.
	logger.Debug("updating settings from flags and environment variables")
	for k, v := range flags {
		logger.Debugf("setting config value: %s=%s", k, v)
		err := c.setConfigValue(k, v)
		if err != nil {
			return c, err
		}
	}

	// Process the RawExportTypes into a map that is easier to use later.
	c.processExportTypes()
	c.Flags = flags

	return c, nil
}

func validateTargetsFileExt(v string) error {
	// If the provided extension matches a valid targets extension, return.
	for _, e := range validTargetsExtensions {
		if v == e {
			return nil
		}
	}

	return fmt.Errorf(
		"%w targets_file_ext: %s, must be one of: %s",
		os.ErrInvalid,
		v,
		strings.Join(validTargetsExtensions, ", "),
	)
}

// Parse all supported environment variables into a map.
func parseEnvVars(flags, env Flags) {
	// Copy the environment variables into the map with normalized keys.
	for k, v := range env {
		if !strings.HasPrefix(k, envPrefix) {
			continue
		}

		k := strings.ToLower(strings.TrimPrefix(k, envPrefix))
		// If key does not exist, add it. Do not overwrite commandline args with env vars.
		if _, ok := flags[k]; !ok {
			flags[k] = v
		}
	}
}

func isValidExportTypes(v string) bool {
	return slices.Contains(validExportTypes, v)
}

func (c *Config) splitExportTypes(v string) error {
	for _, et := range strings.Split(v, ",") {
		if !isValidExportTypes(et) {
			return fmt.Errorf("invalid export_types: %s; must be one of: %s", v, strings.Join(validExportTypes, ", "))
		}

		c.RawExportTypes = append(c.RawExportTypes, et)
	}

	return nil
}

func (c *Config) setConfigValue(k, v string) error {
	switch k {
	// Do nothing with config_file since we've already handled it.
	case "config_file":
		break
	case "debug":
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("%s: %w: %w", k, os.ErrInvalid, err)
		}
		c.Debug = b
	case "export_first":
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("%s: %w: %w", k, os.ErrInvalid, err)
		}
		c.ExportFirst = b
	case "export_types":
		return c.splitExportTypes(v)
	case "targets_dir":
		c.TargetsDir = v
	case "targets_file_ext":
		if err := validateTargetsFileExt(v); err != nil {
			return err
		}

		c.TargetsFileExt = v
	case "targets_file_suffix":
		c.TargetsFileSuffix = v
	case "sources":
		c.Sources = v
	case "command":
		break
	case "http_api_host":
		c.APIHost = v
	case "http_api_port":
		c.APIPort = v
	case "http_tls_cert_file":
		c.TLSCertFile = v
	case "http_tls_key_file":
		c.TLSKeyFile = v
	case "http_shutdown_timeout":
		timeout, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("config: %w: %s was not an int '%s'", os.ErrInvalid, k, v)
		}
		c.ShutdownTimeout = timeout
	default:
		return fmt.Errorf("config: %w: %s", os.ErrInvalid, k)
	}

	return nil
}

// parseConfigFile reads the config file and unmarshals the data into the config struct.
func (c *Config) parseConfigFile(file string) error {
	if file == "" {
		return fmt.Errorf("config: %w", os.ErrInvalid)
	}

	err := tester(file)
	if err != nil {
		return err
	}

	// Get the file extension to see which unmarshal method to use.
	ext := filepath.Ext(file)
	switch ext {
	case ".yml", ".yaml":
		return ReadYAML(file, c)
	case ".json":
		return ReadJSON(file, c)
	default:
		return fmt.Errorf(
			"invalid config_targets_file_ext: %s; must be one of: %s",
			ext,
			strings.Join(validConfigExtensions, ", "),
		)
	}
}

func (c *Config) processExportTypes() {
	for _, v := range c.RawExportTypes {
		c.ExportTypes[strings.ToLower(v)] = true
	}

	if len(c.ExportTypes) == 0 {
		// Default to file_sd if nothing specified.
		c.ExportTypes[DefaultExportType] = true
	}

	// We no longer need RawExportTypes. This will most likely only matter for longer running
	// processes, like in server mode, where the GC has time to run.
	c.RawExportTypes = nil
}
