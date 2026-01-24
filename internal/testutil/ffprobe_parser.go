package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadFFProbeJSON loads and parses an ffprobe JSON fixture from the given path
func LoadFFProbeJSON(path string) (*FFProbeOutput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file: %w", err)
	}

	var output FFProbeOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe JSON: %w", err)
	}

	return &output, nil
}

// VideoInfo contains basic information about a video extracted from fixture
type VideoInfo struct {
	Duration  int64
	Width     int
	Height    int
	Codec     string
	FrameRate float64
}

// ToVideoInfo converts FFProbeOutput to a VideoInfo
func (o *FFProbeOutput) ToVideoInfo() (*VideoInfo, error) {
	// Find the video stream
	var videoStream *FFProbeStream
	for i := range o.Streams {
		if o.Streams[i].CodecType == "video" {
			videoStream = &o.Streams[i]
			break
		}
	}

	if videoStream == nil {
		return nil, fmt.Errorf("no video stream found in fixture")
	}

	info := &VideoInfo{
		Width:     videoStream.Width,
		Height:    videoStream.Height,
		Codec:     MapCodecName(videoStream.CodecName),
		FrameRate: ParseFrameRate(videoStream.AvgFrameRate),
	}

	// Parse duration from format (in seconds as string) or stream
	if o.Format.Duration != "" {
		duration, err := strconv.ParseFloat(o.Format.Duration, 64)
		if err == nil {
			// Convert seconds to microseconds (FFmpeg's time_base)
			info.Duration = int64(duration * 1000000)
		}
	}

	return info, nil
}

// StreamInfo contains extracted track information
type StreamInfo struct {
	Index    int
	MpvIndex int
	Type     string // "audio" or "subtitle"
	Codec    string
	Language string
	Title    string
}

// FileTrackInfo contains track information for a single file
type FileTrackInfo struct {
	Filename       string
	AudioTracks    []StreamInfo
	SubtitleTracks []StreamInfo
}

// ToFileTrackInfo converts FFProbeOutput to a FileTrackInfo
func (o *FFProbeOutput) ToFileTrackInfo(filePath string) *FileTrackInfo {
	info := &FileTrackInfo{
		Filename:       filePath,
		AudioTracks:    []StreamInfo{},
		SubtitleTracks: []StreamInfo{},
	}

	audioTrackCount := 0
	subtitleTrackCount := 0

	for _, stream := range o.Streams {
		switch stream.CodecType {
		case "audio":
			audioTrackCount++
			track := StreamInfo{
				Index:    stream.Index,
				MpvIndex: audioTrackCount,
				Type:     "audio",
				Codec:    MapCodecName(stream.CodecName),
				Language: stream.Tags.Language,
				Title:    stream.Tags.Title,
			}
			info.AudioTracks = append(info.AudioTracks, track)

		case "subtitle":
			subtitleTrackCount++
			track := StreamInfo{
				Index:    stream.Index,
				MpvIndex: subtitleTrackCount,
				Type:     "subtitle",
				Codec:    MapCodecName(stream.CodecName),
				Language: stream.Tags.Language,
				Title:    stream.Tags.Title,
			}
			info.SubtitleTracks = append(info.SubtitleTracks, track)
		}
	}

	return info
}

// ParseFrameRate parses a frame rate string in "num/den" format
func ParseFrameRate(s string) float64 {
	if s == "" {
		return 0
	}

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		// Try parsing as a simple float
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}

	num, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}

	den, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || den == 0 {
		return 0
	}

	return num / den
}

// MapCodecName maps ffprobe codec names to go-astiav style codec ID strings
// The actual FFmpeg library returns codec IDs like "aac", "hevc", etc.
// which are similar but not always identical to ffprobe's codec_name
func MapCodecName(codecName string) string {
	// Map common codec names to their FFmpeg codec ID representations
	codecMap := map[string]string{
		"h264":              "h264",
		"hevc":              "hevc",
		"h265":              "hevc",
		"vp9":               "vp9",
		"av1":               "av1",
		"aac":               "aac",
		"ac3":               "ac3",
		"eac3":              "eac3",
		"dts":               "dts",
		"flac":              "flac",
		"mp3":               "mp3",
		"opus":              "opus",
		"vorbis":            "vorbis",
		"truehd":            "truehd",
		"pcm_s16le":         "pcm_s16le",
		"pcm_s24le":         "pcm_s24le",
		"ass":               "ass",
		"ssa":               "ssa",
		"subrip":            "subrip",
		"srt":               "subrip",
		"hdmv_pgs_subtitle": "hdmv_pgs_subtitle",
		"pgssub":            "hdmv_pgs_subtitle",
		"dvd_subtitle":      "dvd_subtitle",
		"dvdsub":            "dvd_subtitle",
		"webvtt":            "webvtt",
		"mov_text":          "mov_text",
	}

	if mapped, ok := codecMap[codecName]; ok {
		return mapped
	}
	return codecName
}
