package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSite(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		cfg     *Config
		want    string
		wantErr bool
	}{
		{"default is tfc", "", nil, siteTFC, false},
		{"env sets terrakube", "terrakube", nil, siteTerrakube, false},
		{"cfg overrides env", "terrakube", &Config{Site: "tfc"}, siteTFC, false},
		{"cfg sets terrakube", "", &Config{Site: "Terrakube"}, siteTerrakube, false},
		{"unknown site", "", &Config{Site: "nope"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TFCOPEN_DEFAULT_SITE", tt.env)
			got, err := resolveSite(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewTerrakubeClientAddsAPIPath(t *testing.T) {
	assert.Equal(t, "https://api.example.com/api/v1", newTerrakubeClient("https://api.example.com/", "t").apiURL)
	assert.Equal(t, "https://api.example.com/api/v1", newTerrakubeClient("https://api.example.com/api/v1/", "t").apiURL)
}

// fakeTerrakube serves a tiny JSON:API that mimics the Terrakube endpoints tfcopen uses.
func fakeTerrakube(t *testing.T) *httptest.Server {
	type item struct{ id, name string }
	data := map[string][]item{
		"/api/v1/organization":                 {{"org-1", "myorg"}},
		"/api/v1/organization/org-1/workspace": {{"ws-1", "app-prod"}, {"ws-2", "app-dev"}, {"ws-3", "db-prod"}},
		"/api/v1/organization/org-1/project":   {{"prj-1", "platform"}},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))

		var filter string
		for k, v := range r.URL.Query() {
			if strings.HasPrefix(k, "filter[") {
				filter = v[0]
			}
		}
		pattern := strings.TrimPrefix(filter, "name==")

		out := []map[string]any{}
		for _, it := range data[r.URL.Path] {
			match := it.name == pattern
			if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
				match = strings.Contains(it.name, strings.Trim(pattern, "*"))
			}
			if match {
				out = append(out, map[string]any{"id": it.id, "attributes": map[string]string{"name": it.name}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}))
}

func TestBuildTerrakubeURL(t *testing.T) {
	srv := fakeTerrakube(t)
	defer srv.Close()
	client := newTerrakubeClient(srv.URL, "secret")
	ui := "https://ui.example.com/"
	base := "https://ui.example.com/organizations/org-1"

	tests := []struct {
		name     string
		cfg      *Config
		registry bool
		want     string
	}{
		{"workspace", &Config{Workspace: "app-prod"}, false, base + "/workspaces/ws-1"},
		{"project", &Config{Project: "platform"}, false, base + "/projects/prj-1"},
		{"search single match", &Config{Search: "db"}, false, base + "/workspaces/ws-3"},
		{"search many matches prints links, opens nothing", &Config{Search: "app"}, false, ""},
		{"search no matches", &Config{Search: "zzz"}, false, base + "/workspaces"},
		{"registry", nil, true, base + "/registry"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTerrakubeURL(ui, client, "myorg", tt.cfg, tt.registry)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildTerrakubeURLErrors(t *testing.T) {
	srv := fakeTerrakube(t)
	defer srv.Close()
	client := newTerrakubeClient(srv.URL, "secret")

	_, err := buildTerrakubeURL("", client, "myorg", &Config{}, false)
	assert.ErrorContains(t, err, "TFCOPEN_TERRAKUBE_URL")

	_, err = buildTerrakubeURL("https://ui", client, "missing", &Config{}, false)
	assert.ErrorContains(t, err, "no terrakube organization")

	_, err = buildTerrakubeURL("https://ui", client, "myorg", &Config{Workspace: "missing"}, false)
	assert.ErrorContains(t, err, "no terrakube workspace")
}
