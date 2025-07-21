package git_repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthCredentialsHelper(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		auth, err := BasicAuthCredentialsHelper(nil)
		require.NoError(t, err)
		assert.Nil(t, auth)
	})

	t.Run("multiple password sources", func(t *testing.T) {
		cfg := &BasicAuthCredentials{
			Username: "user",
			Password: PasswordSource{
				Env:        "ENV_VAR",
				Src:        "/tmp/secret",
				PlainValue: "secret",
			},
		}
		_, err := BasicAuthCredentialsHelper(cfg)
		assert.Error(t, err)
	})

	t.Run("env var not set", func(t *testing.T) {
		cfg := &BasicAuthCredentials{
			Username: "user",
			Password: PasswordSource{
				Env: "NON_EXISTENT_ENV_VAR",
			},
		}
		auth, err := BasicAuthCredentialsHelper(cfg)
		require.Error(t, err)
		assert.Nil(t, auth)
		assert.Equal(t, `environment variable "NON_EXISTENT_ENV_VAR" is not set`, err.Error())
	})

	t.Run("env var success", func(t *testing.T) {
		os.Setenv("MY_SECRET_ENV", "envpass")
		defer os.Unsetenv("MY_SECRET_ENV")

		cfg := &BasicAuthCredentials{
			Username: "user",
			Password: PasswordSource{
				Env: "MY_SECRET_ENV",
			},
		}
		auth, err := BasicAuthCredentialsHelper(cfg)
		require.NoError(t, err)
		require.NotNil(t, auth)
		assert.Equal(t, "user", auth.Username)
		assert.Equal(t, "envpass", auth.Password)
	})

	t.Run("file source success", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "secret.txt")
		secret := "filepass"
		err := os.WriteFile(tmpFile, []byte(secret), 0600)
		require.NoError(t, err)
		defer os.Remove(tmpFile)

		cfg := &BasicAuthCredentials{
			Username: "user",
			Password: PasswordSource{
				Src: tmpFile,
			},
		}
		auth, err := BasicAuthCredentialsHelper(cfg)
		require.NoError(t, err)
		require.NotNil(t, auth)
		assert.Equal(t, "user", auth.Username)
		assert.Equal(t, secret, auth.Password)
	})

	t.Run("file not found", func(t *testing.T) {
		cfg := &BasicAuthCredentials{
			Username: "user",
			Password: PasswordSource{
				Src: "/nonexistent/path.txt",
			},
		}
		auth, err := BasicAuthCredentialsHelper(cfg)
		require.Error(t, err)
		assert.Nil(t, auth)
		assert.Contains(t, err.Error(), "unable to read secret file")
	})

	t.Run("plain value success", func(t *testing.T) {
		cfg := &BasicAuthCredentials{
			Username: "user",
			Password: PasswordSource{
				PlainValue: "plainpass",
			},
		}
		auth, err := BasicAuthCredentialsHelper(cfg)
		require.NoError(t, err)
		require.NotNil(t, auth)
		assert.Equal(t, "user", auth.Username)
		assert.Equal(t, "plainpass", auth.Password)
	})
}
