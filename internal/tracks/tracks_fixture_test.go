package tracks_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"mpv-launcher/internal/testutil"
	"mpv-launcher/internal/testutil/probers"
	"mpv-launcher/internal/tracks"
)

// getTestdataPath returns the path to a testdata file
func getTestdataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "testdata", filename)
}

func TestProbeFile_WithFixtures(t *testing.T) {
	tests := []struct {
		name               string
		fixture            string
		expectedAudioCount int
		expectedSubCount   int
		expectedAudioLangs []string
		expectedSubLangs   []string
	}{
		{
			name:               "anime dual audio",
			fixture:            "anime_dual_audio.json",
			expectedAudioCount: 2,
			expectedSubCount:   1,
			expectedAudioLangs: []string{"jpn", "eng"},
			expectedSubLangs:   []string{"eng"},
		},
		{
			name:               "multi subtitle",
			fixture:            "multi_subtitle.json",
			expectedAudioCount: 1,
			expectedSubCount:   3,
			expectedAudioLangs: []string{"eng"},
			expectedSubLangs:   []string{"eng", "fre", "ger"},
		},
		{
			name:               "no language tags",
			fixture:            "no_language_tags.json",
			expectedAudioCount: 1,
			expectedSubCount:   1,
			expectedAudioLangs: []string{},
			expectedSubLangs:   []string{},
		},
		{
			name:               "single audio",
			fixture:            "single_audio.json",
			expectedAudioCount: 1,
			expectedSubCount:   0,
			expectedAudioLangs: []string{"eng"},
			expectedSubLangs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := getTestdataPath(tt.fixture)
			output := testutil.MustLoadFixture(t, fixturePath)

			// Use testutil's ToFileTrackInfo which returns testutil types
			info := output.ToFileTrackInfo("/virtual/test.mkv")

			// Check audio track count
			if len(info.AudioTracks) != tt.expectedAudioCount {
				t.Errorf("audio track count: got %d, want %d", len(info.AudioTracks), tt.expectedAudioCount)
			}

			// Check subtitle track count
			if len(info.SubtitleTracks) != tt.expectedSubCount {
				t.Errorf("subtitle track count: got %d, want %d", len(info.SubtitleTracks), tt.expectedSubCount)
			}

			// Check audio languages
			audioLangs := make(map[string]bool)
			for _, track := range info.AudioTracks {
				if track.Language != "" {
					audioLangs[track.Language] = true
				}
			}
			for _, lang := range tt.expectedAudioLangs {
				if !audioLangs[lang] {
					t.Errorf("expected audio language %s not found", lang)
				}
			}

			// Check subtitle languages
			subLangs := make(map[string]bool)
			for _, track := range info.SubtitleTracks {
				if track.Language != "" {
					subLangs[track.Language] = true
				}
			}
			for _, lang := range tt.expectedSubLangs {
				if !subLangs[lang] {
					t.Errorf("expected subtitle language %s not found", lang)
				}
			}
		})
	}
}

func TestAnalyzeFilesWithProber_SingleFile(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	fixturePath := getTestdataPath("anime_dual_audio.json")
	if err := prober.LoadFixture("/virtual/anime.mkv", fixturePath); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	analysis := tracks.AnalyzeFilesWithProber([]string{"/virtual/anime.mkv"}, prober)

	if len(analysis.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(analysis.Files))
	}

	if analysis.Files[0].Error != nil {
		t.Errorf("unexpected error: %v", analysis.Files[0].Error)
	}

	// Check that we found audio languages
	if len(analysis.AllAudioLanguages) != 2 {
		t.Errorf("expected 2 audio languages, got %d: %v", len(analysis.AllAudioLanguages), analysis.AllAudioLanguages)
	}

	// Check that we found subtitle languages
	if len(analysis.AllSubtitleLanguages) != 1 {
		t.Errorf("expected 1 subtitle language, got %d: %v", len(analysis.AllSubtitleLanguages), analysis.AllSubtitleLanguages)
	}
}

func TestAnalyzeFilesWithProber_MultipleFiles(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	// Load fixtures for multiple files
	if err := prober.LoadFixture("/virtual/ep1.mkv", getTestdataPath("anime_dual_audio.json")); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	if err := prober.LoadFixture("/virtual/ep2.mkv", getTestdataPath("anime_dual_audio.json")); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	analysis := tracks.AnalyzeFilesWithProber([]string{"/virtual/ep1.mkv", "/virtual/ep2.mkv"}, prober)

	if len(analysis.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(analysis.Files))
	}

	// Both files have same tracks, so common should equal all
	testutil.CompareStringSlices(t, "AllAudioLanguages", []string{"jpn", "eng"}, analysis.AllAudioLanguages)
	testutil.CompareStringSlices(t, "CommonAudioLanguages", []string{"jpn", "eng"}, analysis.CommonAudioLanguages)
	testutil.CompareStringSlices(t, "AllSubtitleLanguages", []string{"eng"}, analysis.AllSubtitleLanguages)
	testutil.CompareStringSlices(t, "CommonSubtitleLanguages", []string{"eng"}, analysis.CommonSubtitleLanguages)
}

func TestAnalyzeFilesWithProber_DifferingTracks(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	// Load different fixtures to simulate files with different tracks
	if err := prober.LoadFixture("/virtual/anime.mkv", getTestdataPath("anime_dual_audio.json")); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	if err := prober.LoadFixture("/virtual/movie.mkv", getTestdataPath("multi_subtitle.json")); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	analysis := tracks.AnalyzeFilesWithProber([]string{"/virtual/anime.mkv", "/virtual/movie.mkv"}, prober)

	if len(analysis.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(analysis.Files))
	}

	// anime has: jpn+eng audio, eng subs
	// movie has: eng audio, eng+fre+ger subs
	// Common audio: eng
	// Common subs: eng

	testutil.CompareStringSlices(t, "CommonAudioLanguages", []string{"eng"}, analysis.CommonAudioLanguages)
	testutil.CompareStringSlices(t, "CommonSubtitleLanguages", []string{"eng"}, analysis.CommonSubtitleLanguages)

	// All languages should include everything
	if len(analysis.AllAudioLanguages) != 2 {
		t.Errorf("expected 2 all audio languages (jpn, eng), got %d: %v", len(analysis.AllAudioLanguages), analysis.AllAudioLanguages)
	}
	if len(analysis.AllSubtitleLanguages) != 3 {
		t.Errorf("expected 3 all subtitle languages (eng, fre, ger), got %d: %v", len(analysis.AllSubtitleLanguages), analysis.AllSubtitleLanguages)
	}
}

func TestAnalyzeFilesWithProber_MissingFile(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	// Only load one fixture, leaving the other missing
	if err := prober.LoadFixture("/virtual/exists.mkv", getTestdataPath("single_audio.json")); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	analysis := tracks.AnalyzeFilesWithProber([]string{"/virtual/exists.mkv", "/virtual/missing.mkv"}, prober)

	if len(analysis.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(analysis.Files))
	}

	// First file should succeed
	if analysis.Files[0].Error != nil {
		t.Errorf("expected no error for first file, got: %v", analysis.Files[0].Error)
	}

	// Second file should have error
	if analysis.Files[1].Error == nil {
		t.Error("expected error for missing file")
	}
}

func TestFixtureProber_TrackDetails(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	fixturePath := getTestdataPath("anime_dual_audio.json")
	if err := prober.LoadFixture("/virtual/test.mkv", fixturePath); err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	info, err := prober.ProbeFile("/virtual/test.mkv")
	if err != nil {
		t.Fatalf("ProbeFile failed: %v", err)
	}

	// Verify track details
	if len(info.AudioTracks) != 2 {
		t.Fatalf("expected 2 audio tracks, got %d", len(info.AudioTracks))
	}

	// First audio track should be Japanese
	if info.AudioTracks[0].Language != "jpn" {
		t.Errorf("first audio track language: got %s, want jpn", info.AudioTracks[0].Language)
	}
	if info.AudioTracks[0].Title != "Japanese" {
		t.Errorf("first audio track title: got %s, want Japanese", info.AudioTracks[0].Title)
	}
	if info.AudioTracks[0].MpvIndex != 1 {
		t.Errorf("first audio track MpvIndex: got %d, want 1", info.AudioTracks[0].MpvIndex)
	}

	// Second audio track should be English
	if info.AudioTracks[1].Language != "eng" {
		t.Errorf("second audio track language: got %s, want eng", info.AudioTracks[1].Language)
	}
	if info.AudioTracks[1].MpvIndex != 2 {
		t.Errorf("second audio track MpvIndex: got %d, want 2", info.AudioTracks[1].MpvIndex)
	}

	// Subtitle track should be English
	if len(info.SubtitleTracks) != 1 {
		t.Fatalf("expected 1 subtitle track, got %d", len(info.SubtitleTracks))
	}
	if info.SubtitleTracks[0].Language != "eng" {
		t.Errorf("subtitle track language: got %s, want eng", info.SubtitleTracks[0].Language)
	}
	if info.SubtitleTracks[0].MpvIndex != 1 {
		t.Errorf("subtitle track MpvIndex: got %d, want 1", info.SubtitleTracks[0].MpvIndex)
	}
}
