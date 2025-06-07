package config

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type HTTPConfig struct {
	Port int
}

type NATSConfig struct {
	URL    string
	Token  string
	Stream string
}

type AppConfig struct {
	ChunkSize int
}

type VaultConfig struct {
	Address string
	Token   string
}

type Config struct {
	HTTP  HTTPConfig
	NATS  NATSConfig
	App   AppConfig
	Vault VaultConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load("services/uploader/.env")

	viper.SetEnvPrefix("VIDLOCK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{
		HTTP: HTTPConfig{
			Port: viper.GetInt("HTTP.PORT"),
		},
		NATS: NATSConfig{
			URL:    viper.GetString("NATS.URL"),
			Stream: viper.GetString("NATS.STREAM"),
		},
		App: AppConfig{
			ChunkSize: viper.GetInt("APP.CHUNK_SIZE"),
		},
		Vault: VaultConfig{
			Address: viper.GetString("VAULT.ADDRESS"),
			Token:   viper.GetString("VAULT.TOKEN"),
		},
	}

	if err := loadVaultSecrets(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadVaultSecrets(cfg *Config) error {
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
