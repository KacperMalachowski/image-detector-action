package cmd

import "testing"

type MockedDetector struct {
	Supported bool
	Called bool
	images []string
}

func (d *MockedDetector) Detect(file string) ([]string, error) {
	d.Called = true
	if d.Supported {
		return d.images, nil
	}
	return nil, nil
}

func (d *MockedDetector) IsSupported(file string) bool {
	return d.Supported
}

// TestDetectImageURLs tests the detectImageURLs function
// It should identify image URLs using given Detectors.
// If the given file is supproted by teh Detector it should run Detect function
func TestDetectImageURLs(t *testing.T) {
	tc := []struct {
		name string
		detector Detector
		file string
		expected []string
	}{
		{
			name: "File supported by detector and has images",
			detector: &MockedDetector{
				Supported: true,
				images: []string{"europe.dev.pkg/docker/test:123", "europe.dev.pkg/docker/test:456"},
			},
			file: "",
	}
}
