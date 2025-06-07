package ipfs

import (
	"context"
	"fmt"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFSUploader struct {
	sh *shell.Shell
}

func NewIPFSUploader(addr string) *IPFSUploader {
	return &IPFSUploader{
		sh: shell.NewShell(addr),
	}
}

func (u *IPFSUploader) Upload(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	cid, err := u.sh.Add(file)
	if err != nil {
		return "", fmt.Errorf("ipfs add: %w", err)
	}

	return "ipfs://" + cid, nil
}
