package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/chadeldridge/prometheus-import-manager/core"
)

var help = `
Usage:
	pim [options] <command> [command options]
Options:
	--help				Print this help message.
	--version			Print the version.
	-c, --config-file <path>	Path to the configuration file.
	--export-first			Export targets before running any other commands.
	-e, --export-types <type>	Comma separated list of file export types (e.g. file_sd).
	-s, --sources <path>		Path to file or directory to read in the traget groups from.
	-t, --targets <path>		Path to the targets output directory.
	--targets-ext		Targets output file extension (.yml, .ymal, .json, etc.).
	--targets-suffix <string>	Change the suffix of the exported targets files.
					Default "_targets". To remove use "".
	-v, --verbose			Enable verbose logging.

Commands:
	export
		pim [options] export [<sources> [targets_dir]]

		Options:
		sources			File or directory to read in the target groups from.
		targets_dir		Directory to write the targets files to.
	
	run
		pim [options] run [<url>]

		If set --targets, --targets-suffix, and --targets-ext will be used to serve the
		/targets endpoint.

		Options:
		url			Host URL to bind to (e.g. http://0.0.0.0:9900)
		sources			File or directory to read in the target groups from. If
					sources is specified, export will be called before starting
					the http server.
		targets_dir		Directory to write the targets files to.
		--tls-certfile		Location of the cert file to use for tls.
		--tls-keyfile		Location of the key file to use for tls.
		--server-timeout	Server shutdown timeout in seconds
`

//	-e, --env <env>			Environment to run the server in.

// func printHelp(logger *core.Logger) {
func printHelp() string {
	return fmt.Sprintf("%s\n%s", Version(), help)
	// h := fmt.Sprintf("%s\n\n%s", core.Version(), help)
	// logger.PrintOut(h)
}

func getNextValue(args []string, index int) (int, string, error) {
	f := args[index]

	// Check if the value was set with --flag=value and return the value. Do not increment
	// the index if so.
	pair := strings.SplitN(f, "=", 2)
	if len(pair) == 2 {
		return index, pair[1], nil
	}

	// Check that there is another item in args.
	next := index + 1
	if len(args)-1 < next {
		return 0, "", fmt.Errorf("%w: missing value for %s", os.ErrInvalid, f)
	}

	v := args[next]
	if strings.HasPrefix(v, "-") {
		return 0, "", fmt.Errorf("%w: value appears to be a flag %s: %s", os.ErrInvalid, f, v)
	}

	return next, v, nil
}

// Skip the app name and return all flags and arguments.
// func parseArgs(logger *core.Logger, args []string) (map[string]string, []string, error) {
func parseArgs(args []string) (core.Flags, error) {
	flags, args, err := parseOptions(args)
	if err != nil {
		return flags, err
	}

	if _, ok := flags["exit_0"]; ok {
		return flags, nil
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("%w: no command found", os.ErrInvalid)
	}

	switch args[0] {
	case "export":
		args, err = parseExportCommand(flags, args)
	case "run":
		args, err = parseRunCommand(flags, args)
	}

	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		return nil, fmt.Errorf("%s: %w: %s", flags["command"], os.ErrInvalid, args[0])
	}

	return flags, nil
}

func setSources(flags core.Flags, value string) error {
	info, err := os.Stat(value)
	if err != nil {
		return err
	}

	if info.IsDir() {
		flags["sources"] = value
		return nil
	}

	flags["sources"] = value
	return nil
}

func parseOptions(args []string) (core.Flags, []string, error) {
	flags := core.Flags{}
	a := []string{}

	// Start at 1 to skip the app name.
	for i := 1; i < len(args); i++ {
		var v string
		var err error
		f := args[i]

		switch f {
		case "--help":
			// printHelp(logger)
			return map[string]string{"exit_0": "1", "print": printHelp()}, nil, nil
		case "--version":
			// logger.PrintOut(core.Version())
			return map[string]string{"exit_0": "1", "print": Version()}, nil, nil
		case "-c", "--config-file":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, nil, fmt.Errorf("%w for %s", err, f)
			}

			flags["config_file"] = v
		case "--export-first":
			flags["export_first"] = "true"
		case "-e", "--export-types":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, nil, fmt.Errorf("%w for %s", err, f)
			}

			flags["export_types"] = v
		case "-q", "--quiet":
			flags["quiet"] = "true"
		case "-s", "--sources":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, nil, fmt.Errorf("%w for %s", err, f)
			}

			flags["sources"] = v
		case "-t", "--targets":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, nil, fmt.Errorf("%w for %s", err, f)
			}

			flags["targets_dir"] = v
		case "--targets-ext":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, nil, fmt.Errorf("%w for %s", err, f)
			}

			flags["targets_file_ext"] = v
		case "--targets-suffix":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, nil, fmt.Errorf("%w for %s", err, f)
			}

			flags["targets_file_suffix"] = v
		case "-v", "--verbose":
			flags["debug"] = "true"
		default:
			// If arg begins with "-" assume it is a flag option that didn't match.
			/*
				if strings.HasPrefix(args[i], "-") {
					return nil, nil, fmt.Errorf("failed: %w: %s", os.ErrInvalid, args[i])
				}
			*/

			// If the arg is not an unknown flag, see if it matches a command and
			// return the flags and all args from the command to the end of args.
			if _, ok := commands[f]; ok {
				a = args[i:]
				return flags, a, nil
			}

			// If it was not a command, throw an error. We only want to error if an
			// unknown arguement is found before the command. Arges for commands are
			// handled elswhere.
			return nil, nil, fmt.Errorf("failed: %w: %s", os.ErrInvalid, f)
		}
	}

	return flags, a, nil
}

// parseCommand returns returns the modified flags, command options, and any remaining args along
// with an error.
func parseExportCommand(
	flags core.Flags,
	args []string,
) ([]string, error) {
	var a []string

	if len(args) < 1 {
		return nil, fmt.Errorf("%w: no command found", os.ErrInvalid)
	}

	for i, v := range args {
		switch i {
		case 0:
			flags["command"] = v
		case 1:
			flags["sources"] = v
		case 2:
			flags["targets_dir"] = v
		default:
			a = append(a, v)
		}
	}

	return a, nil
}

// parseRunCommand parse arguements for the http server and returns any remaining args along
// with an error.
func parseRunCommand(
	flags core.Flags,
	args []string,
) ([]string, error) {
	var a []string

	if len(args) < 1 {
		return nil, fmt.Errorf("%w: no command found", os.ErrInvalid)
	}

	args, err := parseRunOptions(flags, args)
	if err != nil {
		// setSource determines if value is a file or dir and sets and sets sources_file or sources to
		// value in flags. Returns the updated flags map and an error.

		return nil, err
	}

	for i, v := range args {
		switch i {
		case 0:
			flags["command"] = v
		case 1:
			err := setHttpHostFlags(flags, v)
			if err != nil {
				return nil, err
			}
		case 2:
			flags["sources"] = v
			flags["http_export_first"] = "true"
		case 3:
			flags["targets_dir"] = v
		default:
			a = append(a, v)
		}
	}

	return a, nil
}

func parseRunOptions(flags core.Flags, args []string) ([]string, error) {
	var a []string

	for i := 0; i < len(args); i++ {
		var v string
		var err error
		f := args[i]
		if !strings.HasPrefix(f, "-") {
			a = append(a, f)
			continue
		}

		switch f {
		case "--tls-certfile":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, fmt.Errorf("run: %w", err)
			}

			flags["http_tls_certfile"] = v
		case "--tls-keyfile":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, fmt.Errorf("run: %w", err)
			}

			flags["http_tls_certfile"] = v
		case "--server-timeout":
			i, v, err = getNextValue(args, i)
			if err != nil {
				return nil, fmt.Errorf("run: %w", err)
			}

			flags["http_tls_certfile"] = v
		default:
			return nil, fmt.Errorf("run: %w %s", os.ErrInvalid, f)
		}
	}

	return a, nil
}

func setHttpHostFlags(flags core.Flags, value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("http url '%s': %w", value, err)
	}

	if u.Scheme == "" {
		return fmt.Errorf("http url '%s': %w: no scheme found; must be http or https", value, os.ErrInvalid)
	}

	if u.Host == "" {
		return fmt.Errorf("http url '%s': %w: no host found", value, os.ErrInvalid)
	}

	flags["http_api_host"] = u.Hostname()
	if u.Port() != "" {
		flags["http_api_port"] = u.Port()
	}

	return nil
}
