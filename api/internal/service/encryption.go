package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrEncryptionKeyNotSet = errors.New("encryption key not configured")
	ErrDecryptionFailed    = errors.New("decryption failed: invalid ciphertext")
)

// EncryptionService handles AES-256-GCM encryption for sensitive data.
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service with the given 32-byte key.
func NewEncryptionService(key []byte) *EncryptionService {
	return &EncryptionService{key: key}
}

// IsConfigured returns true if the encryption key is set.
func (s *EncryptionService) IsConfigured() bool {
	return len(s.key) == 32
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext.
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if !s.IsConfigured() {
		return "", ErrEncryptionKeyNotSet
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext and returns the plaintext.
func (s *EncryptionService) Decrypt(ciphertextB64 string) (string, error) {
	if !s.IsConfigured() {
		return "", ErrEncryptionKeyNotSet
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrDecryptionFailed
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}




