package encryption_test

import (
	"crypto/rand"
	"testing"

	enc "go-metrics-server/internal/crypto/encryption"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	data := []byte("secret message")

	encrypted, err := enc.Encrypt(data, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := enc.Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(data) {
		t.Errorf("Decrypted data mismatch: got %s, want %s", decrypted, data)
	}
}

func TestInvalidKey(t *testing.T) {
	_, err := enc.Encrypt([]byte("test"), []byte("short-key"))
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
}

func TestCorruptedData(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	_, err := enc.Decrypt([]byte("too-short"), key)
	if err == nil {
		t.Error("Expected error for corrupted data")
	}
}
