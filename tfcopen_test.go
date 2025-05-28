package main

import (
	"os"
	"path/filepath"
	"testing"
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
			got := hasKnownKeys(&tt.cfg)
			if got != tt.want {
				t.Errorf("hasKnownKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveOrg(t *testing.T) {
	// Save and restore env
	oldEnv := os.Getenv("TFCOPEN_DEFAULT_ORG")
	defer os.Setenv("TFCOPEN_DEFAULT_ORG", oldEnv)

	os.Setenv("TFCOPEN_DEFAULT_ORG", "envorg")
	cfg := Config{}
	if got := resolveOrg(&cfg); got != "envorg" {
		t.Errorf("resolveOrg() = %v, want envorg", got)
	}

	cfg = Config{Org: "cfgorg"}
	if got := resolveOrg(&cfg); got != "cfgorg" {
		t.Errorf("resolveOrg() = %v, want cfgorg", got)
	}
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
			got := buildURI(&tt.cfg)
			if got != tt.want {
				t.Errorf("buildURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".tfcopen")
	content := []byte("workspace: testws\norg: testorg\n")
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := ReadConfig(cfgPath)
	if err != nil {
		t.Fatalf("ReadConfig() error = %v", err)
	}
	if cfg.Workspace != "testws" || cfg.Org != "testorg" {
		t.Errorf("ReadConfig() = %+v, want workspace=testws, org=testorg", cfg)
	}
}

func TestReadConfig_FileNotFound(t *testing.T) {
	_, err := ReadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
