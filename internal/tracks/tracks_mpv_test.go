package tracks

import (
	"path/filepath"
	"testing"
)

const testDir = "/mnt/media/ben/anime/High School DxD Hero"

// TestMpvIndexMapping verifies that FFmpeg stream indices are correctly mapped to MPV track indices
func TestMpvIndexMapping(t *testing.T) {
	testFile := filepath.Join(testDir, "[Judas]_High_School_DxD_Hero_-_01_(1920x1080_V_MPEGH_ISO_HEVC_10bit)_[A2EDB19B].mkv")

	info, err := ProbeFile(testFile)
	if err != nil {
		t.Skipf("Test file not available: %v", err)
	}

	// Verify audio tracks have correct MPV indices
	if len(info.AudioTracks) != 2 {
		t.Fatalf("Expected 2 audio tracks, got %d", len(info.AudioTracks))
	}

	// First audio track should be aid=1
	if info.AudioTracks[0].MpvIndex != 1 {
		t.Errorf("First audio track should have MpvIndex=1, got %d", info.AudioTracks[0].MpvIndex)
	}

	// Second audio track should be aid=2
	if info.AudioTracks[1].MpvIndex != 2 {
		t.Errorf("Second audio track should have MpvIndex=2, got %d", info.AudioTracks[1].MpvIndex)
	}

	// Verify subtitle tracks have correct MPV indices
	if len(info.SubtitleTracks) != 2 {
		t.Fatalf("Expected 2 subtitle tracks, got %d", len(info.SubtitleTracks))
	}

	// First subtitle track should be sid=1 (not sid=3 like FFmpeg index)
	if info.SubtitleTracks[0].MpvIndex != 1 {
		t.Errorf("First subtitle track should have MpvIndex=1, got %d (FFmpeg Index=%d)",
			info.SubtitleTracks[0].MpvIndex, info.SubtitleTracks[0].Index)
	}

	// Second subtitle track should be sid=2 (not sid=4 like FFmpeg index)
	if info.SubtitleTracks[1].MpvIndex != 2 {
		t.Errorf("Second subtitle track should have MpvIndex=2, got %d (FFmpeg Index=%d)",
			info.SubtitleTracks[1].MpvIndex, info.SubtitleTracks[1].Index)
	}
}

// TestMpvIndexVsFFmpegIndex verifies that MPV indices differ from FFmpeg indices
func TestMpvIndexVsFFmpegIndex(t *testing.T) {
	testFile := filepath.Join(testDir, "[Judas]_High_School_DxD_Hero_-_01_(1920x1080_V_MPEGH_ISO_HEVC_10bit)_[A2EDB19B].mkv")

	info, err := ProbeFile(testFile)
	if err != nil {
		t.Skipf("Test file not available: %v", err)
	}

	// For files with video streams, subtitle FFmpeg indices will be > 2
	// but MPV indices should start at 1
	for i, track := range info.SubtitleTracks {
		if track.Index == track.MpvIndex {
			t.Errorf("Subtitle track %d: FFmpeg Index (%d) should not equal MPV Index (%d) in files with video/audio streams",
				i, track.Index, track.MpvIndex)
		}

		// MPV index should be position+1 in the subtitle list
		expectedMpvIndex := i + 1
		if track.MpvIndex != expectedMpvIndex {
			t.Errorf("Subtitle track %d: expected MpvIndex=%d, got %d",
				i, expectedMpvIndex, track.MpvIndex)
		}
	}
}

// TestGetEffectiveLanguageWithMpvIndex ensures language detection works with MpvIndex
func TestGetEffectiveLanguageWithMpvIndex(t *testing.T) {
	testFile := filepath.Join(testDir, "[Judas]_High_School_DxD_Hero_-_01_(1920x1080_V_MPEGH_ISO_HEVC_10bit)_[A2EDB19B].mkv")

	info, err := ProbeFile(testFile)
	if err != nil {
		t.Skipf("Test file not available: %v", err)
	}

	// Find the track with "Japanese" in title
	var jpnTrack *Track
	for i := range info.AudioTracks {
		if info.AudioTracks[i].GetEffectiveLanguage() == "jpn" {
			jpnTrack = &info.AudioTracks[i]
			break
		}
	}

	if jpnTrack == nil {
		t.Fatal("Could not find Japanese audio track")
	}

	// Verify it has correct MPV index (should be aid=1 for first audio)
	if jpnTrack.MpvIndex != 1 {
		t.Errorf("Japanese track should have MpvIndex=1, got %d", jpnTrack.MpvIndex)
	}

	// Find the track with "English" in title
	var engTrack *Track
	for i := range info.AudioTracks {
		if info.AudioTracks[i].GetEffectiveLanguage() == "eng" {
			engTrack = &info.AudioTracks[i]
			break
		}
	}

	if engTrack == nil {
		t.Fatal("Could not find English audio track")
	}

	// Verify it has correct MPV index (should be aid=2 for second audio)
	if engTrack.MpvIndex != 2 {
		t.Errorf("English track should have MpvIndex=2, got %d", engTrack.MpvIndex)
	}
}

// TestAnalyzeFilesPreservesMpvIndex verifies batch analysis preserves MpvIndex
func TestAnalyzeFilesPreservesMpvIndex(t *testing.T) {
	testFile := filepath.Join(testDir, "[Judas]_High_School_DxD_Hero_-_01_(1920x1080_V_MPEGH_ISO_HEVC_10bit)_[A2EDB19B].mkv")

	analysis := AnalyzeFiles([]string{testFile})

	if len(analysis.Files) == 0 || analysis.Files[0].Error != nil {
		t.Skip("Test file not available")
	}

	fileInfo := &analysis.Files[0]

	// Verify all audio tracks have sequential MPV indices starting at 1
	for i, track := range fileInfo.AudioTracks {
		expected := i + 1
		if track.MpvIndex != expected {
			t.Errorf("Audio track %d: expected MpvIndex=%d, got %d",
				i, expected, track.MpvIndex)
		}
	}

	// Verify all subtitle tracks have sequential MPV indices starting at 1
	for i, track := range fileInfo.SubtitleTracks {
		expected := i + 1
		if track.MpvIndex != expected {
			t.Errorf("Subtitle track %d: expected MpvIndex=%d, got %d",
				i, expected, track.MpvIndex)
		}
	}
}
