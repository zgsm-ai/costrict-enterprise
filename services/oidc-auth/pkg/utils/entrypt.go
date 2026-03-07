package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/zgsm-ai/oidc-auth/internal/config"
)

var (
	once              sync.Once
	encryptKeyManager *EncryptKeyManager
	globalConfig      *config.AppConfig
)

type EncryptKeyManager struct {
	privateKeyPEM string
	publicKeyPEM  string
	aesKey        string
	mu            sync.RWMutex
	Config        *config.EncryptConfig
}

func GetEncryptKeyManager() (*EncryptKeyManager, error) {
	var initErr error
	once.Do(func() {
		if globalConfig == nil {
			initErr = fmt.Errorf("global config not initialized")
			return
		}
		encryptKeyManager = &EncryptKeyManager{
			Config: &globalConfig.Encrypt,
		}
		initErr = encryptKeyManager.loadKeys()
	})
	if initErr != nil {
		return nil, initErr
	}
	return encryptKeyManager, nil
}

func SetGlobalConfig(cfg *config.AppConfig) {
	globalConfig = cfg
}

func (m *EncryptKeyManager) loadKeys() error {
	if !m.Config.EnableRsa {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	privateKeyBytes, err := os.ReadFile(m.Config.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %v", err)
	}
	m.privateKeyPEM = string(privateKeyBytes)

	publicKeyBytes, err := os.ReadFile(m.Config.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %v", err)
	}
	m.publicKeyPEM = string(publicKeyBytes)
	m.aesKey = m.Config.AesKey

	return nil
}

func (m *EncryptKeyManager) GetPrivateKeyPEM() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.privateKeyPEM
}

func (m *EncryptKeyManager) GetPublicKeyPEM() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.publicKeyPEM
}

func (m *EncryptKeyManager) GetAesKey() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.aesKey, nil
}

func (m *EncryptKeyManager) ReloadKeys() error {
	return m.loadKeys()
}

func (m *EncryptKeyManager) AESEncrypt(plaintext []byte) (string, error) {
	key := m.Config.AesKey
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (m *EncryptKeyManager) AESDecrypt(encryptedText string) ([]byte, error) {
	key := m.Config.AesKey
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
