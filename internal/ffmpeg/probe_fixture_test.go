package ffmpeg_test

import (
	"math"
	"path/filepath"
	"runtime"
	"testing"

	"mpv-launcher/internal/testutil"
	"mpv-launcher/internal/testutil/probers"
)

// getTestdataPath returns the path to a testdata file
func getTestdataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "testdata", filename)
}

func TestProbeVideo_WithFixtures(t *testing.T) {
	tests := []struct {
		name            string
		fixture         string
		expectedWidth   int
		expectedHeight  int
		expectedCodec   string
		expectedFPS     float64
		expectedSeconds float64
	}{
		{
			name:            "1080p hevc",
			fixture:         "video_1080p_hevc.json",
			expectedWidth:   1920,
			expectedHeight:  1080,
			expectedCodec:   "hevc",
			expectedFPS:     23.976,
			expectedSeconds: 7200.0,
		},
		{
			name:            "4k hdr",
			fixture:         "video_4k_hdr.json",
			expectedWidth:   3840,
			expectedHeight:  2160,
			expectedCodec:   "hevc",
			expectedFPS:     59.94,
			expectedSeconds: 5400.0,
		},
		{
			name:            "720p h264",
			fixture:         "video_720p_h264.json",
			expectedWidth:   1280,
			expectedHeight:  720,
			expectedCodec:   "h264",
			expectedFPS:     30.0,
			expectedSeconds: 1800.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := getTestdataPath(tt.fixture)
			output := testutil.MustLoadFixture(t, fixturePath)

			// Use testutil's ToVideoInfo which returns testutil.VideoInfo
			info, err := output.ToVideoInfo()
			if err != nil {
				t.Fatalf("ToVideoInfo failed: %v", err)
			}

			if info.Width != tt.expectedWidth {
				t.Errorf("Width: got %d, want %d", info.Width, tt.expectedWidth)
			}

			if info.Height != tt.expectedHeight {
				t.Errorf("Height: got %d, want %d", info.Height, tt.expectedHeight)
			}

			if info.Codec != tt.expectedCodec {
				t.Errorf("Codec: got %s, want %s", info.Codec, tt.expectedCodec)
			}

			// Frame rate tolerance of 0.01
			if math.Abs(info.FrameRate-tt.expectedFPS) > 0.01 {
				t.Errorf("FrameRate: got %f, want %f", info.FrameRate, tt.expectedFPS)
			}

			// Duration is in microseconds, convert to seconds
			actualSeconds := float64(info.Duration) / 1000000.0
			if math.Abs(actualSeconds-tt.expectedSeconds) > 0.1 {
				t.Errorf("Duration: got %.2f seconds, want %.2f seconds", actualSeconds, tt.expectedSeconds)
			}
		})
	}
}

func TestFixtureVideoProber(t *testing.T) {
	prober := probers.NewFixtureVideoProber()

	fixturePath := getTestdataPath("video_1080p_hevc.json")
	if err := prober.LoadFixture("/virtual/movie.mkv", fixturePath); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	info, err := prober.ProbeVideo("/virtual/movie.mkv")
	if err != nil {
		t.Fatalf("ProbeVideo failed: %v", err)
	}

	if info.Width != 1920 {
		t.Errorf("Width: got %d, want 1920", info.Width)
	}

	if info.Height != 1080 {
		t.Errorf("Height: got %d, want 1080", info.Height)
	}

	if info.Codec != "hevc" {
		t.Errorf("Codec: got %s, want hevc", info.Codec)
	}
}

func TestFixtureVideoProber_MissingFixture(t *testing.T) {
	prober := probers.NewFixtureVideoProber()

	_, err := prober.ProbeVideo("/nonexistent/file.mkv")
	if err == nil {
		t.Error("expected error for missing fixture")
	}
}

func TestFixtureVideoProber_MultipleFixtures(t *testing.T) {
	prober := probers.NewFixtureVideoProber()

	// Load multiple fixtures
	if err := prober.LoadFixture("/virtual/movie1.mkv", getTestdataPath("video_1080p_hevc.json")); err != nil {
		t.Fatalf("failed to load fixture 1: %v", err)
	}
	if err := prober.LoadFixture("/virtual/movie2.mkv", getTestdataPath("video_4k_hdr.json")); err != nil {
		t.Fatalf("failed to load fixture 2: %v", err)
	}

	// Verify first fixture
	info1, err := prober.ProbeVideo("/virtual/movie1.mkv")
	if err != nil {
		t.Fatalf("ProbeVideo for movie1 failed: %v", err)
	}
	if info1.Width != 1920 || info1.Height != 1080 {
		t.Errorf("movie1: got %dx%d, want 1920x1080", info1.Width, info1.Height)
	}

	// Verify second fixture
	info2, err := prober.ProbeVideo("/virtual/movie2.mkv")
	if err != nil {
		t.Fatalf("ProbeVideo for movie2 failed: %v", err)
	}
	if info2.Width != 3840 || info2.Height != 2160 {
		t.Errorf("movie2: got %dx%d, want 3840x2160", info2.Width, info2.Height)
	}
}

func TestParseFrameRate(t *testing.T) {
	// Test via fixture conversion - the parser is internal
	fixturePath := getTestdataPath("video_720p_h264.json")
	output := testutil.MustLoadFixture(t, fixturePath)

	info, err := output.ToVideoInfo()
	if err != nil {
		t.Fatalf("ToVideoInfo failed: %v", err)
	}

	// 720p fixture has 30/1 frame rate
	if math.Abs(info.FrameRate-30.0) > 0.01 {
		t.Errorf("FrameRate: got %f, want 30.0", info.FrameRate)
	}
}
