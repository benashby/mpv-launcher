package probers

import (
	"fmt"
	"sync"

	"mpv-launcher/internal/ffmpeg"
	"mpv-launcher/internal/testutil"
)

// FixtureVideoProber implements ffmpeg.VideoProber using JSON fixtures
type FixtureVideoProber struct {
	mu       sync.RWMutex
	fixtures map[string]*testutil.FFProbeOutput // virtualPath -> fixture data
}

// NewFixtureVideoProber creates a new fixture-based video prober
func NewFixtureVideoProber() *FixtureVideoProber {
	return &FixtureVideoProber{
		fixtures: make(map[string]*testutil.FFProbeOutput),
	}
}

// LoadFixture associates an ffprobe JSON fixture with a virtual file path
func (p *FixtureVideoProber) LoadFixture(virtualPath, fixturePath string) error {
	output, err := testutil.LoadFFProbeJSON(fixturePath)
	if err != nil {
		return fmt.Errorf("failed to load fixture %s: %w", fixturePath, err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.fixtures[virtualPath] = output
	return nil
}

// LoadFixtureData associates pre-parsed FFProbeOutput with a virtual file path
func (p *FixtureVideoProber) LoadFixtureData(virtualPath string, data *testutil.FFProbeOutput) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fixtures[virtualPath] = data
}

// ProbeVideo implements ffmpeg.VideoProber
func (p *FixtureVideoProber) ProbeVideo(filePath string) (*ffmpeg.VideoInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	fixture, ok := p.fixtures[filePath]
	if !ok {
		return nil, fmt.Errorf("fixture not found for path: %s", filePath)
	}

	return convertToFFmpegVideoInfo(fixture)
}

// convertToFFmpegVideoInfo converts FFProbeOutput to ffmpeg.VideoInfo
func convertToFFmpegVideoInfo(o *testutil.FFProbeOutput) (*ffmpeg.VideoInfo, error) {
	vi, err := o.ToVideoInfo()
	if err != nil {
		return nil, err
	}

	return &ffmpeg.VideoInfo{
		Duration:  vi.Duration,
		Width:     vi.Width,
		Height:    vi.Height,
		Codec:     vi.Codec,
		FrameRate: vi.FrameRate,
	}, nil
}
