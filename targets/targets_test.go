package targets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chadeldridge/prometheus-import-manager/core"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Create test JSON content
var (
	expectedTargetGroups = TargetGroups{
		&TargetGroup{
			Jobs: []string{"blackbox_icmp"},
			Labels: map[string]string{
				"environment": "prod",
				"datacenter":  "us-east-1",
				"role":        "monitoring",
			},
			Targets: []string{"prom.example.com", "grafana.example.com"},
		},
		&TargetGroup{
			Jobs: []string{"node-exporter", "mysql-exporter"},
			Labels: map[string]string{
				"environment": "stg",
				"datacenter":  "us-east-1",
			},
			Targets: []string{"node1.example.com", "node2.example.com"},
		},
	}
	splitTargetGroups = TargetMap{
		"blackbox_icmp_targets.yml": {
			&TargetGroup{
				Jobs: []string{"blackbox_icmp"},
				Labels: map[string]string{
					"environment": "prod",
					"datacenter":  "us-east-1",
					"role":        "monitoring",
				},
				Targets: []string{"prom.example.com", "grafana.example.com"},
			},
		},
		"node-exporter_targets.yml": {
			&TargetGroup{
				Jobs: []string{"node-exporter"},
				Labels: map[string]string{
					"environment": "stg",
					"datacenter":  "us-east-1",
				},
				Targets: []string{"node1.example.com", "node2.example.com"},
			},
		},
		"mysql-exporter_targets.yml": {
			&TargetGroup{
				Jobs: []string{"mysql-exporter"},
				Labels: map[string]string{
					"environment": "stg",
					"datacenter":  "us-east-1",
				},
				Targets: []string{"node1.example.com", "node2.example.com"},
			},
		},
	}
	jsonContent = `[
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
	// Create test YAML content
	yamlContent = `- jobs:
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
)

func newExpectedTargetGroups() *TargetGroups {
	return &TargetGroups{
		&TargetGroup{
			Jobs: []string{"blackbox_icmp"},
			Labels: map[string]string{
				"environment": "prod",
				"datacenter":  "us-east-1",
				"role":        "monitoring",
			},
			Targets: []string{"prom.example.com", "grafana.example.com"},
		},
		&TargetGroup{
			Jobs: []string{"node-exporter", "mysql-exporter"},
			Labels: map[string]string{
				"environment": "stg",
				"datacenter":  "us-east-1",
			},
			Targets: []string{"node1.example.com", "node2.example.com"},
		},
	}
}

func TestTargetGroup(t *testing.T) {
	t.Run("JSONTags", func(t *testing.T) {
		testTargetGroupTags(t, "JSON")
	})

	t.Run("YAMLTags", func(t *testing.T) {
		testTargetGroupTags(t, "YAML")
	})
}

func testTargetGroupTags(t *testing.T, format string) {
	require := require.New(t)
	// Test YAML marshaling/unmarshaling with proper tags
	tg := &TargetGroup{
		Jobs:    []string{"prometheus"},
		Labels:  map[string]string{"environment": "test"},
		Targets: []string{"localhost:9090"},
	}
	expYAML := `jobs:
    - prometheus
labels:
    environment: test
targets:
    - localhost:9090
`
	expJSON := `{"jobs":["prometheus"],"labels":{"environment":"test"},"targets":["localhost:9090"]}`

	// Marshal
	var data []byte
	var err error
	var exp string
	switch format {
	case "YAML":
		data, err = yaml.Marshal(tg)
		exp = expYAML
	case "JSON":
		data, err = json.Marshal(tg)
		exp = expJSON
	}
	require.NoError(err, "failed to marshal TargetGroup to %s", format)
	// Convert data to string, instead of exp to []byte, for better
	// readability if the test fails.
	require.Equal(exp, string(data), "data did not match expected %s", format)

	// Unmarshal back
	var got TargetGroup
	switch format {
	case "YAML":
		err = yaml.Unmarshal(data, &got)
	case "JSON":
		err = json.Unmarshal(data, &got)
	}
	require.NoError(err, "failed to unmarshal %s to TargetGroup", format)

	// Verify the data is preserved. Dereference tg since Unmarshal does not
	// return a refereced TargetGroup.
	require.Equal(*tg, got, "got did not match original tg")
}

func TestTargetGroups(t *testing.T) {
	require := require.New(t)
	t.Run("Types", func(t *testing.T) {
		// Test TargetGroups slice type
		var tgs TargetGroups

		// Test that we can append to it
		tg1 := NewTargetGroup()
		tg1.Jobs = []string{"test-job"}

		tgs = append(tgs, tg1)

		require.Len(tgs, 1, "wrong number of items in TargetGroups")
		require.Equal(tg1, tgs[0], "TargetGroups did not match")
	})
}

func TestNewTargetGroup(t *testing.T) {
	require := require.New(t)
	tg := NewTargetGroup()

	require.NotNil(tg, "NewTargetGroup retuned a nil TargetGroup")
	require.NotNil(tg.Jobs, "Jobs was nil")
	require.NotNil(tg.Labels, "Labels was nil")
	require.NotNil(tg.Targets, "Targets was nil")

	require.Empty(tg.Jobs, "Jobs was not empty")
	require.Empty(tg.Labels, "Labels was not empty")
	require.Empty(tg.Targets, "Targets was not empty")

	// Test that we can add items to the initialized collections
	tg.Labels["environment"] = "prod"
	tg.Targets = append(tg.Targets, "prom.example.com")
	tg.Jobs = append(tg.Jobs, "blackbox_icmp")
	require.Len(tg.Jobs, 1, "Jobs did not contain the correct number of items")
	require.Len(tg.Labels, 1, "Labels did not contain the correct number of items")
	require.Len(tg.Targets, 1, "Targets did not contain the correct number of items")
}

func TestNewTargetGroups(t *testing.T) {
	t.Run("EmptyFile", func(t *testing.T) {
		testNewTargetGroupsEmptyFile(t)
	})

	// Test YAML source
	t.Run("InvalidYAML", func(t *testing.T) {
		// Create invalid YAML content
		invalidYAML := `- jobs:
    - prometheus
  labels:
    env: production
  targets: [unclosed bracket`

		testNewTargetGroupsInvalidSyntax(t, "targets.yml", invalidYAML)
	})

	t.Run("YAML", func(t *testing.T) {
		testNewTargetGroups(t, "targets.yml", yamlContent)
	})

	// Test JSON source
	t.Run("InvalidJSON", func(t *testing.T) {
		// Test the JSON error path (line 42-43: return nil, err)
		invalidJSON := `[
  {
    "jobs": ["test"],
    "labels": {
      "environment": "test"
    },
    "targets": ["localhost:8080"
    // Missing closing bracket and quotes - invalid JSON
]`

		testNewTargetGroupsInvalidSyntax(t, "targets.json", invalidJSON)
	})

	t.Run("JSON", func(t *testing.T) {
		testNewTargetGroups(t, "targets.json", jsonContent)
	})
}

func testNewTargetGroupsEmptyFile(t *testing.T) {
	require := require.New(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "targets_test")
	require.NoError(err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	// Create empty YAML content
	emptyYAML := ``

	// Create the sources directory structure
	sourcesDir := filepath.Join(tempDir, "sources")
	err = os.MkdirAll(sourcesDir, 0o755)
	require.NoError(err, "failed to create sources dir")

	// Write the empty YAML file
	yamlFile := filepath.Join(sourcesDir, "targets.yml")
	err = os.WriteFile(yamlFile, []byte(emptyYAML), 0o644)
	require.NoError(err, "failed to write empty sources file")

	// Create config pointing to our test directory
	config := core.DefaultConfig()
	config.Sources = sourcesDir

	// Test loading the target groups
	targetGroups, err := NewTargetGroups(config)
	require.Error(err, "failed to load targets source file")
	require.Nil(targetGroups, "targetGroups was nil")
}

func testNewTargetGroupsInvalidSyntax(t *testing.T, file, content string) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "targets_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the sources directory structure
	sourcesDir := filepath.Join(tempDir, "sources")
	err = os.MkdirAll(sourcesDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Write the invalid YAML file
	yamlFile := filepath.Join(sourcesDir, file)
	err = os.WriteFile(yamlFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to write invalid content to test file: %v", err)
	}

	// Create config pointing to our test directory
	config := core.DefaultConfig()
	config.Sources = sourcesDir

	// Test loading the target groups
	targetGroups, err := NewTargetGroups(config)

	if err == nil {
		t.Error("Expected error when loading invalid content, got nil")
	}

	if targetGroups != nil {
		t.Error("Expected targetGroups to be nil when error occurs, got non-nil")
	}
}

func testNewTargetGroups(t *testing.T, file string, content string) {
	require := require.New(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "targets_test")
	require.NoError(err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	// Create the sources directory structure
	sourcesDir := filepath.Join(tempDir, "sources")
	err = os.MkdirAll(sourcesDir, 0o755)
	require.NoError(err, "failed to create sources dir")

	// Write the test YAML file
	targetsFile := filepath.Join(sourcesDir, file)
	err = os.WriteFile(targetsFile, []byte(content), 0o644)
	require.NoError(err, "failed to write test file")

	// Create config pointing to our test directory
	config := core.DefaultConfig()
	config.Sources = sourcesDir

	// Test loading the target groups
	targetGroups, err := NewTargetGroups(config)
	require.NoError(err, "failed to load target groups")
	require.NotNil("TargetGroups was nil")
	require.Len(targetGroups, len(expectedTargetGroups), "length of target groups did not match")

	// Verify first target group
	for i, tg := range targetGroups {
		require.Equal(expectedTargetGroups[i].Jobs, tg.Jobs, "Jobs %s did not match", i)
		require.Equal(expectedTargetGroups[i].Labels, tg.Labels, "Labels %s did not match", i)
		require.Equal(expectedTargetGroups[i].Targets, tg.Targets, "Targets %s did not match", i)
	}
}

func TestReadSources(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		testReadSources(t, "targets.txt")
	})
	t.Run("yml", func(t *testing.T) {
		testReadSources(t, "targets.yml")
	})
	t.Run("yaml", func(t *testing.T) {
		testReadSources(t, "targets.yaml")
	})
	t.Run("json", func(t *testing.T) {
		testReadSources(t, "targets.json")
	})
}

func testReadSources(t *testing.T, file string) {
	require := require.New(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "targets_test")
	require.NoError(err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	// Create the sources directory structure
	sourcesDir := filepath.Join(tempDir, "sources")
	err = os.MkdirAll(sourcesDir, 0o755)
	require.NoError(err, "failed to create sources dir")

	tgs := TargetGroups{
		&TargetGroup{
			Jobs:    []string{"prometheus"},
			Labels:  map[string]string{"environment": "test"},
			Targets: []string{"localhost:9090"},
		},
	}
	var content string
	contentYAML := `- jobs:
    - prometheus
  labels:
    environment: test
  targets:
    - localhost:9090`
	contentJSON := `[{
  "jobs": [
    "prometheus"
  ],
  "labels": {
    "environment":"test"
  },
  "targets":[
    "localhost:9090"
  ]
}]`

	if strings.HasSuffix(file, core.DefaultYAMLFileExt) || strings.HasSuffix(file, "yaml") {
		content = contentYAML
	} else {
		content = contentJSON
	}

	targetsFile := filepath.Join(sourcesDir, file)
	err = os.WriteFile(targetsFile, []byte(content), 0o644)
	require.NoError(err, "failed to write test file: %s", file)

	got, err := readSources([]string{targetsFile})

	if strings.HasSuffix(file, "txt") {
		require.Error(err, "readSources did not return an error")
		require.ErrorIs(err, os.ErrInvalid, "readSources did not return the correct error")
		require.Nil(got, "readSources did not return a nil TargetGroups")
		return
	}

	require.NoError(err, "readSources did not return an error")
	require.NotNil(got, "readSources returned a nil TargetGroups")
	require.Equal(tgs, got, "TargetGroups did not match")
}

func TestSplitByJob(t *testing.T) {
	require := require.New(t)

	t.Run("EmptyTargetGroups", func(t *testing.T) {
		config := core.DefaultConfig()
		config.TargetsFileExt = core.DefaultYAMLFileExt

		got := TargetGroups{}.splitByJob(config)
		// require.NotNil(got, "splitByJob returned nil TargetGroups")
		require.Empty(got, "splitByJob returned a non-empty TargetGroups")
	})

	t.Run("Successful", func(t *testing.T) {
		tgs := newExpectedTargetGroups()
		config := core.DefaultConfig()
		config.TargetsFileExt = core.DefaultYAMLFileExt

		got := tgs.splitByJob(config)
		require.NotNil(got, "splitByJob returned nil TargetGroups")
		require.NotEmpty(got, "splitByJob returned an empty TargetGroups")
		require.Equal(splitTargetGroups, got, "TargetGroups did not match")
	})
}

func TestWriteTargets(t *testing.T) {
	t.Run("YAML", func(t *testing.T) {
		testWriteTargets(t, core.DefaultYAMLFileExt)
	})

	t.Run("JSON", func(t *testing.T) {
		testWriteTargets(t, core.DefaultJSONFileExt)
	})
}

func testWriteTargets(t *testing.T, ext string) {
	require := require.New(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "write_targets_test")
	require.NoError(err, "failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Create config pointing to our temp directory
	config := core.DefaultConfig()
	config.TargetsFileExt = ext
	config.TargetsDir = tempDir

	// Create sample target groups
	filename := "blackbox_icmp" + core.DefaultTargetsFileSuffix + config.TargetsFileExt
	targetGroups := make(TargetMap)
	targetGroups[filename] = TargetGroups{
		&TargetGroup{
			Jobs:    []string{"blackbox_icmp"},
			Labels:  map[string]string{"environment": "dev"},
			Targets: []string{"prom.example.com"},
		},
	}

	// Write the targets
	err = writeTargets(config, targetGroups)
	require.NoError(err, "failed to write targets")

	// testYAMLFile(t, filepath.Join(tempDir, filename), targetGroups, filename)
	testFile(t, tempDir, targetGroups)
}

// Read the YAML test file and compare it to content.
// func testYAMLFile(t *testing.T, file string, tgs TargetMap, filename string) {
func testFile(t *testing.T, dir string, targets TargetMap) {
	require := require.New(t)
	for filename, tgs := range targets {
		file := filepath.Join(dir, filename)
		var expectedData []byte
		var err error
		if strings.HasSuffix(file, ".json") {
			expectedData, err = json.MarshalIndent(&tgs, "", "  ")
		} else {
			expectedData, err = yaml.Marshal(&tgs)
		}
		require.NoError(err, "failed to marshal expected target groups to YAML")
		require.FileExists(file, "targets file does not exist")

		// Read and verify the file content
		data, err := os.ReadFile(file)
		require.NoError(err, "failed to read targets file")
		require.Equal(data, []byte(expectedData), "written file content does not match")
	}
}

func TestExportTargets_YAML(t *testing.T) {
	t.Run("YAML", func(t *testing.T) {
		testExportTargets(t, ".yml")
	})

	t.Run("JSON", func(t *testing.T) {
		testExportTargets(t, ".json")
	})
}

func testExportTargets(t *testing.T, ext string) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "write_targets_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config pointing to our temp directory
	config := core.DefaultConfig()
	config.TargetsDir = tempDir
	if ext == "json" {
		config.TargetsFileExt = core.DefaultJSONFileExt
	} else {
		config.TargetsFileExt = core.DefaultYAMLFileExt
	}

	expect := expectedTargetGroups.splitByJob(config)
	err = expectedTargetGroups.ExportTargets(config)
	if err != nil {
		t.Fatalf("Failed to export targets: %v", err)
	}

	// for filename := range expect {
	// testYAMLFile(t, filepath.Join(tempDir, filename), expect, filename)
	testFile(t, tempDir, expect)
	//}
}
