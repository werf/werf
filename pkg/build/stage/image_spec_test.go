package stage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/werf/werf/v2/pkg/config"
)

func TestEnvExpander(t *testing.T) {
	t.Run("basic expand", func(t *testing.T) {
		env, err := modifyEnv([]string{"PATH=/usr/bin"}, []string{}, map[string]string{"PATH": "${PATH}:/add/path"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"PATH=/usr/bin:/add/path"}, env)
	})
	t.Run("multiple expand", func(t *testing.T) {
		existed := []string{
			"PATH=/usr/bin",
			"TEST=test",
		}
		specified := map[string]string{
			"PATH":   "${PATH}:/some/path",
			"GOPATH": "/bin/go",
			"GOROOT": "/usr/local/go",
		}
		env, err := modifyEnv(existed, []string{}, specified)
		assert.NoError(t, err)
		expceted := []string{
			"PATH=/usr/bin:/some/path",
			"GOROOT=/usr/local/go",
			"GOPATH=/bin/go",
			"TEST=test",
		}
		for _, e := range env {
			assert.Contains(t, expceted, e)
		}
	})
	t.Run("multiple expand witch circular dependency", func(t *testing.T) {
		existed := []string{
			"PATH=/usr/bin",
			"TEST=test",
		}
		specified := map[string]string{
			"PATH":   "${PATH}:${GOROOT}/bin:${GOPATH}/bin",
			"GOROOT": "${GOPATH}/usr/local/go",
			"GOPATH": "${GOROOT}/go",
		}
		expceted := []string{
			"PATH=/usr/bin:/bin:/bin",
			"GOROOT=/usr/local/go",
			"GOPATH=/go",
			"TEST=test",
		}
		env, err := modifyEnv(existed, []string{}, specified)
		assert.NoError(t, err)
		for _, e := range env {
			assert.Contains(t, expceted, e)
		}
	})
	t.Run("remove env", func(t *testing.T) {
		existed := []string{
			"GOROOT=/usr/local/go",
			"GOPATH=/go",
		}
		toRemove := []string{"GOROOT"}

		env, err := modifyEnv(existed, toRemove, nil)
		assert.NoError(t, err)
		expceted := []string{
			"GOPATH=/go",
		}
		assert.Equal(t, expceted, env)
	})
}

func TestModifyLabels(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		labels           map[string]string
		addLabels        map[string]string
		removeLabels     []string
		clearWerfLabels  bool
		expectedLabels   map[string]string
		expectedWarnings []string
	}{
		{
			name: "remove exact match",
			labels: map[string]string{
				"test-label": "bar",
				"werf":       "should-stay",
			},
			removeLabels:    []string{"test-label"},
			clearWerfLabels: false,
			expectedLabels: map[string]string{
				"werf": "should-stay",
			},
		},
		{
			name: "remove by regex",
			labels: map[string]string{
				"test-label": "bar",
				"test123":    "should-remove",
			},
			removeLabels:    []string{"/test[0-9]+/"},
			clearWerfLabels: false,
			expectedLabels: map[string]string{
				"test-label": "bar",
			},
		},
		{
			name: "clear werf labels",
			labels: map[string]string{
				"test-label":  "bar",
				"werf.test":   "should-remove",
				"werf.other":  "should-remove",
				"not-werf":    "keep",
				"werf-random": "should-remove",
			},
			clearWerfLabels: true,
			expectedLabels: map[string]string{
				"test-label": "bar",
				"not-werf":   "keep",
			},
		},
		{
			name: "add new labels",
			labels: map[string]string{
				"test-label": "bar",
			},
			addLabels: map[string]string{
				"new":   "value",
				"other": "123",
			},
			expectedLabels: map[string]string{
				"test-label": "bar",
				"new":        "value",
				"other":      "123",
			},
		},
		{
			name: "remove and add combined",
			labels: map[string]string{
				"test-label": "bar",
				"remove":     "this",
				"keep":       "me",
				"werf.me":    "delete",
			},
			removeLabels:    []string{"remove"},
			clearWerfLabels: true,
			addLabels: map[string]string{
				"new":                  "added",
				"project-%project%-id": "image-%image%-name",
			},
			expectedLabels: map[string]string{
				"test-label":              "bar",
				"keep":                    "me",
				"new":                     "added",
				"project-TEST-PROJECT-id": "image-TEST-IMAGE-name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labelsCopy := make(map[string]string)
			for k, v := range tt.labels {
				labelsCopy[k] = v
			}

			s := ImageSpecStage{
				BaseStage: &BaseStage{
					projectName: "TEST-PROJECT",
					imageName:   "TEST-IMAGE",
				},
			}

			modifiedLabels, err := s.modifyLabels(ctx, labelsCopy, tt.addLabels, tt.removeLabels, tt.clearWerfLabels)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLabels, modifiedLabels)
		})
	}
}

func TestModifyVolumes(t *testing.T) {
	tests := []struct {
		name           string
		initialVolumes map[string]struct{}
		removeVolumes  []string
		addVolumes     []string
		expected       map[string]struct{}
	}{
		{
			name: "add new volumes",
			initialVolumes: map[string]struct{}{
				"/data":  {},
				"/cache": {},
			},
			addVolumes: []string{"/logs", "/tmp"},
			expected: map[string]struct{}{
				"/data":  {},
				"/cache": {},
				"/logs":  {},
				"/tmp":   {},
			},
		},
		{
			name: "remove existing volumes",
			initialVolumes: map[string]struct{}{
				"/data":  {},
				"/cache": {},
				"/logs":  {},
			},
			removeVolumes: []string{"/cache", "/logs"},
			expected: map[string]struct{}{
				"/data": {},
			},
		},
		{
			name: "remove and add volumes",
			initialVolumes: map[string]struct{}{
				"/old": {},
			},
			removeVolumes: []string{"/old"},
			addVolumes:    []string{"/new"},
			expected: map[string]struct{}{
				"/new": {},
			},
		},
		{
			name:           "add volumes to empty map",
			initialVolumes: nil,
			addVolumes:     []string{"/mnt", "/backup"},
			expected: map[string]struct{}{
				"/mnt":    {},
				"/backup": {},
			},
		},
		{
			name:           "remove volumes from empty map",
			initialVolumes: nil,
			removeVolumes:  []string{"/mnt"},
			expected:       map[string]struct{}{},
		},
		{
			name: "removing non-existent volumes",
			initialVolumes: map[string]struct{}{
				"/data": {},
			},
			removeVolumes: []string{"/cache"},
			expected: map[string]struct{}{
				"/data": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifiedVolumes := modifyVolumes(tt.initialVolumes, tt.removeVolumes, tt.addVolumes)
			assert.Equal(t, tt.expected, modifiedVolumes)
		})
	}
}

func TestGetDependencies_StableHash(t *testing.T) {
	ctx := context.Background()

	imageSpec1 := &config.ImageSpec{
		Author:          "test-author",
		ClearHistory:    true,
		ClearWerfLabels: false,
		RemoveLabels:    []string{"label1", "label2"},
		RemoveVolumes:   []string{"/data", "/cache"},
		RemoveEnv:       []string{"ENV_VAR"},
		Volumes:         []string{"/var/log", "/opt/app"},
		Labels:          map[string]string{"app": "test", "version": "1.0"},
		Env:             map[string]string{"ENV_VAR": "value", "DEBUG": "true"},
		Expose:          []string{"80", "443"},
		User:            "root",
		Cmd:             []string{"sh", "-c", "echo Hello"},
		ClearCmd:        false,
		Entrypoint:      []string{"/entrypoint.sh"},
		ClearEntrypoint: false,
		WorkingDir:      "/app",
		StopSignal:      "SIGTERM",
	}

	imageSpec2 := &config.ImageSpec{
		Author:          "test-author",
		ClearHistory:    true,
		ClearWerfLabels: false,
		RemoveLabels:    []string{"label2", "label1"},
		RemoveVolumes:   []string{"/cache", "/data"},
		RemoveEnv:       []string{"ENV_VAR"},
		Volumes:         []string{"/opt/app", "/var/log"},
		Labels:          map[string]string{"version": "1.0", "app": "test"},
		Env:             map[string]string{"DEBUG": "true", "ENV_VAR": "value"},
		Expose:          []string{"443", "80"},
		User:            "root",
		Cmd:             []string{"sh", "-c", "echo Hello"},
		ClearCmd:        false,
		Entrypoint:      []string{"/entrypoint.sh"},
		ClearEntrypoint: false,
		WorkingDir:      "/app",
		StopSignal:      "SIGTERM",
	}

	stage1 := &ImageSpecStage{imageSpec: imageSpec1}
	stage2 := &ImageSpecStage{imageSpec: imageSpec2}

	hash1, err1 := stage1.GetDependencies(ctx, nil, nil, nil, nil, nil)
	hash2, err2 := stage2.GetDependencies(ctx, nil, nil, nil, nil, nil)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, hash1, hash2, "Hashes should be identical regardless of element order")
}
