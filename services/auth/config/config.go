package config

import (
	"strconv"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	JWT      JWTConfig
	Vault    VaultConfig
	Redis    RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
}

type HTTPConfig struct {
	Port int
}

type PostgresConfig struct {
	DSN string
}

type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type VaultConfig struct {
	Address string
	Token   string
	Path    string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	cfg := &Config{
		HTTP: HTTPConfig{
			Port: getEnvAsInt("HTTP_PORT", 8080),
		},
		Vault: VaultConfig{
			Address: viper.GetString("VAULT_ADDR"),
			Token:   viper.GetString("VAULT_TOKEN"),
			Path:    viper.GetString("VAULT_SECRET_PATH"),
		},
	}

	if err := loadSecretsFromVault(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadSecretsFromVault(cfg *Config) error {
	client, err := api.NewClient(&api.Config{
		Address: cfg.Vault.Address,
	})
	if err != nil {
		return err
	}
	client.SetToken(cfg.Vault.Token)

	secret, err := client.Logical().Read(cfg.Vault.Path)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if raw, ok := secret.Data["data"]; ok {
		data = raw.(map[string]interface{})
	} else {
		data = secret.Data
	}

	cfg.Postgres = PostgresConfig{
		DSN: data["POSTGRES_DSN"].(string),
	}

	cfg.JWT = JWTConfig{
		SecretKey:       data["JWT_SECRET"].(string),
		AccessTokenTTL:  parseDuration(data["ACCESS_TOKEN_TTL"], "15m"),
		RefreshTokenTTL: parseDuration(data["REFRESH_TOKEN_TTL"], "720h"),
	}

	cfg.Redis = RedisConfig{
		Addr:     data["REDIS_ADDR"].(string),
		Password: data["REDIS_PASSWORD"].(string),
	}

	return nil
}

func getEnvAsInt(name string, defaultVal int) int {
	valStr := viper.GetString(name)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func parseDuration(val interface{}, fallback string) time.Duration {
	str, ok := val.(string)
	if !ok {
		str = fallback
	}
	dur, err := time.ParseDuration(str)
	if err != nil {
		dur, _ = time.ParseDuration(fallback)
	}
	return dur
}
