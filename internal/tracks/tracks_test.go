package tracks

import (
	"os"
	"path/filepath"
	"testing"
)

// Test file paths - these will be actual media files for integration tests
const (
	testMediaDir = "/mnt/media/family/tv/Anne Shirley/Season 1"
)

// TestProbeFile_RealMedia tests probing an actual media file
func TestProbeFile_RealMedia(t *testing.T) {
	// Find a test file
	testFile := filepath.Join(testMediaDir, "Anne Shirley - S01E05 - Let Us Look on the Bright Side of Things HDTV-720p Proper.mkv")

	// Check if file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	info, err := ProbeFile(testFile)
	if err != nil {
		t.Fatalf("ProbeFile failed: %v", err)
	}

	// Verify we got the file info
	if info.Filename != testFile {
		t.Errorf("Filename mismatch: got %s, want %s", info.Filename, testFile)
	}

	// Check that we found audio tracks
	if len(info.AudioTracks) == 0 {
		t.Error("Expected at least one audio track, got none")
	}

	// Check that we found subtitle tracks
	if len(info.SubtitleTracks) == 0 {
		t.Error("Expected at least one subtitle track, got none")
	}

	// Verify audio tracks have expected properties
	for i, track := range info.AudioTracks {
		if track.Type != TrackTypeAudio {
			t.Errorf("Audio track %d has wrong type: %v", i, track.Type)
		}
		if track.Codec == "" {
			t.Errorf("Audio track %d has empty codec", i)
		}
		if track.Language == "" {
			t.Logf("Warning: Audio track %d has no language", i)
		}
	}

	// Verify subtitle tracks have expected properties
	for i, track := range info.SubtitleTracks {
		if track.Type != TrackTypeSubtitle {
			t.Errorf("Subtitle track %d has wrong type: %v", i, track.Type)
		}
		if track.Codec == "" {
			t.Errorf("Subtitle track %d has empty codec", i)
		}
		if track.Language == "" {
			t.Logf("Warning: Subtitle track %d has no language", i)
		}
	}

	// Log track info for debugging
	t.Logf("Found %d audio tracks:", len(info.AudioTracks))
	for _, track := range info.AudioTracks {
		t.Logf("  [%d] %s - %s (%s) - %s", track.Index, track.Type, track.Language, track.Codec, track.Title)
	}

	t.Logf("Found %d subtitle tracks:", len(info.SubtitleTracks))
	for _, track := range info.SubtitleTracks {
		t.Logf("  [%d] %s - %s (%s) - %s", track.Index, track.Type, track.Language, track.Codec, track.Title)
	}
}

// TestProbeFile_NonExistent tests probing a non-existent file
func TestProbeFile_NonExistent(t *testing.T) {
	info, err := ProbeFile("/nonexistent/file.mkv")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
	if info == nil {
		t.Error("Expected non-nil info even on error")
	}
}

// TestHasLanguage tests the HasLanguage helper function
func TestHasLanguage(t *testing.T) {
	info := &FileTrackInfo{
		Filename: "test.mkv",
		AudioTracks: []Track{
			{Index: 0, Type: TrackTypeAudio, Language: "eng", Title: "English"},
			{Index: 1, Type: TrackTypeAudio, Language: "jpn", Title: "Japanese"},
		},
		SubtitleTracks: []Track{
			{Index: 2, Type: TrackTypeSubtitle, Language: "eng", Title: "English"},
			{Index: 3, Type: TrackTypeSubtitle, Language: "fre", Title: "French"},
		},
	}

	tests := []struct {
		trackType TrackType
		language  string
		expected  bool
	}{
		{TrackTypeAudio, "eng", true},
		{TrackTypeAudio, "jpn", true},
		{TrackTypeAudio, "fre", false},
		{TrackTypeSubtitle, "eng", true},
		{TrackTypeSubtitle, "fre", true},
		{TrackTypeSubtitle, "jpn", false},
	}

	for _, tt := range tests {
		result := info.HasLanguage(tt.trackType, tt.language)
		if result != tt.expected {
			t.Errorf("HasLanguage(%v, %s) = %v, want %v", tt.trackType, tt.language, result, tt.expected)
		}
	}
}

// TestGetTracksByLanguage tests the GetTracksByLanguage helper function
func TestGetTracksByLanguage(t *testing.T) {
	info := &FileTrackInfo{
		Filename: "test.mkv",
		AudioTracks: []Track{
			{Index: 0, Type: TrackTypeAudio, Language: "eng", Title: "English"},
			{Index: 1, Type: TrackTypeAudio, Language: "jpn", Title: "Japanese"},
		},
		SubtitleTracks: []Track{
			{Index: 2, Type: TrackTypeSubtitle, Language: "eng", Title: "English"},
			{Index: 3, Type: TrackTypeSubtitle, Language: "eng", Title: "English[CC]"},
			{Index: 4, Type: TrackTypeSubtitle, Language: "fre", Title: "French"},
		},
	}

	// Test audio tracks
	engAudio := info.GetTracksByLanguage(TrackTypeAudio, "eng")
	if len(engAudio) != 1 {
		t.Errorf("Expected 1 English audio track, got %d", len(engAudio))
	}

	// Test subtitle tracks with multiple matches
	engSubtitles := info.GetTracksByLanguage(TrackTypeSubtitle, "eng")
	if len(engSubtitles) != 2 {
		t.Errorf("Expected 2 English subtitle tracks, got %d", len(engSubtitles))
	}

	// Test language not present
	gerSubtitles := info.GetTracksByLanguage(TrackTypeSubtitle, "ger")
	if len(gerSubtitles) != 0 {
		t.Errorf("Expected 0 German subtitle tracks, got %d", len(gerSubtitles))
	}
}

// TestAnalyzeFiles_RealMedia tests analyzing multiple real media files
func TestAnalyzeFiles_RealMedia(t *testing.T) {
	// Find test files
	testFiles := []string{
		filepath.Join(testMediaDir, "Anne Shirley - S01E05 - Let Us Look on the Bright Side of Things HDTV-720p Proper.mkv"),
		filepath.Join(testMediaDir, "Anne Shirley - S01E09 - Next to Trying and Winning, the Best Thing Is Trying and Failing HDTV-720p Proper.mkv"),
	}

	// Check if files exist
	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Skipf("Test files not found in: %s", testMediaDir)
		}
	}

	analysis := AnalyzeFiles(testFiles)

	// Verify we analyzed all files
	if len(analysis.Files) != len(testFiles) {
		t.Errorf("Expected %d files analyzed, got %d", len(testFiles), len(analysis.Files))
	}

	// Verify no errors
	for i, fileInfo := range analysis.Files {
		if fileInfo.Error != nil {
			t.Errorf("File %d (%s) had error: %v", i, fileInfo.Filename, fileInfo.Error)
		}
	}

	// Check that we found common languages
	if len(analysis.CommonAudioLanguages) == 0 {
		t.Error("Expected at least one common audio language")
	}

	if len(analysis.CommonSubtitleLanguages) == 0 {
		t.Error("Expected at least one common subtitle language")
	}

	// Log analysis results
	t.Logf("Analysis of %d files:", len(testFiles))
	t.Logf("All audio languages: %v", analysis.AllAudioLanguages)
	t.Logf("Common audio languages: %v", analysis.CommonAudioLanguages)
	t.Logf("All subtitle languages: %v", analysis.AllSubtitleLanguages)
	t.Logf("Common subtitle languages: %v", analysis.CommonSubtitleLanguages)

	t.Logf("Common audio tracks (%d):", len(analysis.CommonAudioTracks))
	for _, track := range analysis.CommonAudioTracks {
		t.Logf("  %s: %s (%s)", track.Type, track.Language, track.Title)
	}

	t.Logf("Common subtitle tracks (%d):", len(analysis.CommonSubtitleTracks))
	for _, track := range analysis.CommonSubtitleTracks {
		t.Logf("  %s: %s (%s)", track.Type, track.Language, track.Title)
	}
}

// TestAnalyzeFiles_WithError tests analyzing when some files have errors
func TestAnalyzeFiles_WithError(t *testing.T) {
	testFiles := []string{
		"/nonexistent/file1.mkv",
		"/nonexistent/file2.mkv",
	}

	analysis := AnalyzeFiles(testFiles)

	// Verify we attempted to analyze all files
	if len(analysis.Files) != len(testFiles) {
		t.Errorf("Expected %d files in analysis, got %d", len(testFiles), len(analysis.Files))
	}

	// Verify all files have errors
	for i, fileInfo := range analysis.Files {
		if fileInfo.Error == nil {
			t.Errorf("File %d should have error, got nil", i)
		}
	}

	// When all files error, common languages should be empty
	if len(analysis.CommonAudioLanguages) != 0 {
		t.Error("Expected no common audio languages when all files error")
	}

	if len(analysis.CommonSubtitleLanguages) != 0 {
		t.Error("Expected no common subtitle languages when all files error")
	}
}

// TestAnalyzeFiles_MixedTracks tests analyzing files with different track configurations
func TestAnalyzeFiles_MixedTracks(t *testing.T) {
	// This is a conceptual test - in reality would need different test files
	// For now, we'll create FileTrackInfo manually

	analysis := &TrackAnalysis{
		Files: []FileTrackInfo{
			{
				Filename: "file1.mkv",
				AudioTracks: []Track{
					{Language: "eng", Title: "English"},
					{Language: "jpn", Title: "Japanese"},
				},
				SubtitleTracks: []Track{
					{Language: "eng", Title: "English"},
					{Language: "fre", Title: "French"},
				},
			},
			{
				Filename: "file2.mkv",
				AudioTracks: []Track{
					{Language: "eng", Title: "English"},
				},
				SubtitleTracks: []Track{
					{Language: "eng", Title: "English"},
					{Language: "fre", Title: "French"},
					{Language: "ger", Title: "German"},
				},
			},
		},
	}

	// Manually calculate what we expect
	// Common audio: eng (in both files)
	// All audio: eng, jpn
	// Common subtitles: eng, fre (in both files)
	// All subtitles: eng, fre, ger

	// This test demonstrates the data structure
	// The actual AnalyzeFiles function would populate this correctly
	t.Logf("Test demonstrates structure with %d files", len(analysis.Files))
	t.Logf("File 1 audio tracks: %d, subtitle tracks: %d",
		len(analysis.Files[0].AudioTracks), len(analysis.Files[0].SubtitleTracks))
	t.Logf("File 2 audio tracks: %d, subtitle tracks: %d",
		len(analysis.Files[1].AudioTracks), len(analysis.Files[1].SubtitleTracks))
}

// TestTrackSummary_Uniqueness tests that TrackSummary can be used as a map key
func TestTrackSummary_Uniqueness(t *testing.T) {
	trackCounts := make(map[TrackSummary]int)

	// Add some tracks
	trackCounts[TrackSummary{Type: TrackTypeAudio, Language: "eng", Title: "English"}]++
	trackCounts[TrackSummary{Type: TrackTypeAudio, Language: "eng", Title: "English"}]++ // Same
	trackCounts[TrackSummary{Type: TrackTypeAudio, Language: "jpn", Title: "Japanese"}]++

	if len(trackCounts) != 2 {
		t.Errorf("Expected 2 unique tracks, got %d", len(trackCounts))
	}

	engCount := trackCounts[TrackSummary{Type: TrackTypeAudio, Language: "eng", Title: "English"}]
	if engCount != 2 {
		t.Errorf("Expected English track count of 2, got %d", engCount)
	}
}
