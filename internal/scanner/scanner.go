package scanner

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"myt-v/internal/db"
)

type FFProbeFormat struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func RunFFProbe(path string) (*db.Movie, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info FFProbeFormat
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, err
	}

	movie := &db.Movie{
		Path:  path,
		Title: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}

	if d := info.Format.Duration; d != "" {
		var dur float64
		fmt.Sscanf(d, "%f", &dur)
		movie.Duration = dur
	}

	for _, s := range info.Streams {
		if s.CodecType == "video" {
			movie.CodecV = s.CodecName
			movie.Width = s.Width
			movie.Height = s.Height
		}
		if s.CodecType == "audio" && movie.CodecA == "" {
			movie.CodecA = s.CodecName
		}
	}

	return movie, nil
}

func ScanDir(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// common extensions
		lower := strings.ToLower(filepath.Ext(path))
		if lower != ".mp4" && lower != ".mkv" && lower != ".avi" {
			return nil
		}

		// Does it already exist in the database?
		var count int64
		db.DB.Model(&db.Movie{}).Where("path = ?", path).Count(&count)
		if count > 0 {
			return nil
		}

		log.Printf("Processing %s ...", path)
		movie, err := RunFFProbe(path)
		if err != nil {
			log.Printf("Error ffprobe in %s: %v", path, err)
			return nil
		}
		db.DB.Create(movie)
		return nil
	})
}
