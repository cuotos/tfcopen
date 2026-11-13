package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	siteTFC       = "tfc"
	siteTerrakube = "terrakube"
)

// resolveSite picks the site from the .tfcopen file, then TFCOPEN_DEFAULT_SITE, then defaults to tfc.
func resolveSite(cfg *Config) (string, error) {
	site := ""
	if cfg != nil && cfg.Site != "" {
		site = cfg.Site
	} else {
		site = os.Getenv("TFCOPEN_DEFAULT_SITE")
	}

	site = strings.ToLower(strings.TrimSpace(site))
	switch site {
	case "", siteTFC:
		return siteTFC, nil
	case siteTerrakube:
		return siteTerrakube, nil
	default:
		return "", fmt.Errorf("unknown site %q. expected %q or %q", site, siteTFC, siteTerrakube)
	}
}

// terrakubeClient looks up Terrakube IDs from names using the Terrakube API.
type terrakubeClient struct {
	apiURL string // e.g. https://terrakube-api.example.com/api/v1
	token  string
	http   *http.Client
}

func newTerrakubeClientFor(apiURL string) (*terrakubeClient, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("terrakube_api_url is not set in ~/.tfcopen (or TFCOPEN_TERRAKUBE_API_URL). it is needed to look up terrakube ids")
	}
	token, err := terrakubeToken(apiURL)
	if err != nil {
		return nil, err
	}
	return newTerrakubeClient(apiURL, token), nil
}

func newTerrakubeClient(apiURL, token string) *terrakubeClient {
	apiURL = strings.TrimRight(apiURL, "/")
	if !strings.HasSuffix(apiURL, "/api/v1") {
		apiURL += "/api/v1"
	}
	return &terrakubeClient{
		apiURL: apiURL,
		token:  token,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

type jsonAPIItem struct {
	ID         string `json:"id"`
	Attributes struct {
		Name string `json:"name"`
	} `json:"attributes"`
}

// list GETs a JSON:API collection, filtered by an RSQL expression on the given type.
func (c *terrakubeClient) list(path, filterType, rsql string) ([]jsonAPIItem, error) {
	q := url.Values{}
	q.Set(fmt.Sprintf("filter[%s]", filterType), rsql)
	reqURL := fmt.Sprintf("%s/%s?%s", c.apiURL, path, q.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling terrakube api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("terrakube api returned %s for %s", resp.Status, reqURL)
	}

	var body struct {
		Data []jsonAPIItem `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding terrakube api response: %w", err)
	}
	return body.Data, nil
}

// findOne returns the ID of the single item with exactly this name.
func (c *terrakubeClient) findOne(path, filterType, kind, name string) (string, error) {
	items, err := c.list(path, filterType, "name=="+name)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no terrakube %s named %q found", kind, name)
	}
	return items[0].ID, nil
}

func (c *terrakubeClient) orgID(name string) (string, error) {
	return c.findOne("organization", "organization", "organization", name)
}

func (c *terrakubeClient) workspaceID(orgID, name string) (string, error) {
	return c.findOne("organization/"+orgID+"/workspace", "workspace", "workspace", name)
}

func (c *terrakubeClient) projectID(orgID, name string) (string, error) {
	return c.findOne("organization/"+orgID+"/project", "project", "project", name)
}

func (c *terrakubeClient) searchWorkspaces(orgID, search string) ([]jsonAPIItem, error) {
	return c.list("organization/"+orgID+"/workspace", "workspace", "name==*"+search+"*")
}

// buildTerrakubeURL turns the config into a Terrakube UI URL, looking up IDs via the API.
func buildTerrakubeURL(uiURL string, client *terrakubeClient, org string, cfg *Config, registry bool) (string, error) {
	if uiURL == "" {
		return "", fmt.Errorf("terrakube_url is not set in ~/.tfcopen (or TFCOPEN_TERRAKUBE_URL). it is needed to build terrakube links")
	}
	uiURL = strings.TrimRight(uiURL, "/")

	orgID, err := client.orgID(org)
	if err != nil {
		return "", err
	}
	orgBase := fmt.Sprintf("%s/organizations/%s", uiURL, orgID)
	workspacesList := orgBase + "/workspaces"

	if registry {
		return orgBase + "/registry", nil
	}

	switch {
	case cfg.Workspace != "":
		id, err := client.workspaceID(orgID, cfg.Workspace)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s/workspaces/%s", orgBase, id), nil

	case cfg.Search != "":
		// Terrakube's workspace list page has no search URL param, so search via the API instead.
		matches, err := client.searchWorkspaces(orgID, cfg.Search)
		if err != nil {
			return "", err
		}
		switch len(matches) {
		case 0:
			fmt.Fprintf(os.Stderr, "no terrakube workspaces match %q, opening the workspace list\n", cfg.Search)
			return workspacesList, nil
		case 1:
			return fmt.Sprintf("%s/workspaces/%s", orgBase, matches[0].ID), nil
		default:
			// several matches: print a clickable link for each and open nothing, so the user picks one
			fmt.Printf("%d terrakube workspaces match %q:\n", len(matches), cfg.Search)
			width := 0
			for _, m := range matches {
				width = max(width, len(m.Attributes.Name))
			}
			for _, m := range matches {
				fmt.Printf("  %-*s  %s/workspaces/%s\n", width, m.Attributes.Name, orgBase, m.ID)
			}
			return "", nil
		}

	case cfg.Project != "":
		id, err := client.projectID(orgID, cfg.Project)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s/projects/%s", orgBase, id), nil

	default:
		return workspacesList, nil
	}
}
