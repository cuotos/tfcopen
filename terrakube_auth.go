package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// terrakubeToken finds a token for the Terrakube API.
// Order: TFCOPEN_TERRAKUBE_TOKEN, then the `terraform login` credentials file.
// If the saved token is missing or expired, it runs `terraform login` and reads the file again.
func terrakubeToken(apiURL string) (string, error) {
	if token := os.Getenv("TFCOPEN_TERRAKUBE_TOKEN"); token != "" {
		return token, nil
	}

	host, err := apiHost(apiURL)
	if err != nil {
		return "", err
	}

	credsPath, err := terraformCredentialsPath()
	if err != nil {
		return "", err
	}

	return tokenFromCredentials(credsPath, host, time.Now(), runTerraformLogin)
}

// tokenFromCredentials reads the token for host, calling login once if it is missing or expired.
func tokenFromCredentials(credsPath, host string, now time.Time, login func(host string) error) (string, error) {
	token, _ := readCredentialsToken(credsPath, host)
	if token != "" && !tokenExpired(token, now) {
		return token, nil
	}

	if token == "" {
		fmt.Fprintf(os.Stderr, "no saved terrakube token for %s, running terraform login\n", host)
	} else {
		fmt.Fprintf(os.Stderr, "terrakube token for %s has expired, running terraform login\n", host)
	}

	if err := login(host); err != nil {
		return "", fmt.Errorf("terraform login failed: %w", err)
	}

	token, err := readCredentialsToken(credsPath, host)
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("terraform login finished but no token for %s was saved in %s", host, credsPath)
	}
	return token, nil
}

func apiHost(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("TFCOPEN_TERRAKUBE_API_URL %q is not a valid url", apiURL)
	}
	return u.Host, nil
}

func terraformCredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, ".terraform.d", "credentials.tfrc.json"), nil
}

func readCredentialsToken(path, host string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	var creds struct {
		Credentials map[string]struct {
			Token string `json:"token"`
		} `json:"credentials"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}
	return creds.Credentials[host].Token, nil
}

// tokenExpired reports whether a JWT's exp claim is in the past (with a minute's margin).
// Tokens that aren't JWTs, or have no exp, are treated as not expired.
func tokenExpired(token string, now time.Time) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return false
	}
	return now.Add(time.Minute).After(time.Unix(claims.Exp, 0))
}

// runTerraformLogin runs `terraform login` interactively.
// Its output goes to stderr so `tfcopen -p` still prints only the url on stdout.
func runTerraformLogin(host string) error {
	cmd := exec.Command("terraform", "login", host)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
