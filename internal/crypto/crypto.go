package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

type EncryptedMessage struct {
	EncryptedKey []byte `json:"encrypted_key"`
	Data         []byte `json:"data"`
}

func Encrypt(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(data))
	stream.XORKeyStream(ciphertext, data)

	encryptedData := make([]byte, aes.BlockSize+len(ciphertext))
	copy(encryptedData[:aes.BlockSize], iv)
	copy(encryptedData[aes.BlockSize:], ciphertext)

	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, pub, key)
	if err != nil {
		return nil, err
	}

	msg := EncryptedMessage{
		EncryptedKey: encryptedKey,
		Data:         encryptedData,
	}

	return json.Marshal(msg)
}

func Decrypt(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	var msg EncryptedMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	key, err := rsa.DecryptPKCS1v15(rand.Reader, priv, msg.EncryptedKey)
	if err != nil {
		return nil, err
	}

	if len(msg.Data) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := msg.Data[:aes.BlockSize]
	ciphertext := msg.Data[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}

	return rsaPub, nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not RSA private key")
	}

	return rsaPriv, nil
}
