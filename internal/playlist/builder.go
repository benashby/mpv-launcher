package playlist

import (
	"mpv-launcher/internal/priority"
	"mpv-launcher/internal/tracks"
)

// FilePlayback stores the selected tracks for a single file
type FilePlayback struct {
	Filename      string
	AudioTrack    int  // MpvIndex for audio track
	SubtitleTrack int  // MpvIndex for subtitle track, -1 = disabled
	UsedPriority  int  // Which priority level was used (0-based, -1 for fallback)
	IsFallback    bool // True if no priority matched
}

// Builder creates per-file track selections based on priorities
type Builder struct {
	priorities []priority.Permutation
	files      []tracks.FileTrackInfo
}

// NewBuilder creates a new playlist builder
func NewBuilder(priorities []priority.Permutation, files []tracks.FileTrackInfo) *Builder {
	return &Builder{
		priorities: priorities,
		files:      files,
	}
}

// Build returns FilePlayback for each file based on priority order
func (b *Builder) Build() []FilePlayback {
	result := make([]FilePlayback, len(b.files))

	for i, file := range b.files {
		result[i] = b.selectTracksForFile(&file)
	}

	return result
}

// selectTracksForFile selects the best tracks for a single file based on priorities
func (b *Builder) selectTracksForFile(file *tracks.FileTrackInfo) FilePlayback {
	playback := FilePlayback{
		Filename:     file.Filename,
		UsedPriority: -1,
		IsFallback:   true,
	}

	// Try each priority in order
	for i, perm := range b.priorities {
		audioIdx, subIdx, ok := perm.GetTracks(file)
		if ok {
			playback.AudioTrack = audioIdx
			playback.SubtitleTrack = subIdx
			playback.UsedPriority = i
			playback.IsFallback = false
			return playback
		}
	}

	// Fallback: use first available tracks
	if len(file.AudioTracks) > 0 {
		playback.AudioTrack = file.AudioTracks[0].MpvIndex
	}
	if len(file.SubtitleTracks) > 0 {
		playback.SubtitleTrack = file.SubtitleTracks[0].MpvIndex
	} else {
		playback.SubtitleTrack = -1
	}

	return playback
}

// Summary returns a human-readable summary of track selections
func (b *Builder) Summary(playback []FilePlayback) map[string]int {
	summary := make(map[string]int)

	for _, pb := range playback {
		var label string
		if pb.IsFallback {
			label = "Fallback"
		} else if pb.UsedPriority >= 0 && pb.UsedPriority < len(b.priorities) {
			label = b.priorities[pb.UsedPriority].Label
		} else {
			label = "Unknown"
		}
		summary[label]++
	}

	return summary
}
