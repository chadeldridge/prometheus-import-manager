package targets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chadeldridge/prometheus-import-manager/core"
	"gopkg.in/yaml.v2"
)

var targetsSourceFiles = []string{"targets.yml", "targets.yaml", "targets.json"}

type TargetGroup struct {
	Jobs    []string          `json:"jobs,omitempty" yaml:"jobs,omitempty"`
	Labels  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Targets []string          `json:"targets,omitempty" yaml:"targets,omitempty"`
}

type (
	TargetGroups []*TargetGroup
	TargetMap    map[string]TargetGroups
)

// NewTargetGroup creates a new TargetGroup with initialized fields.
func NewTargetGroup() *TargetGroup {
	return &TargetGroup{
		Labels:  make(map[string]string),
		Targets: make([]string, 0),
		Jobs:    make([]string, 0),
	}
}

func findFiles(config *core.Config) ([]string, error) {
	info, err := os.Stat(config.Sources)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading source file %s: %w", config.Sources, err)
	}

	if err == nil {
		if !info.IsDir() {
			// Found a valid targets file, break out of the loop
			files := []string{config.Sources}
			return files, nil
		}
	}

	// See if an exact match exists. (e.g. targets.yml, targets.json)
	for _, p := range targetsSourceFiles {
		f := filepath.Join(config.Sources, p)
		_, err := os.Stat(f)
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return nil, err
		}

		// Found a valid targets file, break out of the loop
		files := []string{f}
		return files, nil
	}

	// If no direct match exists look for _targets pattern match. Assume
	// names are ${descriptor}_${targetsSourceFile} and return a list of all
	// matches.
	// Example: blackbox_targets.yml
	for _, p := range targetsSourceFiles {
		p = "_" + p
		files, err := filepath.Glob(filepath.Join(config.Sources, p))
		if err != nil {
			return nil, err
		}

		if len(files) > 0 {
			return files, nil
		}
	}

	return nil, os.ErrNotExist
}

func readSources(files []string) (TargetGroups, error) {
	tgs := make(TargetGroups, 0)

	for _, f := range files {
		t := make(TargetGroups, 0)

		// Get the contents of fileYAML
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		// Determine if the file is JSON or YAML
		if strings.HasSuffix(f, core.DefaultJSONFileExt) {
			// If JSON, unmarshal file as JSON
			if err := json.Unmarshal(data, &t); err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(f, core.DefaultYAMLFileExt) || strings.HasSuffix(f, "yaml") {
			// If YAML, unmarshal file as YAML
			if err := yaml.Unmarshal(data, &t); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("%w: unknown file extension for file: %s", os.ErrInvalid, f)
		}

		tgs = append(tgs, t...)
	}

	if len(tgs) == 0 {
		return nil, fmt.Errorf("no content found in targets source file")
	}

	return tgs, nil
}

// NewTargetGroups loads target groups from a file in the sources directory.
func NewTargetGroups(config *core.Config) (TargetGroups, error) {
	// Look for a valid targets source file in the sources directory.
	files, err := findFiles(config)
	if err != nil {
		return nil, fmt.Errorf("error finding sources file: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf(
			"%w in source dir: %s; no target source files found",
			os.ErrNotExist,
			config.Sources,
		)
	}

	return readSources(files)
}

// ExportTargets arranges jobs, targets, and labels into target files based on config settings.
func (t TargetGroups) ExportTargets(config *core.Config) error {
	files := t.splitByJob(config)

	err := writeTargets(config, files)
	if err != nil {
		return err
	}

	return nil
}

func (t TargetGroups) splitByJob(config *core.Config) TargetMap {
	files := make(TargetMap)
	for _, tg := range t {
		for _, job := range tg.Jobs {
			filename := job + core.DefaultTargetsFileSuffix + config.TargetsFileExt
			if files[filename] == nil {
				files[filename] = make(TargetGroups, 0)
			}

			files[filename] = append(files[filename], &TargetGroup{
				Jobs:    []string{job},
				Labels:  tg.Labels,
				Targets: tg.Targets,
			})
		}
	}

	return files
}

// writeTargets writes the target groups to files based on the config settings.
func writeTargets(config *core.Config, files TargetMap) error {
	for filename, tgs := range files {
		// Write the file
		f := filepath.Join(config.TargetsDir, filename)

		// Ensure the directory exists
		dir := filepath.Dir(f)
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			return fmt.Errorf("targets dir %w: %s", os.ErrNotExist, dir)
		}

		// We were creating the dir path if it diesn't exist but this can be unwanted or
		// dangerous and probably shouldn't be the defautl behavior. If target path is
		// inside a mounted dir and the dir is unmounted we would create a new dir instead
		// of erroring. This would then get overwritten when the dir mounts.
		/*
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		*/

		// Write the file
		if config.TargetsFileExt == core.DefaultJSONFileExt {
			if err := core.WriteJSON(f, &tgs, 0o644); err != nil {
				return err
			}

			continue
		}

		if err := core.WriteYAML(f, &tgs, 0o644); err != nil {
			return err
		}
	}

	return nil
}
