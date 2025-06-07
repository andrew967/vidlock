package ffmpeg

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type WatermarkProcessor struct {
	FontPath string
}

func NewWatermarkProcessor(fontPath string) *WatermarkProcessor {
	return &WatermarkProcessor{
		FontPath: fontPath,
	}
}

func (p *WatermarkProcessor) ApplyWatermark(inputPath string) (string, error) {
	outputPath := tempOutputPath(inputPath)

	args := []string{
		"-i", inputPath,
		"-vf", fmt.Sprintf(`drawtext=fontfile=%s:text='VIDLOCK':fontcolor=white:fontsize=24:x=10:y=H-th-10`, p.FontPath),
		"-codec:a", "copy",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	return outputPath, nil
}

func tempOutputPath(input string) string {
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	timestamp := time.Now().UnixNano()
	return filepath.Join("/tmp", fmt.Sprintf("%s_watermarked_%d%s", name, timestamp, ext))
}
