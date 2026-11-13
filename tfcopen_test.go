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
		{"project", Config{Project: "bar"}, "/projects/bar/workspaces"},
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

func TestWithGlobalDefaults(t *testing.T) {
	global := &Config{Org: "globalorg", Site: "terrakube", TerrakubeURL: "https://ui", TerrakubeAPIURL: "https://api", Workspace: "ignored"}

	merged := withGlobalDefaults(&Config{Workspace: "ws", Site: "tfc"}, global)
	assert.Equal(t, "ws", merged.Workspace)
	assert.Equal(t, "tfc", merged.Site, "repo config wins")
	assert.Equal(t, "globalorg", merged.Org)
	assert.Equal(t, "https://ui", merged.TerrakubeURL)
	assert.Equal(t, "https://api", merged.TerrakubeAPIURL)

	merged = withGlobalDefaults(nil, global)
	assert.Equal(t, "globalorg", merged.Org)
	assert.Empty(t, merged.Workspace, "workspace keys never come from the global config")
}

func TestLoadGlobalConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := loadGlobalConfig()
	require.NoError(t, err, "missing file is fine")
	assert.Equal(t, Config{}, *cfg)

	content := []byte("org: myorg\nsite: terrakube\nterrakube_url: https://ui\nterrakube_api_url: https://api\n")
	require.NoError(t, os.WriteFile(filepath.Join(home, ".tfcopen"), content, 0644))

	cfg, err = loadGlobalConfig()
	require.NoError(t, err)
	assert.Equal(t, "myorg", cfg.Org)
	assert.Equal(t, "terrakube", cfg.Site)
	assert.Equal(t, "https://ui", cfg.TerrakubeURL)
	assert.Equal(t, "https://api", cfg.TerrakubeAPIURL)
}

func TestFindConfigSkipsGlobalFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".tfcopen"), []byte("org: myorg\n"), 0644))

	repo := filepath.Join(home, "myrepo")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0755))
	t.Chdir(repo)

	cfg, err := findConfig()
	require.NoError(t, err)
	assert.Equal(t, "myrepo", cfg.Search)

	// a folder under home with no repo and no .tfcopen must not pick up ~/.tfcopen as a workspace config
	loose := filepath.Join(home, "loose")
	require.NoError(t, os.MkdirAll(loose, 0755))
	t.Chdir(loose)
	_, err = findConfig()
	assert.ErrorContains(t, err, "reached /")
}

func TestFindConfigSettingsOnlySearchesByFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	repo := filepath.Join(home, "s3-buckets")
	require.NoError(t, os.MkdirAll(repo, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".tfcopen"), []byte("site: tfc\n"), 0644))
	t.Chdir(repo)

	cfg, err := findConfig()
	require.NoError(t, err)
	assert.Equal(t, "s3-buckets", cfg.Search)
	assert.Equal(t, "tfc", cfg.Site)
}
