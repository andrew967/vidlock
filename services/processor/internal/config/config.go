package config

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
)

type VaultConfig struct {
	Address string
	Token   string
}

type NATSConfig struct {
	URL    string
	Token  string
	Stream string
}

type IPFSConfig struct {
	APIAddress string
}

type Config struct {
	Vault VaultConfig
	NATS  NATSConfig
	IPFS  IPFSConfig
}

func Load() *Config {
	_ = godotenv.Load("/app/.env")

	cfg := &Config{
		Vault: VaultConfig{
			Address: getEnv("VAULT_ADDR", "http://localhost:8200"),
			Token:   getEnv("VAULT_TOKEN", "root"),
		},
		NATS: NATSConfig{
			URL:    getEnv("NATS_URL", "nats://localhost:4222"),
			Stream: getEnv("NATS_STREAM", "VIDEO_UPLOADS"),
		},
		IPFS: IPFSConfig{
			APIAddress: getEnv("IPFS_API", "localhost:5001"),
		},
	}

	return cfg
}

func LoadVaultSecrets(cfg *Config) error {
	client, err := api.NewClient(&api.Config{Address: cfg.Vault.Address})
	if err != nil {
		return fmt.Errorf("vault client error: %w", err)
	}
	client.SetToken(cfg.Vault.Token)

	secret, err := client.Logical().Read("secret/data/vidlock")
	if err != nil {
		return fmt.Errorf("vault read error: %w", err)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid secret data structure")
	}

	if token, ok := data["nats_token"].(string); ok {
		cfg.NATS.Token = token
	}

	return nil
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
