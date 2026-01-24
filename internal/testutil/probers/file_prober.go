// Package probers provides fixture-based implementations of prober interfaces
// for testing without real media files.
package probers

import (
	"fmt"
	"sync"

	"mpv-launcher/internal/testutil"
	"mpv-launcher/internal/tracks"
)

// FixtureFileProber implements tracks.FileProber using JSON fixtures
type FixtureFileProber struct {
	mu       sync.RWMutex
	fixtures map[string]*testutil.FFProbeOutput // virtualPath -> fixture data
}

// NewFixtureFileProber creates a new fixture-based file prober
func NewFixtureFileProber() *FixtureFileProber {
	return &FixtureFileProber{
		fixtures: make(map[string]*testutil.FFProbeOutput),
	}
}

// LoadFixture associates an ffprobe JSON fixture with a virtual file path
func (p *FixtureFileProber) LoadFixture(virtualPath, fixturePath string) error {
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
func (p *FixtureFileProber) LoadFixtureData(virtualPath string, data *testutil.FFProbeOutput) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fixtures[virtualPath] = data
}

// ProbeFile implements tracks.FileProber
func (p *FixtureFileProber) ProbeFile(filePath string) (*tracks.FileTrackInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	fixture, ok := p.fixtures[filePath]
	if !ok {
		// Return error similar to what real probing would return
		return &tracks.FileTrackInfo{
			Filename:       filePath,
			AudioTracks:    []tracks.Track{},
			SubtitleTracks: []tracks.Track{},
		}, fmt.Errorf("fixture not found for path: %s", filePath)
	}

	return convertToTracksFileTrackInfo(fixture, filePath), nil
}

// convertToTracksFileTrackInfo converts FFProbeOutput to tracks.FileTrackInfo
func convertToTracksFileTrackInfo(o *testutil.FFProbeOutput, filePath string) *tracks.FileTrackInfo {
	info := &tracks.FileTrackInfo{
		Filename:       filePath,
		AudioTracks:    []tracks.Track{},
		SubtitleTracks: []tracks.Track{},
	}

	audioTrackCount := 0
	subtitleTrackCount := 0

	for _, stream := range o.Streams {
		switch stream.CodecType {
		case "audio":
			audioTrackCount++
			track := tracks.Track{
				Index:    stream.Index,
				MpvIndex: audioTrackCount,
				Type:     tracks.TrackTypeAudio,
				Codec:    testutil.MapCodecName(stream.CodecName),
				Language: stream.Tags.Language,
				Title:    stream.Tags.Title,
			}
			info.AudioTracks = append(info.AudioTracks, track)

		case "subtitle":
			subtitleTrackCount++
			track := tracks.Track{
				Index:    stream.Index,
				MpvIndex: subtitleTrackCount,
				Type:     tracks.TrackTypeSubtitle,
				Codec:    testutil.MapCodecName(stream.CodecName),
				Language: stream.Tags.Language,
				Title:    stream.Tags.Title,
			}
			info.SubtitleTracks = append(info.SubtitleTracks, track)
		}
	}

	return info
}
