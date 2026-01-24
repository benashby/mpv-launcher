package ffmpeg

import (
	"testing"
)

func TestProbeVideo(t *testing.T) {
	// This is an example test structure
	// In a real scenario, you would need a test video file

	t.Run("invalid file path", func(t *testing.T) {
		_, err := ProbeVideo("/nonexistent/file.mp4")
		if err == nil {
			t.Error("Expected error for nonexistent file, got nil")
		}
	})

	// Add more tests with actual test video files
	// Example:
	// t.Run("valid video file", func(t *testing.T) {
	//     info, err := ProbeVideo("testdata/sample.mp4")
	//     if err != nil {
	//         t.Fatalf("Unexpected error: %v", err)
	//     }
	//     if info.Width <= 0 || info.Height <= 0 {
	//         t.Error("Invalid video dimensions")
	//     }
	// })
}

func TestVideoInfo(t *testing.T) {
	info := &VideoInfo{
		Duration:  1000000,
		Width:     1920,
		Height:    1080,
		Codec:     "h264",
		FrameRate: 30.0,
	}

	if info.Width != 1920 {
		t.Errorf("Expected width 1920, got %d", info.Width)
	}

	if info.Height != 1080 {
		t.Errorf("Expected height 1080, got %d", info.Height)
	}

	if info.Codec != "h264" {
		t.Errorf("Expected codec h264, got %s", info.Codec)
	}
}
