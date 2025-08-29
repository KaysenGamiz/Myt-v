package stream

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"myt-v/internal/db"
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
	if exists(m3u8) {
		return fmt.Sprintf("/hls/%d/master.m3u8", movie.ID), nil
	}

	ext := strings.ToLower(filepath.Ext(movie.Path))
	isMP4Family := ext == ".mp4" || ext == ".mov" || ext == ".m4v"
	useCopy := (movie.CodecV == "h264" && movie.CodecA == "aac")
	gpu := os.Getenv("GPU")        // "nvenc" to enable NVIDIA encode
	down := os.Getenv("DOWNSCALE") // optional: "1080" to downscale in GPU

	// ---- 1) INPUT ARGS (must come before -i) ----
	inArgs := []string{}
	if !useCopy && strings.EqualFold(gpu, "nvenc") {
		inArgs = append(inArgs, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda")
	}

	// ---- 2) BASE CMD WITH -i ----
	args := append([]string{}, inArgs...)
	args = append(args, "-i", movie.Path)

	// Map only first video+audio and skip subtitles (faster, no .vtt)
	args = append(args, "-map", "0:v:0", "-map", "0:a:0?", "-sn")

	// ---- 3) CODECS ----
	if useCopy {
		// fastest path: remux only
		args = append(args, "-c:v", "copy", "-c:a", "copy")
		if isMP4Family {
			args = append(args, "-bsf:v", "h264_mp4toannexb")
		}
	} else {
		if strings.EqualFold(gpu, "nvenc") {
			// --- Build the GPU filter ---
			vf := "scale_cuda=format=nv12"
			if down == "1080" || (down == "" && movie.Width > 1920) {
				vf = "scale_cuda=-2:1080:format=nv12"
			}

			args = append(args,
				"-c:v", "h264_nvenc",
				"-preset", "p1", "-tune", "ll",
				"-rc", "vbr", "-cq", "21",
				"-b:v", "8M", "-maxrate", "10M", "-bufsize", "16M",
				"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
				"-force_key_frames", "expr:gte(t,n_forced*2)",
				"-bf", "0",
				"-profile:v", "high", "-level", "4.1",
				"-vf", vf, // <-- forces 10-bit -> 8-bit NV12 (and optional 4K->1080p)
			)
		} else {
			// CPU fallback (ultra fast)
			args = append(args,
				"-c:v", "libx264", "-preset", "ultrafast", "-crf", "23",
				"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
				"-force_key_frames", "expr:gte(t,n_forced*2)",
				"-pix_fmt", "yuv420p", "-profile:v", "high", "-level", "4.1",
			)
		}
		// audio (keep simple for compatibility)
		args = append(args, "-c:a", "aac", "-b:a", "160k", "-ac", "2")
	}

	// ---- 4) HLS (.ts) ----
	args = append(args,
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "0",
		"-hls_flags", "append_list+independent_segments",
		"-hls_playlist_type", "event",
		"-hls_segment_filename", segfmt,
		m3u8,
	)

	// Run ffmpeg
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

	// Wait: playlist + at least one segment
	for i := 0; i < 80; i++ { // ~12s
		if exists(m3u8) {
			if matches, _ := filepath.Glob(filepath.Join(dir, "seg_*.ts")); len(matches) > 0 {
				break
			}
		}
		time.Sleep(150 * time.Millisecond)
	}

	return fmt.Sprintf("/hls/%d/master.m3u8", movie.ID), nil
}
