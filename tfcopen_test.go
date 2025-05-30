package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasKnownKeys(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{"workspace set", Config{Workspace: "ws"}, true},
		{"search set", Config{Search: "foo"}, true},
		{"project set", Config{Project: "bar"}, true},
		{"none set", Config{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := hasKnownKeys(&tt.cfg)
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestResolveOrg(t *testing.T) {
	t.Setenv("TFCOPEN_DEFAULT_ORG", "envorg")
	cfg := Config{}
	actual, err := resolveOrg(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "envorg", actual)

	cfg = Config{Org: "cfgorg"}
	actual, err = resolveOrg(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "cfgorg", actual)
}

func TestBuildURI(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"workspace", Config{Workspace: "ws"}, "/workspaces/ws"},
		{"search", Config{Search: "foo"}, "/workspaces?search=foo"},
		{"project", Config{Project: "bar"}, "/projects/bar"},
		{"none", Config{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildWorkspacesURI(&tt.cfg)
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestReadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".tfcopen")
	content := []byte("workspace: testws\norg: testorg\n")
	err := os.WriteFile(cfgPath, content, 0644)
	require.NoError(t, err, "failed to write temp config")

	cfg, err := ReadConfig(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "testws", cfg.Workspace)
	assert.Equal(t, "testorg", cfg.Org)
}

func TestReadConfig_FileNotFound(t *testing.T) {
	_, err := ReadConfig("nonexistent.yaml")
	assert.Error(t, err, "expected error for missing file")
}
