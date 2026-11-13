package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeJWT(exp time.Time) string {
	payload := base64.RawURLEncoding.EncodeToString(fmt.Appendf(nil, `{"exp":%d}`, exp.Unix()))
	return "header." + payload + ".sig"
}

func writeCreds(t *testing.T, path, host, token string) {
	t.Helper()
	content := fmt.Sprintf(`{"credentials":{%q:{"token":%q}}}`, host, token)
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func TestTokenExpired(t *testing.T) {
	now := time.Now()
	assert.False(t, tokenExpired(fakeJWT(now.Add(time.Hour)), now))
	assert.True(t, tokenExpired(fakeJWT(now.Add(-time.Hour)), now))
	assert.False(t, tokenExpired("not-a-jwt", now))
}

func TestApiHost(t *testing.T) {
	host, err := apiHost("https://terrakube-api.example.com/api/v1")
	require.NoError(t, err)
	assert.Equal(t, "terrakube-api.example.com", host)

	_, err = apiHost("nonsense")
	assert.Error(t, err)
}

func TestTokenFromCredentials(t *testing.T) {
	host := "terrakube-api.example.com"
	now := time.Now()

	t.Run("valid token, no login", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "creds.json")
		good := fakeJWT(now.Add(time.Hour))
		writeCreds(t, path, host, good)

		token, err := tokenFromCredentials(path, host, now, func(string) error {
			t.Fatal("login should not be called")
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, good, token)
	})

	t.Run("expired token triggers login", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "creds.json")
		writeCreds(t, path, host, fakeJWT(now.Add(-time.Hour)))
		fresh := fakeJWT(now.Add(time.Hour))

		token, err := tokenFromCredentials(path, host, now, func(h string) error {
			assert.Equal(t, host, h)
			writeCreds(t, path, host, fresh)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, fresh, token)
	})

	t.Run("missing file triggers login", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "creds.json")
		fresh := fakeJWT(now.Add(time.Hour))

		token, err := tokenFromCredentials(path, host, now, func(string) error {
			writeCreds(t, path, host, fresh)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, fresh, token)
	})

	t.Run("login fails", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "creds.json")
		_, err := tokenFromCredentials(path, host, now, func(string) error {
			return fmt.Errorf("cancelled")
		})
		assert.ErrorContains(t, err, "terraform login failed")
	})
}
