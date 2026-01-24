package tracks_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"mpv-launcher/internal/testutil"
	"mpv-launcher/internal/testutil/probers"
	"mpv-launcher/internal/tracks"
)

// getHsDxDTestdataPath returns the path to HsDxD Born testdata files
func getHsDxDTestdataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "testdata", "hsdxd_born", filename)
}

// TestHsDxDBorn_TrackDetection tests that audio tracks are correctly identified
// despite mislabeled language metadata
func TestHsDxDBorn_TrackDetection(t *testing.T) {
	fixtures := []string{"ep01.json", "ep02.json", "ep03.json"}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			fixturePath := getHsDxDTestdataPath(fixture)
			output := testutil.MustLoadFixture(t, fixturePath)
			info := output.ToFileTrackInfo("/virtual/" + fixture[:4] + ".mkv")

			// Should have 2 audio tracks
			if len(info.AudioTracks) != 2 {
				t.Fatalf("expected 2 audio tracks, got %d", len(info.AudioTracks))
			}

			// Should have 2 subtitle tracks
			if len(info.SubtitleTracks) != 2 {
				t.Fatalf("expected 2 subtitle tracks, got %d", len(info.SubtitleTracks))
			}

			// Log the raw track info
			t.Logf("Audio track 1: lang=%s, title=%s", info.AudioTracks[0].Language, info.AudioTracks[0].Title)
			t.Logf("Audio track 2: lang=%s, title=%s", info.AudioTracks[1].Language, info.AudioTracks[1].Title)
			t.Logf("Subtitle track 1: lang=%s, title=%s", info.SubtitleTracks[0].Language, info.SubtitleTracks[0].Title)
			t.Logf("Subtitle track 2: lang=%s, title=%s", info.SubtitleTracks[1].Language, info.SubtitleTracks[1].Title)
		})
	}
}

// TestHsDxDBorn_EffectiveLanguageDetection tests GetEffectiveLanguage
// correctly identifies Japanese from the title despite "eng" language code
func TestHsDxDBorn_EffectiveLanguageDetection(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	// Load fixtures
	fixtures := []struct {
		virtualPath string
		fixturePath string
	}{
		{"/virtual/ep01.mkv", getHsDxDTestdataPath("ep01.json")},
		{"/virtual/ep02.mkv", getHsDxDTestdataPath("ep02.json")},
		{"/virtual/ep03.mkv", getHsDxDTestdataPath("ep03.json")},
	}

	virtualPaths := make([]string, len(fixtures))
	for i, f := range fixtures {
		if err := prober.LoadFixture(f.virtualPath, f.fixturePath); err != nil {
			t.Fatalf("failed to load fixture %s: %v", f.fixturePath, err)
		}
		virtualPaths[i] = f.virtualPath
	}

	// Test single file
	info, err := prober.ProbeFile("/virtual/ep01.mkv")
	if err != nil {
		t.Fatalf("ProbeFile failed: %v", err)
	}

	// First audio track should be detected as Japanese from title
	jpnTrack := info.AudioTracks[0]
	effectiveLang := jpnTrack.GetEffectiveLanguage()
	if effectiveLang != "jpn" {
		t.Errorf("First audio track: expected effective language 'jpn', got '%s' (title: %s)", effectiveLang, jpnTrack.Title)
	}

	// Second audio track should be detected as English from title
	engTrack := info.AudioTracks[1]
	effectiveLang = engTrack.GetEffectiveLanguage()
	if effectiveLang != "eng" {
		t.Errorf("Second audio track: expected effective language 'eng', got '%s' (title: %s)", effectiveLang, engTrack.Title)
	}

	// Subtitle tracks should be English
	for i, sub := range info.SubtitleTracks {
		effectiveLang = sub.GetEffectiveLanguage()
		if effectiveLang != "eng" {
			t.Errorf("Subtitle track %d: expected effective language 'eng', got '%s'", i+1, effectiveLang)
		}
	}
}

// TestHsDxDBorn_AnalysisCommonLanguages tests that analyzing multiple episodes
// correctly identifies common languages using effective language detection
func TestHsDxDBorn_AnalysisCommonLanguages(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	// Load all fixtures
	fixtures := []struct {
		virtualPath string
		fixturePath string
	}{
		{"/virtual/ep01.mkv", getHsDxDTestdataPath("ep01.json")},
		{"/virtual/ep02.mkv", getHsDxDTestdataPath("ep02.json")},
		{"/virtual/ep03.mkv", getHsDxDTestdataPath("ep03.json")},
	}

	virtualPaths := make([]string, len(fixtures))
	for i, f := range fixtures {
		if err := prober.LoadFixture(f.virtualPath, f.fixturePath); err != nil {
			t.Fatalf("failed to load fixture %s: %v", f.fixturePath, err)
		}
		virtualPaths[i] = f.virtualPath
	}

	analysis := tracks.AnalyzeFilesWithProber(virtualPaths, prober)

	// Should have 3 files with no errors
	if len(analysis.Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(analysis.Files))
	}
	for i, f := range analysis.Files {
		if f.Error != nil {
			t.Errorf("file %d had error: %v", i, f.Error)
		}
	}

	// Common audio languages should include both Japanese and English
	// (after GetEffectiveLanguage detection from titles)
	t.Logf("All audio languages: %v", analysis.AllAudioLanguages)
	t.Logf("Common audio languages: %v", analysis.CommonAudioLanguages)
	t.Logf("All subtitle languages: %v", analysis.AllSubtitleLanguages)
	t.Logf("Common subtitle languages: %v", analysis.CommonSubtitleLanguages)

	// Check we found Japanese audio (detected from title)
	hasJpn := false
	for _, lang := range analysis.CommonAudioLanguages {
		if lang == "jpn" {
			hasJpn = true
			break
		}
	}
	if !hasJpn {
		t.Errorf("expected to find 'jpn' in common audio languages: %v", analysis.CommonAudioLanguages)
	}

	// Check we found English audio
	hasEng := false
	for _, lang := range analysis.CommonAudioLanguages {
		if lang == "eng" {
			hasEng = true
			break
		}
	}
	if !hasEng {
		t.Errorf("expected to find 'eng' in common audio languages: %v", analysis.CommonAudioLanguages)
	}

	// Check English subtitles are found
	hasEngSub := false
	for _, lang := range analysis.CommonSubtitleLanguages {
		if lang == "eng" {
			hasEngSub = true
			break
		}
	}
	if !hasEngSub {
		t.Errorf("expected to find 'eng' in common subtitle languages: %v", analysis.CommonSubtitleLanguages)
	}
}

// TestHsDxDBorn_ExpectedUserOptions validates that the user would be presented
// with the expected options: 1) English no subs  2) Japanese + English subs
func TestHsDxDBorn_ExpectedUserOptions(t *testing.T) {
	prober := probers.NewFixtureFileProber()

	fixtures := []struct {
		virtualPath string
		fixturePath string
	}{
		{"/virtual/ep01.mkv", getHsDxDTestdataPath("ep01.json")},
		{"/virtual/ep02.mkv", getHsDxDTestdataPath("ep02.json")},
		{"/virtual/ep03.mkv", getHsDxDTestdataPath("ep03.json")},
	}

	virtualPaths := make([]string, len(fixtures))
	for i, f := range fixtures {
		if err := prober.LoadFixture(f.virtualPath, f.fixturePath); err != nil {
			t.Fatalf("failed to load fixture %s: %v", f.fixturePath, err)
		}
		virtualPaths[i] = f.virtualPath
	}

	analysis := tracks.AnalyzeFilesWithProber(virtualPaths, prober)

	// The user should be able to:
	// Option 1: English audio (no subtitles) - hasLanguage("eng", audio) = true
	// Option 2: Japanese audio + English subs - hasLanguage("jpn", audio) && hasLanguage("eng", sub)

	// Verify Option 1: English audio is available in all files
	for i, f := range analysis.Files {
		if !f.HasLanguage(tracks.TrackTypeAudio, "eng") {
			t.Errorf("File %d: expected to have English audio for 'English no subs' option", i+1)
		}
	}

	// Verify Option 2: Japanese audio + English subtitles available in all files
	for i, f := range analysis.Files {
		if !f.HasLanguage(tracks.TrackTypeAudio, "jpn") {
			t.Errorf("File %d: expected to have Japanese audio for 'Japanese + English subs' option", i+1)
		}
		if !f.HasLanguage(tracks.TrackTypeSubtitle, "eng") {
			t.Errorf("File %d: expected to have English subtitles for 'Japanese + English subs' option", i+1)
		}
	}

	t.Logf("User options verified:")
	t.Logf("  Option 1: English audio (no subs) - AVAILABLE")
	t.Logf("  Option 2: Japanese audio + English subs - AVAILABLE")
}
