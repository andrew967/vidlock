package ffmpeg

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ChunkSplitter struct {
	ChunkDurationSeconds int
}

func NewChunkSplitter(duration int) *ChunkSplitter {
	return &ChunkSplitter{
		ChunkDurationSeconds: duration,
	}
}

func (s *ChunkSplitter) Split(inputPath string) ([]string, error) {
	outputTemplate := tempChunkPattern(inputPath)

	args := []string{
		"-i", inputPath,
		"-c", "copy",
		"-map", "0",
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%d", s.ChunkDurationSeconds),
		outputTemplate,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg split failed: %w", err)
	}

	matches, err := filepath.Glob(strings.Replace(outputTemplate, "%03d", "*", 1))
	if err != nil {
		return nil, fmt.Errorf("failed to list chunks: %w", err)
	}

	return matches, nil
}

func tempChunkPattern(input string) string {
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	timestamp := time.Now().UnixNano()
	return filepath.Join("/tmp", fmt.Sprintf("%s_chunk_%d_%%03d%s", name, timestamp, ext))
}
