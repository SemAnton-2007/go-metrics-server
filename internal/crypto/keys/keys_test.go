package keys_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"go-metrics-server/internal/crypto/keys"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadKeys(t *testing.T) {
	tmpDir := t.TempDir()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("load valid public key", func(t *testing.T) {
		pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)

		pubFile := filepath.Join(tmpDir, "pub.pem")
		require.NoError(t, os.WriteFile(pubFile, pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubBytes,
		}), 0644))

		pub, err := keys.LoadPublicKey(pubFile)
		assert.NoError(t, err)
		assert.IsType(t, &rsa.PublicKey{}, pub)
	})

	t.Run("load valid private key", func(t *testing.T) {
		privFile := filepath.Join(tmpDir, "priv.pem")
		require.NoError(t, os.WriteFile(privFile, pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}), 0644))

		priv, err := keys.LoadPrivateKey(privFile)
		assert.NoError(t, err)
		assert.IsType(t, &rsa.PrivateKey{}, priv)
	})

	t.Run("invalid key file", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.pem")
		require.NoError(t, os.WriteFile(invalidFile, []byte("invalid data"), 0644))

		_, err := keys.LoadPublicKey(invalidFile)
		assert.Error(t, err)

		_, err = keys.LoadPrivateKey(invalidFile)
		assert.Error(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := keys.LoadPublicKey("nonexistent.pem")
		assert.Error(t, err)

		_, err = keys.LoadPrivateKey("nonexistent.pem")
		assert.Error(t, err)
	})
}
