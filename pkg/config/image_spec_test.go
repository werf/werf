package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeSlices(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "No duplicates",
			a:        []string{"a", "b"},
			b:        []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "With duplicates",
			a:        []string{"a", "b"},
			b:        []string{"b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Empty first slice",
			a:        []string{},
			b:        []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "Empty second slice",
			a:        []string{"a", "b"},
			b:        []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "Both empty",
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeSlices(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeImageSpec(t *testing.T) {
	meta := &ImageSpec{
		Author:                  "meta_author",
		ClearHistory:            true,
		KeepEssentialWerfLabels: true,
		RemoveLabels:            []string{"label1", "label2"},
		Labels:                  map[string]string{"key1": "value1"},
	}

	image := &ImageSpec{
		Author:                  "image_author",
		ClearHistory:            false,
		KeepEssentialWerfLabels: false,
		ClearWerfLabels:         true,
		RemoveLabels:            []string{"label3"},
		RemoveVolumes:           []string{"volume1"},
		RemoveEnv:               []string{"ENV1"},
		ClearCmd:                true,
		ClearEntrypoint:         true,
		ClearUser:               true,
		ClearWorkingDir:         true,
		Volumes:                 []string{"volume_old"},
		Labels:                  map[string]string{"key2": "value2"},
		Env:                     map[string]string{"ENV_OLD": "VALUE_OLD"},
		User:                    "image_user",
		Cmd:                     []string{"cmd2"},
		Entrypoint:              []string{"entry2"},
		WorkingDir:              "/image_workdir",
		StopSignal:              "SIGTERM",
		Expose:                  []string{"9090"},
	}

	expected := ImageSpec{
		Author:                  "image_author",
		ClearHistory:            true,
		KeepEssentialWerfLabels: true,
		ClearWerfLabels:         true,
		RemoveLabels:            []string{"label3", "label1", "label2"},
		RemoveVolumes:           []string{"volume1"},
		RemoveEnv:               []string{"ENV1"},
		ClearCmd:                true,
		ClearEntrypoint:         true,
		ClearUser:               true,
		ClearWorkingDir:         true,
		Volumes:                 []string{"volume_old"},
		Labels:                  map[string]string{"key1": "value1", "key2": "value2"},
		Env:                     map[string]string{"ENV_OLD": "VALUE_OLD"},
		User:                    "image_user",
		Cmd:                     []string{"cmd2"},
		Entrypoint:              []string{"entry2"},
		WorkingDir:              "/image_workdir",
		StopSignal:              "SIGTERM",
		Expose:                  []string{"9090"},
	}

	result := mergeImageSpec(meta, image)

	assert.Equal(t, expected, result)
}

func TestRawImageSpecToDirective(t *testing.T) {
	raw := &rawImageSpec{
		Author:       "raw_author",
		ClearHistory: true,
		Config: &rawImageSpecConfig{
			KeepEssentialWerfLabels: true,
			ClearWerfLabels:         true,
			ClearCmd:                true,
			ClearEntrypoint:         true,
			ClearUser:               true,
			ClearWorkingDir:         true,
			RemoveLabels:            []string{"label1", "label2"},
			RemoveVolumes:           []string{"vol1", "vol2"},
			RemoveEnv:               []string{"ENV1"},
			Volumes:                 []string{"vol3"},
			Labels:                  map[string]string{"key1": "value1"},
			Env:                     map[string]string{"ENV_NEW": "VALUE_NEW"},
			User:                    "raw_user",
			Cmd:                     []string{"cmd1"},
			Entrypoint:              []string{"entry1"},
			WorkingDir:              "/raw_workdir",
			StopSignal:              "SIGKILL",
			Expose:                  []string{"8080"},
		},
	}

	expected := &ImageSpec{
		Author:                  "raw_author",
		ClearHistory:            true,
		KeepEssentialWerfLabels: true,
		ClearWerfLabels:         true,
		ClearCmd:                true,
		ClearEntrypoint:         true,
		ClearUser:               true,
		ClearWorkingDir:         true,
		RemoveLabels:            []string{"label1", "label2"},
		RemoveVolumes:           []string{"vol1", "vol2"},
		RemoveEnv:               []string{"ENV1"},
		Volumes:                 []string{"vol3"},
		Labels:                  map[string]string{"key1": "value1"},
		Env:                     map[string]string{"ENV_NEW": "VALUE_NEW"},
		User:                    "raw_user",
		Cmd:                     []string{"cmd1"},
		Entrypoint:              []string{"entry1"},
		WorkingDir:              "/raw_workdir",
		StopSignal:              "SIGKILL",
		Expose:                  []string{"8080"},
		raw:                     raw,
	}

	result := raw.toDirective()

	assert.Equal(t, expected, result)
}

func TestRawImageSpecConfigFieldsMapped(t *testing.T) {
	rawType := reflect.TypeOf(rawImageSpecConfig{})
	imageType := reflect.TypeOf(ImageSpec{})

	testToDirectiveStructs(t, rawType, imageType)
}

func testToDirectiveStructs(t *testing.T, rawType, directiveType reflect.Type) {
	mappedFields := map[string]struct{}{}

	for i := 0; i < directiveType.NumField(); i++ {
		mappedFields[directiveType.Field(i).Name] = struct{}{}
	}

	for i := 0; i < rawType.NumField(); i++ {
		fieldName := rawType.Field(i).Name
		if strings.HasPrefix(fieldName, "raw") || fieldName == "UnsupportedAttributes" {
			continue
		}
		_, found := mappedFields[fieldName]
		assert.True(t, found, "no field %s from raw in directive", fieldName)
	}
}
