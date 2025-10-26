package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func createTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	return tempDir
}

func writeFile(file string, config *Config) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0o600)
}

func removeFile(file string) error {
	return os.Remove(file)
}

func TestFilesSetTester(t *testing.T) {
	require := require.New(t)
	SetTester(MockTester)
	MockWriteFile("test.file", []byte("test"), true, nil)

	err := tester("test.file")
	require.NoError(err, "tester() returned an error: %s", err)
}

func TestFilesSetReader(t *testing.T) {
	require := require.New(t)
	SetReader(MockReader)
	MockWriteFile("test.file", []byte("test"), true, nil)

	data, err := reader("test.file")
	require.NoError(err, "reader() returned an error: %s", err)
	require.Equal([]byte("test"), data, "reader() did not return the correct data")
}

func TestFilesParseYAML(t *testing.T) {
	require := require.New(t)
	reader = MockReader
	MockWriteFile("test.file", []byte("test: test"), true, nil)

	var out map[string]string
	err := ReadYAML("test.file", &out)
	require.NoError(err, "ParseYAML() returned an error: %s", err)
	require.NotNil(out, "ParseYAML() set out to nil")
	require.Equal("test", out["test"], "ParseYAML() did not return the correct data")
}

func TestFilesParseJSON(t *testing.T) {
	require := require.New(t)
	reader = MockReader
	MockWriteFile("test.file", []byte(`{"test": "test"}`), true, nil)

	var out map[string]string
	err := ReadJSON("test.file", &out)
	require.NoError(err, "ParseJSON() returned an error: %s", err)
	require.NotNil(out, "ParseJSON() set out to nil")
	require.Equal("test", out["test"], "ParseJSON() did not return the correct data")
}

func TestFilesFindInDir(t *testing.T) {
	require := require.New(t)
	tester = MockTester

	t.Run("no filenames", func(t *testing.T) {
		f, err := FindInDir("")
		require.Error(err, "FindInDir() did not return an error")
		require.ErrorIs(err, os.ErrInvalid, "FindInDir() did not return the expected error")
		require.Equal("", f, "FindInDir() did not return expected file string")
	})

	t.Run("empty filename", func(t *testing.T) {
		f, err := FindInDir("", "")
		require.Error(err, "FindInDir() returned an error: %s", err)
		require.ErrorIs(err, os.ErrInvalid, "FindInDir() did not return the expected error")
		require.Equal("", f, "FindInDir() did not return the expected file string")
	})

	t.Run("empty dir", func(t *testing.T) {
		MockWriteFile("test.file", []byte(""), true, nil)
		f, err := FindInDir("", "test.file")
		require.NoError(err, "FindInDir() returned an error: %s", err)
		require.Equal("test.file", f, "FindInDir() did not return the expected file string")
	})

	t.Run("root dir", func(t *testing.T) {
		MockWriteFile("/test.file", []byte(""), true, nil)
		f, err := FindInDir("/", "test.file")
		require.NoError(err, "FindInDir() returned an error: %s", err)
		require.Equal("/test.file", f, "FindInDir() did not return the expected file string")
	})

	t.Run("valid dir", func(t *testing.T) {
		MockWriteFile("/home/tester/test.file", []byte(""), true, nil)
		f, err := FindInDir("/home/tester", "test.file")
		require.NoError(err, "FindInDir() returned an error: %s", err)
		require.Equal("/home/tester/test.file", f, "FindInDir() did not return the expected file string")
	})

	t.Run("return first", func(t *testing.T) {
		MockWriteFile("/home/tester/test.file", []byte(""), true, nil)
		f, err := FindInDir("/home/tester", "test.file", "test2.file")
		require.NoError(err, "FindInDir() returned an error: %s", err)
		require.Equal("/home/tester/test.file", f, "FindInDir() did not return the expected file string")
	})

	t.Run("return second", func(t *testing.T) {
		MockClearFS()
		MockWriteFile("/home/tester/test2.file", []byte(""), true, nil)
		f, err := FindInDir("/home/tester", "test.file", "test2.file")
		require.NoError(err, "FindInDir() returned an error: %s", err)
		require.Equal("/home/tester/test2.file", f, "FindInDir() did not return the expected file string")
	})

	t.Run("not found", func(t *testing.T) {
		MockClearFS()
		f, err := FindInDir("/home/tester", "test.file")
		require.Error(err, "FindInDir() did not return an error")
		require.Equal(os.ErrNotExist, err, "FindInDir() did not return the expected error")
		require.Equal("", f, "FindInDir() did not return the expected file string")
	})
}

func TestFilesReadFile(t *testing.T) {
	require := require.New(t)
	reader = MockReader

	t.Run("no file", func(t *testing.T) {
		MockClearFS()
		_, err := ReadFile("")
		require.Error(err, "ReadFile() did not return an error")
		require.ErrorIs(err, os.ErrInvalid, "ReadFile() did not return the expected error")
	})

	t.Run("missing file", func(t *testing.T) {
		MockWriteFile("test.file", []byte(""), true, os.ErrNotExist)
		_, err := ReadFile("test.file")
		require.Error(err, "ReadFile() did not return an error")
		require.ErrorIs(err, os.ErrNotExist, "ReadFile() did not return the expected error")
	})

	t.Run("valid file", func(t *testing.T) {
		MockWriteFile("test.file", []byte("test"), true, nil)
		data, err := ReadFile("test.file")
		require.NoError(err, "ReadFile() returned an error: %s", err)
		require.Equal([]byte("test"), data, "ReadFile() did not return the expected data")
	})
}

func TestFilesAssertReadable(t *testing.T) {
	require := require.New(t)
	tempDir := createTempDir(t)

	defer os.RemoveAll(tempDir)

	t.Run("no file", func(t *testing.T) {
		err := AssertReadable("")
		require.Error(err, "AssertReadable() did not return an error")
		require.ErrorIs(err, os.ErrInvalid, "AssertReadable() did not return the expected error")
	})

	t.Run("missing file", func(t *testing.T) {
		err := AssertReadable("test.file")
		require.Error(err, "AssertReadable() did not return an error")
		require.Contains(err.Error(), "no such file or directory", "AssertReadable() did not return the expected error")
	})

	t.Run("is dir", func(t *testing.T) {
		err := AssertReadable("/tmp")
		require.Error(err, "AssertReadable() did not return an error")
		require.Equal("read /tmp: is a directory", err.Error(), "AssertReadable() did not return the expected error")
	})

	t.Run("no permissions", func(t *testing.T) {
		err := AssertReadable("/etc/shadow")
		require.Error(err, "AssertReadable() did not return an error")
		require.Contains(err.Error(), "permission denied", "AssertReadable() did not return the expected error")
	})

	t.Run("valid file", func(t *testing.T) {
		file := filepath.Join(tempDir, "test.file")
		writeFile(file, DefaultConfig())
		err := AssertReadable(file)
		require.NoError(err, "AssertReadable() returned an error: %s", err)
	})
}

func TestFilesHasReadPerm(t *testing.T) {
	require := require.New(t)
	tempDir := createTempDir(t)
	file := filepath.Join(tempDir, "test.file")

	defer os.RemoveAll(tempDir)

	t.Run("other read", func(t *testing.T) {
		writeFile(file, DefaultConfig())
		err := os.Chmod(file, 0o666)
		require.NoError(err, "os.Chmod() returned an error: %s", err)

		s, err := os.Stat(file)
		require.NoError(err, "os.Stat() returned an error: %s", err)

		err = HasReadPerm(s)
		require.NoError(err, "HasReadPerm() did not return an error")
		removeFile(file)
	})

	t.Run("user read", func(t *testing.T) {
		writeFile(file, DefaultConfig())
		err := os.Chmod(file, 0o660)
		require.NoError(err, "os.Chmod() returned an error: %s", err)

		s, err := os.Stat(file)
		require.NoError(err, "os.Stat() returned an error: %s", err)

		err = HasReadPerm(s)
		require.NoError(err, "HasReadPerm() did not return an error")
		removeFile(file)
	})

	t.Run("group read", func(t *testing.T) {
		writeFile(file, DefaultConfig())
		err := os.Chmod(file, 0o600)
		require.NoError(err, "os.Chmod() returned an error: %s", err)

		s, err := os.Stat(file)
		require.NoError(err, "os.Stat() returned an error: %s", err)

		err = HasReadPerm(s)
		require.NoError(err, "HasReadPerm() did not return an error")
		removeFile(file)
	})

	t.Run("no permissions", func(t *testing.T) {
		s, err := os.Stat("/etc/shadow")
		require.NoError(err, "os.Stat() returned an error: %s", err)

		err = HasReadPerm(s)
		require.Error(err, "AssertReadable() did not return an error")
		require.Contains(err.Error(), "permission denied", "AssertReadable() did not return the expected error")
	})
}
