package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const credentialCipherPrefix = "enc:v1:"

// CredentialCipher encrypts and decrypts sensitive credential values.
// Ciphertext format: enc:v1:<base64(nonce|ciphertext)>
type CredentialCipher struct {
	aead cipher.AEAD
}

func NewCredentialCipher(key string) (*CredentialCipher, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return nil, fmt.Errorf("credential encryption key is required")
	}

	raw, err := decodeCipherKey(trimmed)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES block cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-GCM cipher: %w", err)
	}

	return &CredentialCipher{aead: aead}, nil
}

func decodeCipherKey(key string) ([]byte, error) {
	if decoded, err := base64.StdEncoding.DecodeString(key); err == nil {
		switch len(decoded) {
		case 32:
			return decoded, nil
		}
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("credential encryption key must be 32 bytes raw or base64-encoded 32 bytes")
	}

	return []byte(key), nil
}

func (c *CredentialCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	sealed := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, sealed...)

	return credentialCipherPrefix + base64.StdEncoding.EncodeToString(payload), nil
}

// Decrypt decrypts encrypted values. For backward compatibility, plaintext values
// without the encryption prefix are returned as-is.
func (c *CredentialCipher) Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	if !strings.HasPrefix(value, credentialCipherPrefix) {
		return value, nil
	}

	b64 := strings.TrimPrefix(value, credentialCipherPrefix)
	payload, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted payload encoding: %w", err)
	}

	nonceSize := c.aead.NonceSize()
	if len(payload) <= nonceSize {
		return "", fmt.Errorf("invalid encrypted payload size")
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plain, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	return string(plain), nil
}
