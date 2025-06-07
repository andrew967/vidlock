package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

type VaultKeyStore struct {
	client *api.Client
	prefix string
}

func NewVaultKeyStore(addr, token, prefix string) (*VaultKeyStore, error) {
	cfg := &api.Config{
		Address: addr,
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("vault client init: %w", err)
	}

	client.SetToken(token)

	return &VaultKeyStore{
		client: client,
		prefix: prefix,
	}, nil
}

func (v *VaultKeyStore) Save(videoID, chunkID string, key []byte) error {
	path := fmt.Sprintf("%s/%s/%s", v.prefix, videoID, chunkID)

	data := map[string]interface{}{
		"key": key,
	}

	_, err := v.client.KVv2("secret").Put(context.Background(), path, data)
	if err != nil {
		return fmt.Errorf("vault put: %w", err)
	}

	return nil
}
