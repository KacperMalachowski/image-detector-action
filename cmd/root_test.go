package cmd

import (
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockDetector struct {
	visitedFiles     []string
	supportedPattern string
	images           []string
	err              bool
}

func (d *MockDetector) Detect(file io.Reader) ([]string, error) {
	if d.err {
		return nil, assert.AnError
	}
	return d.images, nil
}

func (d *MockDetector) IsSupported(file string) bool {
	d.visitedFiles = append(d.visitedFiles, file)

	if d.supportedPattern == "" {
		return true
	}

	matched, _ := doublestar.Match(d.supportedPattern, file)
	return matched
}

func TestSetOutput(t *testing.T) {
	tests := []struct {
		name   string
		images []string
	}{
		{
			name:   "Single image",
			images: []string{"test/image:1234"},
		},
		{
			name:   "Multiple images",
			images: []string{"test/image1:1234", "test/image2:5678"},
		},
		{
			name:   "Empty images",
			images: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "")
			assert.NoError(t, err, "Failed to create temporary file")
			defer os.Remove(file.Name())

			t.Setenv("GITHUB_OUTPUT", file.Name())

			setImagesOutput(tt.images)

			actualImages := getImagesOutput(t, file)
			assert.Equal(t, tt.images, actualImages, "Images output does not match expected")
		})
	}
}

func TestFindImagesInFiles(t *testing.T) {
	tc := []struct {
		name                 string
		exludes              []string
		supportedPattern     string
		images               []string
		prepareDirectoryTree func(path string)
	}{
		{
			name:             "No files",
			exludes:          []string{},
			supportedPattern: "**/*.txt",
			images:           []string{},
			prepareDirectoryTree: func(path string) {
				// No files to prepare
			},
		},
		{
			name:             "All files included",
			exludes:          []string{},
			supportedPattern: "**/*.tf",
			images:           []string{"test/image1:1234", "test/image2:5678"},
			prepareDirectoryTree: func(path string) {
				// Create a directory structure with files
				os.MkdirAll(path+"/dir1", 0755)
				os.WriteFile(path+"/dir1/file1.tf", []byte("content"), 0644)
				os.WriteFile(path+"/dir1/file2.tf", []byte("content"), 0644)
				os.WriteFile(path+"/dir1/file3.tf", []byte("content"), 0644)
			},
		},
		{
			name:             "Excluded files",
			exludes:          []string{"**/excluded.txt"},
			supportedPattern: "**/*.txt",
			images:           []string{"test/image1:1234"},
			prepareDirectoryTree: func(path string) {
				// Create a directory structure with files
				os.MkdirAll(path+"/dir1", 0755)
				os.WriteFile(path+"/dir1/file1.txt", []byte("content"), 0644)
				os.WriteFile(path+"/dir1/excluded.txt", []byte("content"), 0644)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			tempDir := t.TempDir()

			mockedDetector := &MockDetector{
				supportedPattern: test.supportedPattern,
				images:           test.images,
			}
			detectors := []Detector{mockedDetector}

			test.prepareDirectoryTree(tempDir)

			images, err := findImagesInFiles(tempDir, detectors, test.exludes)
			assert.NoError(t, err, "Failed to find images in files")

			// Check if the excluded files were not processed
			for _, excluded := range test.exludes {
				for _, visited := range mockedDetector.visitedFiles {
					assert.NotContains(t, visited, excluded, "Excluded file was processed: %s", excluded)
				}
			}

			// Check if not supported files were not processed
			for _, visited := range mockedDetector.visitedFiles {
				matched, _ := doublestar.Match(mockedDetector.supportedPattern, visited)
				if !matched {
					assert.NotContains(t, images, visited, "Unsupported file was processed: %s", visited)
				}
			}

			// Check if supported files were processed
			for _, visited := range mockedDetector.visitedFiles {
				if match, _ := doublestar.Match(mockedDetector.supportedPattern, visited); match {
					assert.Contains(t, mockedDetector.visitedFiles, visited, "Supported file was not processed: %s", visited)
				}
			}

			assert.ElementsMatch(t, test.images, images, "Detected images do not match expected")
		})
	}
}

// getImagesOutputFromPath reads the GITHUB_OUTPUT file and perses the "images" key
func getImagesOutput(t *testing.T, file io.Reader) []string {
	t.Helper()

	data, err := io.ReadAll(file)
	require.NoError(t, err, "Failed to read GITHUB_OUTPUT file")

	s := string(data)

	var lines []string

	// Regex: Runs on multiline input to find the "images" key
	// and captures everything until the next delimiter
	// The delimiter is defined as _GitHubActionsFileCommandDelimeter_
	re := regexp.MustCompile(`(?ms)^images<<_GitHubActionsFileCommandDelimeter_\r?\n(.*)\r?\n_GitHubActionsFileCommandDelimeter_`)
	matches := re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		content := matches[1]

		// Normalize CRLF and split
		content = strings.ReplaceAll(content, "\r\n", "\n")
		lines = strings.Split(content, "\n")
		// Remove empty lines
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}

	// Fallback: support single line key=value format
	// images=["test/image1:1234", "test/image2:5678"]
	if m := regexp.MustCompile(`(?m)^images=(.+)$`).FindStringSubmatch(s); len(m) == 2 {
		val := strings.TrimSpace(m[1])

		lines = []string{val}
	}

	if len(lines) == 0 {
		require.Fail(t, "No images found in GITHUB_OUTPUT", matches)
	}

	// Parse JSON array containing image names
	var images []string
	err = json.Unmarshal([]byte(lines[0]), &images)
	require.NoError(t, err, "Failed to parse images from GITHUB_OUTPUT")

	return images
}
