package hybrid_test

import (
	"path/filepath"
	"testing"

	"go-metrics-server/internal/crypto/hybrid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	pubKeyPath := filepath.Join("testdata", "public.pem")
	privKeyPath := filepath.Join("testdata", "private.pem")

	encryptor, err := hybrid.NewEncryptor(pubKeyPath)
	require.NoError(t, err)
	require.NotNil(t, encryptor)

	decryptor, err := hybrid.NewDecryptor(privKeyPath)
	require.NoError(t, err)
	require.NotNil(t, decryptor)

	testData := []byte("secret message")

	encrypted, err := encryptor.Encrypt(testData)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := decryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}

func TestNewEncryptorWithInvalidKey(t *testing.T) {
	_, err := hybrid.NewEncryptor("invalid_path.pem")
	require.Error(t, err)
}

func TestNewDecryptorWithInvalidKey(t *testing.T) {
	_, err := hybrid.NewDecryptor("invalid_path.pem")
	require.Error(t, err)
}
