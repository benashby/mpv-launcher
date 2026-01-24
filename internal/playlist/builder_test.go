package playlist

import (
	"testing"

	"mpv-launcher/internal/priority"
	"mpv-launcher/internal/tracks"
)

func TestBuilder_Build_PriorityOrder(t *testing.T) {
	// Setup priorities: jpn_eng first, then eng_none
	priorities := []priority.Permutation{
		{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng", Label: "Japanese + Eng Subtitles"},
		{ID: "eng_none", AudioLang: "eng", SubtitleLang: "", Label: "English - No Subtitles"},
	}

	// File with dual audio and eng subs - should match jpn_eng
	files := []tracks.FileTrackInfo{
		{
			Filename: "episode1.mkv",
			AudioTracks: []tracks.Track{
				{MpvIndex: 1, Language: "jpn"},
				{MpvIndex: 2, Language: "eng"},
			},
			SubtitleTracks: []tracks.Track{
				{MpvIndex: 1, Language: "eng"},
			},
		},
	}

	builder := NewBuilder(priorities, files)
	result := builder.Build()

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if result[0].AudioTrack != 1 {
		t.Errorf("Expected audio track 1 (jpn), got %d", result[0].AudioTrack)
	}
	if result[0].SubtitleTrack != 1 {
		t.Errorf("Expected subtitle track 1 (eng), got %d", result[0].SubtitleTrack)
	}
	if result[0].UsedPriority != 0 {
		t.Errorf("Expected priority 0, got %d", result[0].UsedPriority)
	}
	if result[0].IsFallback {
		t.Error("Expected IsFallback to be false")
	}
}

func TestBuilder_Build_Fallback(t *testing.T) {
	// Setup priorities that won't match
	priorities := []priority.Permutation{
		{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng", Label: "Japanese + Eng Subtitles"},
	}

	// File with only English audio, no Japanese
	files := []tracks.FileTrackInfo{
		{
			Filename: "episode1.mkv",
			AudioTracks: []tracks.Track{
				{MpvIndex: 1, Language: "eng"},
			},
			SubtitleTracks: []tracks.Track{
				{MpvIndex: 1, Language: "fre"},
			},
		},
	}

	builder := NewBuilder(priorities, files)
	result := builder.Build()

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if !result[0].IsFallback {
		t.Error("Expected IsFallback to be true")
	}
	if result[0].UsedPriority != -1 {
		t.Errorf("Expected priority -1 for fallback, got %d", result[0].UsedPriority)
	}
	if result[0].AudioTrack != 1 {
		t.Errorf("Expected fallback to first audio track (1), got %d", result[0].AudioTrack)
	}
	if result[0].SubtitleTrack != 1 {
		t.Errorf("Expected fallback to first subtitle track (1), got %d", result[0].SubtitleTrack)
	}
}

func TestBuilder_Build_MixedFiles(t *testing.T) {
	priorities := []priority.Permutation{
		{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng", Label: "Japanese + Eng Subtitles"},
		{ID: "eng_none", AudioLang: "eng", SubtitleLang: "", Label: "English - No Subtitles"},
	}

	// Mix of files with different track configurations
	files := []tracks.FileTrackInfo{
		{
			// Regular episode: has both jpn and eng audio
			Filename: "S01E01.mkv",
			AudioTracks: []tracks.Track{
				{MpvIndex: 1, Language: "jpn"},
				{MpvIndex: 2, Language: "eng"},
			},
			SubtitleTracks: []tracks.Track{
				{MpvIndex: 1, Language: "eng"},
			},
		},
		{
			// Special: only has eng audio
			Filename: "S00E01.mkv",
			AudioTracks: []tracks.Track{
				{MpvIndex: 1, Language: "eng"},
			},
			SubtitleTracks: []tracks.Track{
				{MpvIndex: 1, Language: "eng"},
			},
		},
	}

	builder := NewBuilder(priorities, files)
	result := builder.Build()

	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result))
	}

	// First file should use jpn_eng (priority 0)
	if result[0].UsedPriority != 0 {
		t.Errorf("File 1: Expected priority 0, got %d", result[0].UsedPriority)
	}
	if result[0].AudioTrack != 1 {
		t.Errorf("File 1: Expected audio 1 (jpn), got %d", result[0].AudioTrack)
	}
	if result[0].SubtitleTrack != 1 {
		t.Errorf("File 1: Expected subtitle 1 (eng), got %d", result[0].SubtitleTrack)
	}

	// Second file should use eng_none (priority 1) since no jpn audio
	if result[1].UsedPriority != 1 {
		t.Errorf("File 2: Expected priority 1, got %d", result[1].UsedPriority)
	}
	if result[1].AudioTrack != 1 {
		t.Errorf("File 2: Expected audio 1 (eng), got %d", result[1].AudioTrack)
	}
	if result[1].SubtitleTrack != -1 {
		t.Errorf("File 2: Expected subtitle -1 (disabled), got %d", result[1].SubtitleTrack)
	}
}

func TestBuilder_Summary(t *testing.T) {
	priorities := []priority.Permutation{
		{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng", Label: "Japanese + Eng Subtitles"},
		{ID: "eng_none", AudioLang: "eng", SubtitleLang: "", Label: "English - No Subtitles"},
	}

	files := []tracks.FileTrackInfo{
		{Filename: "e1.mkv", AudioTracks: []tracks.Track{{MpvIndex: 1, Language: "jpn"}}, SubtitleTracks: []tracks.Track{{MpvIndex: 1, Language: "eng"}}},
		{Filename: "e2.mkv", AudioTracks: []tracks.Track{{MpvIndex: 1, Language: "jpn"}}, SubtitleTracks: []tracks.Track{{MpvIndex: 1, Language: "eng"}}},
		{Filename: "e3.mkv", AudioTracks: []tracks.Track{{MpvIndex: 1, Language: "eng"}}, SubtitleTracks: []tracks.Track{}},
	}

	builder := NewBuilder(priorities, files)
	result := builder.Build()
	summary := builder.Summary(result)

	if summary["Japanese + Eng Subtitles"] != 2 {
		t.Errorf("Expected 2 files with jpn_eng, got %d", summary["Japanese + Eng Subtitles"])
	}
	if summary["English - No Subtitles"] != 1 {
		t.Errorf("Expected 1 file with eng_none, got %d", summary["English - No Subtitles"])
	}
}
