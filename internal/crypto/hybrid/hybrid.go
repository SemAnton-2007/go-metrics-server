package hybrid

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"

	"go-metrics-server/internal/crypto/encryption"
	"go-metrics-server/internal/crypto/keys"
)

type Message struct {
	EncryptedKey []byte `json:"ek"`
	Data         []byte `json:"d"`
}

func Encrypt(data []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}

	encryptedData, err := encryption.Encrypt(data, aesKey)
	if err != nil {
		return nil, err
	}

	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, aesKey)
	if err != nil {
		return nil, err
	}

	msg := Message{
		EncryptedKey: encryptedKey,
		Data:         encryptedData,
	}

	return json.Marshal(msg)
}

func Decrypt(data []byte, privKey *rsa.PrivateKey) ([]byte, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, msg.EncryptedKey)
	if err != nil {
		return nil, err
	}

	return encryption.Decrypt(msg.Data, aesKey)
}

func NewEncryptor(publicKeyPath string) (*Encryptor, error) {
	pubKey, err := keys.LoadPublicKey(publicKeyPath)
	if err != nil {
		return nil, err
	}
	return &Encryptor{pubKey: pubKey}, nil
}

func NewDecryptor(privateKeyPath string) (*Decryptor, error) {
	privKey, err := keys.LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return &Decryptor{privKey: privKey}, nil
}

type Encryptor struct {
	pubKey *rsa.PublicKey
}

func (e *Encryptor) Encrypt(data []byte) ([]byte, error) {
	return Encrypt(data, e.pubKey)
}

type Decryptor struct {
	privKey *rsa.PrivateKey
}

func (d *Decryptor) Decrypt(data []byte) ([]byte, error) {
	return Decrypt(data, d.privKey)
}
