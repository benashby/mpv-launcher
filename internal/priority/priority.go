package priority

import (
	"mpv-launcher/internal/tracks"
)

// Permutation defines an audio/subtitle combination
type Permutation struct {
	ID           string // "jpn_eng", "eng_none", etc.
	AudioLang    string // "eng", "jpn"
	SubtitleLang string // "eng", "" (empty means disabled)
	Label        string // Human-readable label
}

// DefaultPermutations returns the 4 permutations in default priority order
func DefaultPermutations() []Permutation {
	return []Permutation{
		{
			ID:           "jpn_eng",
			AudioLang:    "jpn",
			SubtitleLang: "eng",
			Label:        "Japanese + Eng Subtitles",
		},
		{
			ID:           "eng_none",
			AudioLang:    "eng",
			SubtitleLang: "",
			Label:        "English - No Subtitles",
		},
		{
			ID:           "jpn_none",
			AudioLang:    "jpn",
			SubtitleLang: "",
			Label:        "Japanese - No Subtitles",
		},
		{
			ID:           "eng_eng",
			AudioLang:    "eng",
			SubtitleLang: "eng",
			Label:        "English + Eng Subtitles",
		},
	}
}

// CanSatisfy checks if a FileTrackInfo can satisfy this permutation
func (p Permutation) CanSatisfy(info *tracks.FileTrackInfo) bool {
	// Check if the required audio language is available
	hasAudio := false
	for _, track := range info.AudioTracks {
		if track.GetEffectiveLanguage() == p.AudioLang {
			hasAudio = true
			break
		}
	}
	if !hasAudio {
		return false
	}

	// If subtitles are disabled, we're satisfied
	if p.SubtitleLang == "" {
		return true
	}

	// Check if the required subtitle language is available
	for _, track := range info.SubtitleTracks {
		if track.GetEffectiveLanguage() == p.SubtitleLang {
			return true
		}
	}
	return false
}

// GetTracks returns the MpvIndex for audio and subtitle tracks
// audioIdx is always > 0, subIdx is > 0 for subtitles or -1 if disabled
// ok is false if the permutation cannot be satisfied
func (p Permutation) GetTracks(info *tracks.FileTrackInfo) (audioIdx, subIdx int, ok bool) {
	// Find audio track
	audioIdx = 0
	for _, track := range info.AudioTracks {
		if track.GetEffectiveLanguage() == p.AudioLang {
			audioIdx = track.MpvIndex
			break
		}
	}
	if audioIdx == 0 {
		return 0, 0, false
	}

	// If subtitles are disabled
	if p.SubtitleLang == "" {
		return audioIdx, -1, true
	}

	// Find subtitle track
	subIdx = 0
	for _, track := range info.SubtitleTracks {
		if track.GetEffectiveLanguage() == p.SubtitleLang {
			subIdx = track.MpvIndex
			break
		}
	}
	if subIdx == 0 {
		return 0, 0, false
	}

	return audioIdx, subIdx, true
}
