package core

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"syscall"

	"gopkg.in/yaml.v3"
)

const PermStdRead = 0o644

var (
	tester func(string) error
	reader func(string) ([]byte, error)
	// func(name string, data []byte, perm os.FileMode) error
	writer func(string, []byte, os.FileMode) error
)

func init() {
	tester = AssertReadable
	reader = os.ReadFile
	writer = os.WriteFile
}

func SetTester(t func(string) error) {
	tester = t
}

func SetReader(t func(string) ([]byte, error)) {
	reader = t
}

func SetWriter(t func(string, []byte, os.FileMode) error) {
	writer = t
}

func ReadYAML[T any](file string, obj *T) error {
	data, err := ReadFile(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, obj)
}

func ReadJSON[T any](file string, obj *T) error {
	data, err := ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, obj)
}

func WriteYAML[T any](file string, obj *T, perm os.FileMode) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	return WriteFile(file, data, perm)
}

func WriteJSON[T any](file string, obj *T, perm os.FileMode) error {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	return WriteFile(file, data, perm)
}

// FindInDir returns the contents of the first file found in the directory. dir should not contain
// a trailing "/". If no file is found, return FileNotFound.
func FindInDir(dir string, fileNames ...string) (string, error) {
	switch dir {
	case "":
		break
	case "/":
		break
	default:
		dir = dir + "/"
	}

	if len(fileNames) == 0 || (len(fileNames) == 1 && fileNames[0] == "") {
		return "", os.ErrInvalid
	}

	// Check default locations for the file.
	for _, name := range fileNames {
		err := tester(dir + name)
		if err == nil {
			return dir + name, nil
		}
	}

	return "", os.ErrNotExist
}

func ReadFile(file string) ([]byte, error) {
	if file == "" {
		return nil, os.ErrInvalid
	}

	return reader(file)
}

func WriteFile(file string, data []byte, perm os.FileMode) error {
	if file == "" {
		return os.ErrInvalid
	}

	return writer(file, data, perm)
}

func MapFiles(docRoot string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(docRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == "includes" {
				return filepath.SkipDir
			}

			return nil
		}

		// Make sure we can read the file.
		if err := AssertReadable(path); err != nil {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// CheckReadability returns nil if file is readable.
func AssertReadable(file string) error {
	if file == "" {
		return fmt.Errorf("read: %w", os.ErrInvalid)
	}

	s, err := os.Stat(file)
	if err != nil {
		return err
	}

	if s.IsDir() {
		return fmt.Errorf("read %s: is a directory", file)
	}

	err = HasReadPerm(s)
	if err != nil {
		return err
	}

	return nil
}

// HasReadPerm returns nil if the app has read permission to the file.
func HasReadPerm(info fs.FileInfo) error {
	if info == nil {
		return fmt.Errorf("read: %w", os.ErrInvalid)
	}

	u, err := user.Current()
	if err != nil {
		return err
	}

	if info.Mode().Perm()&0o004 == 0o004 {
		return nil
	}

	fileUid := fmt.Sprint(info.Sys().(*syscall.Stat_t).Uid)
	if u.Uid == fileUid {
		if info.Mode().Perm()&0o400 == 0o400 {
			return nil
		}
	}

	groups, err := u.GroupIds()
	if err != nil {
		return err
	}

	fileGid := fmt.Sprint(info.Sys().(*syscall.Stat_t).Gid)
	for _, group := range groups {
		if group == fileGid {
			if info.Mode().Perm()&0o040 == 0o040 {
				return nil
			}
		}
	}

	return fmt.Errorf("read: %w", os.ErrPermission)
}
