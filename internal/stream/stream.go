package stream

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"mitv/internal/db"
)

type HLSConfig struct {
	HLSDir string
}

func paths(hlsDir string, id uint) (dir, m3u8, segfmt string) {
	idStr := strconv.FormatUint(uint64(id), 10)
	dir = filepath.Join(hlsDir, idStr)
	m3u8 = filepath.Join(dir, "master.m3u8")
	segfmt = filepath.Join(dir, "seg_%05d.ts")
	return
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func StartHLS(ctx context.Context, movie db.Movie, cfg HLSConfig) (string, error) {
	dir, m3u8, segfmt := paths(cfg.HLSDir, movie.ID)

	if err := ensureDir(dir); err != nil {
		return "", fmt.Errorf("no se pudo crear dir HLS: %w", err)
	}
	// if a playlist already exists, return it
	if exists(m3u8) {
		return fmt.Sprintf("/hls/%d/master.m3u8", movie.ID), nil
	}

	args := []string{
		// For VOD, we do NOT use -re -> faster startup
		"-i", movie.Path,
	}

	useCopy := (movie.CodecV == "h264" && movie.CodecA == "aac")
	if useCopy {
		args = append(args, "-c:v", "copy", "-c:a", "copy")
	} else {
		args = append(args,
			"-c:v", "libx264", "-preset", "veryfast", "-crf", "20",
			"-c:a", "aac", "-b:a", "160k",
			// Fixed GOP for predictable keyframes (improved zapping/startup)
			"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
		)
	}

	args = append(args,
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "0", // 0 = keep all segments
		"-hls_flags", "append_list+independent_segments",
		"-hls_playlist_type", "event", // list only grows
		"-hls_segment_filename", segfmt,
		m3u8,
	)

	go func() {
		c, cancel := context.WithCancel(ctx)
		defer cancel()

		cmd := exec.CommandContext(c, "ffmpeg", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("[HLS] ffmpeg ID=%d (copy=%v) -> %s", movie.ID, useCopy, dir)
		if err := cmd.Run(); err != nil {
			log.Printf("[HLS] ffmpeg finished ID=%d: %v", movie.ID, err)
		}
	}()

	// Wait until m3u8 is created (ffmpeg startup time)
	for i := 0; i < 40; i++ { // ~6s max
		if exists(m3u8) {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	return fmt.Sprintf("/hls/%d/master.m3u8", movie.ID), nil
}
