package tracks

// FileProber abstracts media file probing for testing
type FileProber interface {
	ProbeFile(filePath string) (*FileTrackInfo, error)
}

// DefaultFileProber uses real FFmpeg via go-astiav
type DefaultFileProber struct{}

// ProbeFile implements FileProber using the real FFmpeg-based ProbeFile function
func (p *DefaultFileProber) ProbeFile(filePath string) (*FileTrackInfo, error) {
	return ProbeFile(filePath)
}

// AnalyzeFilesWithProber analyzes multiple media files using the provided prober
// This allows injecting a custom prober for testing without real media files
func AnalyzeFilesWithProber(filePaths []string, prober FileProber) *TrackAnalysis {
	analysis := &TrackAnalysis{
		Files:                   []FileTrackInfo{},
		AllAudioLanguages:       []string{},
		AllSubtitleLanguages:    []string{},
		CommonAudioLanguages:    []string{},
		CommonSubtitleLanguages: []string{},
		AllAudioTracks:          []TrackSummary{},
		AllSubtitleTracks:       []TrackSummary{},
		CommonAudioTracks:       []TrackSummary{},
		CommonSubtitleTracks:    []TrackSummary{},
	}

	// Probe all files using the provided prober
	for _, path := range filePaths {
		fileInfo, err := prober.ProbeFile(path)
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
