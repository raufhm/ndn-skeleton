package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Manager struct {
	mu      sync.RWMutex
	secrets *Secrets
}

type Secrets struct {
	JWTSecret     string `json:"jwt_secret"`
	DatabaseURL   string `json:"database_url"`
	AdminAPIKey   string `json:"admin_api_key"`
	StorageKey    string `json:"storage_key"`
	EncryptionKey string `json:"encryption_key"`
}

var (
	instance *Manager
	once     sync.Once
)

// GetManager returns a singleton instance of the secrets manager
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{}
	})
	return instance
}

// LoadSecrets loads secrets from the encrypted secrets file
func (m *Manager) LoadSecrets() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get environment-specific secrets file path
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	secretsPath := filepath.Join("config", "secrets."+env+".json")
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	var secrets Secrets
	if err := json.Unmarshal(data, &secrets); err != nil {
		return fmt.Errorf("failed to parse secrets: %w", err)
	}

	// Override with environment variables if present
	if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		secrets.DatabaseURL = envURL
	}
	if envJWT := os.Getenv("JWT_SECRET"); envJWT != "" {
		secrets.JWTSecret = envJWT
	}
	if envAdmin := os.Getenv("ADMIN_API_KEY"); envAdmin != "" {
		secrets.AdminAPIKey = envAdmin
	}
	if envStorage := os.Getenv("STORAGE_KEY"); envStorage != "" {
		secrets.StorageKey = envStorage
	}
	if envEncryption := os.Getenv("ENCRYPTION_KEY"); envEncryption != "" {
		secrets.EncryptionKey = envEncryption
	}

	m.secrets = &secrets
	return nil
}

// GetSecrets returns the current secrets
func (m *Manager) GetSecrets() *Secrets {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.secrets
}

// UpdateSecrets updates the secrets file
func (m *Manager) UpdateSecrets(secrets *Secrets) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	secretsPath := filepath.Join("config", "secrets."+env+".json")
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	if err := os.WriteFile(secretsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	m.secrets = secrets
	return nil
}
