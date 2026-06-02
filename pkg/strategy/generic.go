package strategy

import (
	"io"
	"regexp"
)

type Generic struct {}


// Generic is a strategy that supports all files
func (g *Generic) IsSupported(_ string) bool {
	return true
}

// Detect scans the provided file for image URLs and returns them.
func (g *Generic) Detect(file io.Reader) ([]string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`([a-z0-9]+(?:[.-][a-z0-9]+)*/)*([a-z0-9]+(?:[.-][a-z0-9]+)*)(?::[a-z0-9.-]+)?/([a-z0-9-]+)/([a-z0-9-]+)(?::[a-z0-9.-]+)`)
	matches := re.FindAllString(string(data), -1)

	var images []string
	for _, match := range matches {
		if match != "" {
			images = append(images, match)
		}
	}

	return images, nil
}

