package ffmpeg

// VideoProber abstracts video file probing for testing
type VideoProber interface {
	ProbeVideo(filePath string) (*VideoInfo, error)
}

// DefaultVideoProber uses real FFmpeg via go-astiav
type DefaultVideoProber struct{}

// ProbeVideo implements VideoProber using the real FFmpeg-based ProbeVideo function
func (p *DefaultVideoProber) ProbeVideo(filePath string) (*VideoInfo, error) {
	return ProbeVideo(filePath)
}
