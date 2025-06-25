package includes

import "testing"

func TestPrepareRelPath(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		add      string
		to       string
		want     string
	}{
		{
			name:     "root include",
			fileName: "docs/examples/includes/000_common/.werf/meta/cleanup.yaml",
			add:      "/docs/examples/includes/000_common",
			to:       "/",
			want:     ".werf/meta/cleanup.yaml",
		},
		{
			name:     "custom to",
			fileName: "docs/examples/includes/010_js_framework/.helm/Chart.yaml",
			add:      "/docs/examples/includes/010_js_framework",
			to:       "/my-app",
			want:     "my-app/.helm/Chart.yaml",
		},
		{
			name:     "deep nested",
			fileName: "docs/examples/includes/010_js_framework/.helm/charts/frontend/templates/deployment.yaml",
			add:      "/docs/examples/includes/010_js_framework",
			to:       "/project",
			want:     "project/.helm/charts/frontend/templates/deployment.yaml",
		},
		{
			name:     "no leading slash in add",
			fileName: "docs/examples/includes/000_common/.werf/meta/imagespec.yaml",
			add:      "docs/examples/includes/000_common",
			to:       "/",
			want:     ".werf/meta/imagespec.yaml",
		},
		{
			name:     "relative to",
			fileName: "docs/examples/includes/000_common/backend.Dockerfile",
			add:      "docs/examples/includes/000_common",
			to:       "backend",
			want:     "backend/backend.Dockerfile",
		},
		{
			name:     "add is root, to is root",
			fileName: "README.md",
			add:      "/",
			to:       "/",
			want:     "README.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareRelPath(tt.fileName, tt.add, tt.to)
			if got != tt.want {
				t.Errorf("prepareRelPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
