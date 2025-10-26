package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chadeldridge/prometheus-import-manager/core"
	"github.com/chadeldridge/prometheus-import-manager/router"
	"github.com/chadeldridge/prometheus-import-manager/targets"
	"github.com/chadeldridge/prometheus-import-manager/web"
)

var commands = map[string]bool{
	"export": true,
	"run":    true,
}

func main() {
	ctx := context.Background()
	logger, config, err := prep(os.Stdout, os.Args, getEnv())
	if err != nil {
		log.Printf("%v\n", err)
		if errors.Is(err, os.ErrInvalid) {
			fmt.Fprintln(core.Stderr, printHelp())
		}
		os.Exit(1)
	}

	if config == nil {
		os.Exit(0)
	}

	err = handler(ctx, logger, config)
	if err != nil {
		logger.Printf("%v\n", err)
		if errors.Is(err, os.ErrInvalid) {
			logger.PrintErr(printHelp())
		}
		os.Exit(1)
	}
}

func getEnv() map[string]string {
	env := map[string]string{}
	for _, e := range os.Environ() {
		v := strings.SplitN(e, "=", 2)
		env[v[0]] = v[1]
	}

	return env
}

func prep(out io.Writer, args []string, env map[string]string) (*core.Logger, *core.Config, error) {
	// Get CLI flags and arguements. If we got nothing from the CLI, print help and exit.
	flags, err := parseArgs(args)
	_, exit0 := flags["exit_0"]
	if err != nil && !exit0 {
		return nil, nil, err
	}
	/*
		flags, args, err := parseFlags(logger, args)
		if err != nil {
			return logger, nil, err
		}
	*/

	// Setup a logger.
	logger := core.NewLogger(out, "pim: ", log.LstdFlags, false)
	_, ok := flags["command"]
	if len(flags) == 0 && !ok {
		return logger, nil, fmt.Errorf("pim: %w", os.ErrInvalid)
	}

	// If "exit_0" exists in flags, check for a "print" flag and print it. This is to prevent
	// having to pass logger to the flags functions for just one or two things and to try to
	// handle output as high up as possible.
	if exit0 {
		if _, ok := flags["print"]; ok {
			logger.Println(flags["print"])
		}
		return logger, nil, nil
	}

	// Setup the configuration.
	config, err := core.NewConfig(logger, flags, env)
	if err != nil {
		return logger, nil, err
	}

	// Update logger with config value.
	logger.DebugMode = config.Debug
	logger.Debug("Debug: on")

	// Print config if in debug mode.
	logger.Debugf("Config: %+v\n", config)

	return logger, config, nil
}

func handler(ctx context.Context, logger *core.Logger, config *core.Config) error {
	command, ok := config.Flags["command"]
	if !ok {
		return fmt.Errorf("handler: %w: missing command", os.ErrInvalid)
	}

	if config.ExportFirst && command != "export" {
		logger.Debugf("handler: export_first: %s\n", config.ExportFirst)
		logger.Debugf("handler: exporting targets before %s\n", command)
		err := export(logger, config)
		if err != nil {
			return fmt.Errorf("handler: error running export first: %s", err)
		}
	}

	switch command {
	case "export":
		logger.Debug("running exporter")
		return export(logger, config)
	case "run":
		logger.Debug("running http server")
		return run(ctx, logger, config)
	}

	logger.Debugf("handler: invalid command %s", command)
	return fmt.Errorf("handler: %w", os.ErrInvalid)
}

// export targets from sources to file_sd.
func export(logger *core.Logger, config *core.Config) error {
	logger.Debugf("export: importing targets from %s\n", config.Sources)
	tgs, err := targets.NewTargetGroups(config)
	if err != nil {
		return fmt.Errorf("export: error loading source: %w", err)
	}

	logger.Debugf(
		"export: exporting targets to %s as %s%s\n",
		config.TargetsDir,
		config.TargetsFileSuffix,
		config.TargetsFileExt,
	)
	err = tgs.ExportTargets(config)
	if err != nil {
		return fmt.Errorf("export: error exporting targets: %w", err)
	}

	logger.Debug("export: targets exported successfully")
	return nil
}

// run allows us to setup and implement in testing and production.
func run(ctx context.Context, logger *core.Logger, config *core.Config) error {
	logger.PrintOut(config)
	// Setup the HTTP server.
	srv := router.NewHTTPServer(logger, config)
	// Add routes and do anything else we need to do before starting the server.

	// Add web routes.
	err := web.AddRoutes(&srv)
	if err != nil {
		return err
	}

	/*
		// Add API routes.
		err = api.AddRoutes(&srv)
		if err != nil {
			return err
		}
	*/

	// Start API Server.
	logger.Debug("run: starting http server")
	return srv.Start(ctx, config.ShutdownTimeout)
}
