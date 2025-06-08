package config

import (
	"fmt"
	"log"
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

type DBConfig struct {
	DSN string
}

type VaultConfig struct {
	Address string
	Token   string
}

type JWTConfig struct {
	SecretKey string
}

type Config struct {
	HTTP  HTTPConfig
	NATS  NATSConfig
	DB    DBConfig
	Vault VaultConfig
	JWT   JWTConfig
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	_ = godotenv.Load()

	viper.SetEnvPrefix("META")
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
		DB: DBConfig{
			DSN: "",
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

	if secret, err := client.Logical().Read("secret/data/vidlock"); err == nil && secret != nil {
		if data, ok := secret.Data["data"].(map[string]interface{}); ok {
			if token, ok := data["nats_token"].(string); ok {
				cfg.NATS.Token = token
			}
		}
	}

	if secret, err := client.Logical().Read("secret/data/auth-service"); err == nil && secret != nil {
		if data, ok := secret.Data["data"].(map[string]interface{}); ok {
			cfg.JWT.SecretKey, _ = data["JWT_SECRET"].(string)
		}
	}

	if secret, err := client.Logical().Read("secret/data/metadata-service"); err == nil && secret != nil {
		if data, ok := secret.Data["data"].(map[string]interface{}); ok {
			cfg.DB.DSN, _ = data["db_dsn"].(string)
		}
	}
	log.Println("DSN from Vault:", cfg.DB.DSN)

	return nil
}
