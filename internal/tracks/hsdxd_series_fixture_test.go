package tracks_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"mpv-launcher/internal/testutil"
	"mpv-launcher/internal/testutil/probers"
	"mpv-launcher/internal/tracks"
)

// getSeriesTestdataPath returns the path to a series testdata file
func getSeriesTestdataPath(series, filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "testdata", series, filename)
}

// testSeriesCommonOptions tests that a series provides the two expected options:
// 1) English audio (no subtitles)
// 2) Japanese audio + English subtitles
func testSeriesCommonOptions(t *testing.T, seriesName, seriesDir string, fixtures []string) {
	t.Helper()

	prober := probers.NewFixtureFileProber()
	virtualPaths := make([]string, len(fixtures))

	for i, fixture := range fixtures {
		virtualPath := "/virtual/" + seriesDir + "/" + fixture[:4] + ".mkv"
		fixturePath := getSeriesTestdataPath(seriesDir, fixture)

		if err := prober.LoadFixture(virtualPath, fixturePath); err != nil {
			t.Fatalf("failed to load fixture %s: %v", fixturePath, err)
		}
		virtualPaths[i] = virtualPath
	}

	analysis := tracks.AnalyzeFilesWithProber(virtualPaths, prober)

	// Should have all files with no errors
	if len(analysis.Files) != len(fixtures) {
		t.Fatalf("expected %d files, got %d", len(fixtures), len(analysis.Files))
	}
	for i, f := range analysis.Files {
		if f.Error != nil {
			t.Errorf("file %d had error: %v", i, f.Error)
		}
	}

	// Log detected languages
	t.Logf("%s - All audio languages: %v", seriesName, analysis.AllAudioLanguages)
	t.Logf("%s - Common audio languages: %v", seriesName, analysis.CommonAudioLanguages)
	t.Logf("%s - All subtitle languages: %v", seriesName, analysis.AllSubtitleLanguages)
	t.Logf("%s - Common subtitle languages: %v", seriesName, analysis.CommonSubtitleLanguages)

	// Check Option 1: English audio is available in all files
	hasEngAudio := false
	for _, lang := range analysis.CommonAudioLanguages {
		if lang == "eng" {
			hasEngAudio = true
			break
		}
	}
	if !hasEngAudio {
		t.Errorf("%s: Option 1 FAILED - no common English audio found: %v", seriesName, analysis.CommonAudioLanguages)
	}

	// Check Option 2: Japanese audio is available in all files
	hasJpnAudio := false
	for _, lang := range analysis.CommonAudioLanguages {
		if lang == "jpn" {
			hasJpnAudio = true
			break
		}
	}
	if !hasJpnAudio {
		t.Errorf("%s: Option 2 FAILED - no common Japanese audio found: %v", seriesName, analysis.CommonAudioLanguages)
	}

	// Check Option 2: English subtitles are available in all files
	hasEngSubs := false
	for _, lang := range analysis.CommonSubtitleLanguages {
		if lang == "eng" {
			hasEngSubs = true
			break
		}
	}
	if !hasEngSubs {
		t.Errorf("%s: Option 2 FAILED - no common English subtitles found: %v", seriesName, analysis.CommonSubtitleLanguages)
	}

	// Verify each file individually
	for i, f := range analysis.Files {
		// Option 1: English audio
		if !f.HasLanguage(tracks.TrackTypeAudio, "eng") {
			t.Errorf("%s file %d: missing English audio for Option 1", seriesName, i+1)
		}
		// Option 2: Japanese audio + English subs
		if !f.HasLanguage(tracks.TrackTypeAudio, "jpn") {
			t.Errorf("%s file %d: missing Japanese audio for Option 2", seriesName, i+1)
		}
		if !f.HasLanguage(tracks.TrackTypeSubtitle, "eng") {
			t.Errorf("%s file %d: missing English subtitles for Option 2", seriesName, i+1)
		}
	}

	t.Logf("%s - User options verified:", seriesName)
	t.Logf("  Option 1: English audio (no subs) - %v", hasEngAudio)
	t.Logf("  Option 2: Japanese audio + English subs - %v", hasJpnAudio && hasEngSubs)
}

// TestHsDxD_Original tests the original High School DxD series
// Note: Both audio tracks are mislabeled as "eng" but titles contain "Japanese"/"English"
func TestHsDxD_Original(t *testing.T) {
	fixtures := []string{"ep01.json", "ep02.json", "ep03.json"}
	testSeriesCommonOptions(t, "High School DxD", "hsdxd", fixtures)
}

// TestHsDxD_Hero tests High School DxD Hero
// Note: Both audio tracks are mislabeled as "eng" but titles contain "Japanese"/"English"
// Also note: English title has typo "Englsih"
func TestHsDxD_Hero(t *testing.T) {
	fixtures := []string{"ep01.json", "ep02.json", "ep03.json"}
	testSeriesCommonOptions(t, "High School DxD Hero", "hsdxd_hero", fixtures)
}

// TestHsDxD_New tests High School DxD New
// Note: Japanese track correctly has lang=jpn, English has lang=eng
func TestHsDxD_New(t *testing.T) {
	fixtures := []string{"ep01.json", "ep02.json", "ep03.json"}
	testSeriesCommonOptions(t, "High School DxD New", "hsdxd_new", fixtures)
}

// TestHsDxD_AllSeries_TrackStructure logs the track structure for debugging
func TestHsDxD_AllSeries_TrackStructure(t *testing.T) {
	series := []struct {
		name      string
		dir       string
		fixtures  []string
	}{
		{"High School DxD", "hsdxd", []string{"ep01.json"}},
		{"High School DxD Born", "hsdxd_born", []string{"ep01.json"}},
		{"High School DxD Hero", "hsdxd_hero", []string{"ep01.json"}},
		{"High School DxD New", "hsdxd_new", []string{"ep01.json"}},
	}

	for _, s := range series {
		t.Run(s.name, func(t *testing.T) {
			fixturePath := getSeriesTestdataPath(s.dir, s.fixtures[0])
			output := testutil.MustLoadFixture(t, fixturePath)
			info := output.ToFileTrackInfo("/virtual/test.mkv")

			t.Logf("=== %s ===", s.name)
			t.Logf("Audio tracks: %d", len(info.AudioTracks))
			for i, track := range info.AudioTracks {
				t.Logf("  Audio %d: lang=%s, title=%s", i+1, track.Language, track.Title)
			}
			t.Logf("Subtitle tracks: %d", len(info.SubtitleTracks))
			for i, track := range info.SubtitleTracks {
				t.Logf("  Sub %d: lang=%s, title=%s", i+1, track.Language, track.Title)
			}
		})
	}
}
