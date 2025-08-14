package strategy_test

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/KacperMalachowski/image-detector-action/pkg/strategy"
)

func TestGenericIsSupported(t *testing.T) {
	tc := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Generic strategy supports all files",
			path: "anyfile.txt",
			want: true,
		},
		{
			name: "Generic strategy supports terraform file",
			path: "main.tf",
			want: true,
		},
		{
			name: "Generic strategy supports markdown file",
			path: "README.md",
			want: true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			generic := &strategy.Generic{}
			got := generic.IsSupported(tt.path)
			if got != tt.want {
				t.Errorf("IsSupported(%s) = %v; want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGenericDetect(t *testing.T) {
	tc := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "Detects docker image URLs in text",
			data: "docker.io/library/ubuntu:latest\nregistry.gitlab.com/mygroup/myproject/myimage:1.0\n",
			want: []string{"docker.io/library/ubuntu:latest", "registry.gitlab.com/mygroup/myproject/myimage:1.0"},
		},
		{
			name: "Detects docker image URLs in markdown",
			data: "![Image](docker.io/library/ubuntu:latest)\n![Image](registry.gitlab.com/mygroup/myproject/myimage:1.0)\n",
			want: []string{"docker.io/library/ubuntu:latest", "registry.gitlab.com/mygroup/myproject/myimage:1.0"},
		},
		{
			name: "Detects docker image URLs in terraform",
			data: `resource "docker_image" "example" {
				name = "docker.io/library/ubuntu:latest"
			}`,
			want: []string{"docker.io/library/ubuntu:latest"},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			generic := &strategy.Generic{}
			reader := io.NopCloser(strings.NewReader(tt.data))
			got, err := generic.Detect(reader)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Detect() = %v; want %v", got, tt.want)
			}
		})
	}
}
