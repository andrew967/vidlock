package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ChunkEncryptor struct{}

func NewChunkEncryptor() *ChunkEncryptor {
	return &ChunkEncryptor{}
}

func (e *ChunkEncryptor) Encrypt(inputPath string) (string, []byte, error) {

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", nil, fmt.Errorf("failed to generate key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil, fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil, fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", nil, fmt.Errorf("generate nonce: %w", err)
	}

	plainData, err := os.ReadFile(inputPath)
	if err != nil {
		return "", nil, fmt.Errorf("read input: %w", err)
	}

	cipherData := gcm.Seal(nonce, nonce, plainData, nil)

	encPath := tempEncryptedPath(inputPath)
	if err := os.WriteFile(encPath, cipherData, 0600); err != nil {
		return "", nil, fmt.Errorf("write encrypted file: %w", err)
	}

	return encPath, key, nil
}

func tempEncryptedPath(input string) string {
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	timestamp := time.Now().UnixNano()
	return filepath.Join("/tmp", fmt.Sprintf("%s_encrypted_%d.enc", name, timestamp))
}
