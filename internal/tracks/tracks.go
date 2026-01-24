package tracks

import (
	"fmt"
	"strings"

	"github.com/asticode/go-astiav"
)

// TrackType represents the type of media track
type TrackType string

const (
	TrackTypeAudio    TrackType = "audio"
	TrackTypeSubtitle TrackType = "subtitle"
)

// Track represents a single audio or subtitle track
type Track struct {
	Index    int       // Stream index in the file (FFmpeg absolute index)
	MpvIndex int       // MPV track index (position within track type: aid=1,2,3 or sid=1,2,3)
	Type     TrackType // audio or subtitle
	Codec    string    // Codec name (e.g., "aac", "ass", "subrip")
	Language string    // Language code (e.g., "eng", "jpn", "fre")
	Title    string    // Human-readable title (e.g., "English", "Japanese", "English[CC]")
}

// FileTrackInfo contains track information for a single file
type FileTrackInfo struct {
	Filename       string   // Path to the file
	AudioTracks    []Track  // All audio tracks
	SubtitleTracks []Track  // All subtitle tracks
	Error          error    // Error if probing failed
}

// TrackSummary represents a unique track configuration
type TrackSummary struct {
	Type     TrackType
	Language string
	Title    string
}

// TrackAnalysis contains aggregated analysis across multiple files
type TrackAnalysis struct {
	Files []FileTrackInfo // Individual file track info

	// Aggregated information
	AllAudioLanguages     []string       // All unique audio languages found
	AllSubtitleLanguages  []string       // All unique subtitle languages found
	CommonAudioLanguages  []string       // Audio languages present in ALL files
	CommonSubtitleLanguages []string     // Subtitle languages present in ALL files

	// Detailed track summaries
	AllAudioTracks        []TrackSummary // All unique audio tracks
	AllSubtitleTracks     []TrackSummary // All unique subtitle tracks
	CommonAudioTracks     []TrackSummary // Audio tracks in ALL files
	CommonSubtitleTracks  []TrackSummary // Subtitle tracks in ALL files
}

// ProbeFile analyzes a single media file and extracts audio and subtitle track information
func ProbeFile(filePath string) (*FileTrackInfo, error) {
	info := &FileTrackInfo{
		Filename:       filePath,
		AudioTracks:    []Track{},
		SubtitleTracks: []Track{},
	}

	// Open the input file
	inputFormatContext := astiav.AllocFormatContext()
	if inputFormatContext == nil {
		return info, fmt.Errorf("failed to allocate format context")
	}
	defer inputFormatContext.Free()

	if err := inputFormatContext.OpenInput(filePath, nil, nil); err != nil {
		return info, fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFormatContext.CloseInput()

	if err := inputFormatContext.FindStreamInfo(nil); err != nil {
		return info, fmt.Errorf("failed to find stream info: %w", err)
	}

	// Iterate through all streams
	audioTrackCount := 0
	subtitleTrackCount := 0

	for _, stream := range inputFormatContext.Streams() {
		codecParams := stream.CodecParameters()
		mediaType := codecParams.MediaType()

		// Only process audio and subtitle streams
		if mediaType != astiav.MediaTypeAudio && mediaType != astiav.MediaTypeSubtitle {
			continue
		}

		track := Track{
			Index: stream.Index(),
			Codec: codecParams.CodecID().String(),
		}

		// Determine track type and MPV index
		if mediaType == astiav.MediaTypeAudio {
			track.Type = TrackTypeAudio
			audioTrackCount++
			track.MpvIndex = audioTrackCount
		} else {
			track.Type = TrackTypeSubtitle
			subtitleTrackCount++
			track.MpvIndex = subtitleTrackCount
		}

		// Extract language and title from metadata
		metadata := stream.Metadata()
		if metadata != nil {
			// Get language
			if langEntry := metadata.Get("language", nil, 0); langEntry != nil {
				track.Language = langEntry.Value()
			}

			// Get title
			if titleEntry := metadata.Get("title", nil, 0); titleEntry != nil {
				track.Title = titleEntry.Value()
			}
		}

		// Add track to appropriate list
		if track.Type == TrackTypeAudio {
			info.AudioTracks = append(info.AudioTracks, track)
		} else {
			info.SubtitleTracks = append(info.SubtitleTracks, track)
		}
	}

	return info, nil
}

// AnalyzeFiles analyzes multiple media files and produces an aggregated analysis
func AnalyzeFiles(filePaths []string) *TrackAnalysis {
	analysis := &TrackAnalysis{
		Files:                    []FileTrackInfo{},
		AllAudioLanguages:        []string{},
		AllSubtitleLanguages:     []string{},
		CommonAudioLanguages:     []string{},
		CommonSubtitleLanguages:  []string{},
		AllAudioTracks:           []TrackSummary{},
		AllSubtitleTracks:        []TrackSummary{},
		CommonAudioTracks:        []TrackSummary{},
		CommonSubtitleTracks:     []TrackSummary{},
	}

	// Probe all files
	for _, path := range filePaths {
		fileInfo, err := ProbeFile(path)
		if err != nil {
			fileInfo.Error = err
		}
		analysis.Files = append(analysis.Files, *fileInfo)
	}

	// Calculate aggregated information
	audioLangCounts := make(map[string]int)
	subtitleLangCounts := make(map[string]int)
	audioTrackCounts := make(map[TrackSummary]int)
	subtitleTrackCounts := make(map[TrackSummary]int)

	validFileCount := 0
	for _, fileInfo := range analysis.Files {
		if fileInfo.Error != nil {
			continue
		}
		validFileCount++

		// Track audio languages
		audioLangs := make(map[string]bool)
		for _, track := range fileInfo.AudioTracks {
			effectiveLang := track.GetEffectiveLanguage()
			if effectiveLang != "" {
				audioLangs[effectiveLang] = true

				summary := TrackSummary{
					Type:     TrackTypeAudio,
					Language: effectiveLang,
					Title:    track.Title,
				}
				audioTrackCounts[summary]++
			}
		}
		for lang := range audioLangs {
			audioLangCounts[lang]++
		}

		// Track subtitle languages
		subtitleLangs := make(map[string]bool)
		for _, track := range fileInfo.SubtitleTracks {
			effectiveLang := track.GetEffectiveLanguage()
			if effectiveLang != "" {
				subtitleLangs[effectiveLang] = true

				summary := TrackSummary{
					Type:     TrackTypeSubtitle,
					Language: effectiveLang,
					Title:    track.Title,
				}
				subtitleTrackCounts[summary]++
			}
		}
		for lang := range subtitleLangs {
			subtitleLangCounts[lang]++
		}
	}

	// Extract all unique languages
	for lang := range audioLangCounts {
		analysis.AllAudioLanguages = append(analysis.AllAudioLanguages, lang)
		if audioLangCounts[lang] == validFileCount {
			analysis.CommonAudioLanguages = append(analysis.CommonAudioLanguages, lang)
		}
	}

	for lang := range subtitleLangCounts {
		analysis.AllSubtitleLanguages = append(analysis.AllSubtitleLanguages, lang)
		if subtitleLangCounts[lang] == validFileCount {
			analysis.CommonSubtitleLanguages = append(analysis.CommonSubtitleLanguages, lang)
		}
	}

	// Extract all unique tracks
	for summary := range audioTrackCounts {
		analysis.AllAudioTracks = append(analysis.AllAudioTracks, summary)
		if audioTrackCounts[summary] == validFileCount {
			analysis.CommonAudioTracks = append(analysis.CommonAudioTracks, summary)
		}
	}

	for summary := range subtitleTrackCounts {
		analysis.AllSubtitleTracks = append(analysis.AllSubtitleTracks, summary)
		if subtitleTrackCounts[summary] == validFileCount {
			analysis.CommonSubtitleTracks = append(analysis.CommonSubtitleTracks, summary)
		}
	}

	return analysis
}

// HasLanguage checks if a FileTrackInfo contains a specific language in audio or subtitle tracks
func (f *FileTrackInfo) HasLanguage(trackType TrackType, language string) bool {
	tracks := f.AudioTracks
	if trackType == TrackTypeSubtitle {
		tracks = f.SubtitleTracks
	}

	for _, track := range tracks {
		if track.GetEffectiveLanguage() == language {
			return true
		}
	}
	return false
}

// GetTracksByLanguage returns all tracks matching a specific language
func (f *FileTrackInfo) GetTracksByLanguage(trackType TrackType, language string) []Track {
	var result []Track
	tracks := f.AudioTracks
	if trackType == TrackTypeSubtitle {
		tracks = f.SubtitleTracks
	}

	for _, track := range tracks {
		if track.GetEffectiveLanguage() == language {
			result = append(result, track)
		}
	}
	return result
}

// detectLanguageFromTitle attempts to detect the actual language from a track title
// when metadata is incorrect (common with fan-encoded anime)
func detectLanguageFromTitle(title string) string {
	titleLower := strings.ToLower(title)

	// Check for Japanese indicators
	if strings.Contains(titleLower, "japanese") ||
	   strings.Contains(titleLower, "jpn") ||
	   strings.Contains(titleLower, "日本語") {
		return "jpn"
	}

	// Check for English indicators
	if strings.Contains(titleLower, "english") ||
	   strings.Contains(titleLower, "eng") {
		return "eng"
	}

	return ""
}

// GetEffectiveLanguage returns the actual language, checking title if metadata is incorrect
func (t *Track) GetEffectiveLanguage() string {
	// If we have a language code, try to detect from title first
	if t.Title != "" {
		detectedLang := detectLanguageFromTitle(t.Title)
		if detectedLang != "" {
			return detectedLang
		}
	}

	// Fall back to metadata language
	return t.Language
}
