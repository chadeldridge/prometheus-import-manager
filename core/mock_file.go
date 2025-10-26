package core

import (
	"errors"
	"fmt"
	"os"
)

var (
	MockFS          = make(map[string]MockFile)
	MockTesterPerms = false
)

type MockFile struct {
	Name       string
	Contents   []byte
	IsReadable bool
	Error      error
}

func getFile(file string) (MockFile, error) {
	if f, ok := MockFS[file]; ok {
		return f, nil
	}
	/*
		for n, f := range MockFS {
			if file == n {
				return f, nil
			}
		}
	*/

	return MockFile{}, fmt.Errorf("mock_file: %w", os.ErrNotExist)
}

func MockClearFS() { MockFS = map[string]MockFile{} }

// MockWriteFile creates a new MockFile and appends it to the MockTesterFiles slice. If the file
// already exists, it will update the file with the new data.
func MockWriteFile(fname string, fdata []byte, isReadable bool, ferr error) {
	if fname == "" {
		return
	}

	MockFS[fname] = MockFile{
		Name:       fname,
		Contents:   fdata,
		IsReadable: isReadable,
		Error:      ferr,
	}
}

func MockTester(file string) error {
	if file == "" {
		return fmt.Errorf("mock_file: %w", os.ErrInvalid)
	}

	f, err := getFile(file)
	if err != nil {
		return err
	}

	if f.IsReadable {
		return nil
	}

	return fmt.Errorf("mock_file: %w", os.ErrPermission)
}

func MockReader(file string) ([]byte, error) {
	if file == "" {
		return nil, fmt.Errorf("mock_file: %w", os.ErrNotExist)
	}

	f, err := getFile(file)
	if err != nil {
		return nil, err
	}

	if f.Error != nil {
		return nil, f.Error
	}

	return f.Contents, nil
}

func MockWriter(file string, data []byte, perm os.FileMode) error {
	if file == "" {
		return fmt.Errorf("mock_file: %w", os.ErrNotExist)
	}

	f, err := getFile(file)
	if errors.Is(err, os.ErrNotExist) {
		f = MockFile{
			Name:       file,
			Contents:   data,
			IsReadable: true,
			Error:      nil,
		}
	}

	if f.Error != nil {
		return f.Error
	}

	MockFS[file] = f
	return nil
}

var (
	MockTestConfigYAML = []byte(`debug: true
config_file: /tmp/pim.yml
export_types:
  - file_sd
targets_file_ext: ".yaml"
sources: /tmp/sources
targets_dir: /tmp/targets
targets_file_suffix: "_sd_targets"
`)
	MockTestConfigJSON = []byte(`{
"debug":          true,
"config_file":    "/tmp/pim.yml",
"export_types": [ "file_sd" ],
"targets_file_ext":       ".yaml",
"sources":  "/tmp/sources",
"targets_dir":    "/tmp/targets",
"targets_file_suffix": "_sd_targets"
}`)
	MockTestCert = []byte(`-----BEGIN CERTIFICATE-----
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
prDLZKx0Xw3n
-----END CERTIFICATE-----`)
	MockTestKey = []byte(`-----BEGIN PRIVATE KEY-----
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
R3zzWGC9q0+JqmI7jKW/g9d5TBvy
-----END PRIVATE KEY-----`)
)
