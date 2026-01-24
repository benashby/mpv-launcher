package ffmpeg

import (
	"fmt"

	"github.com/asticode/go-astiav"
)

// VideoInfo contains basic information about a video file
type VideoInfo struct {
	Duration  int64
	Width     int
	Height    int
	Codec     string
	FrameRate float64
}

// ProbeVideo analyzes a video file using FFmpeg and returns its information
func ProbeVideo(filePath string) (*VideoInfo, error) {
	// Open input file
	inputFormatContext := astiav.AllocFormatContext()
	if inputFormatContext == nil {
		return nil, fmt.Errorf("failed to allocate format context")
	}
	defer inputFormatContext.Free()

	if err := inputFormatContext.OpenInput(filePath, nil, nil); err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFormatContext.CloseInput()

	if err := inputFormatContext.FindStreamInfo(nil); err != nil {
		return nil, fmt.Errorf("failed to find stream info: %w", err)
	}

	// Find video stream
	videoStream, _, err := inputFormatContext.FindBestStream(astiav.MediaTypeVideo, -1, -1)
	if err != nil {
		return nil, fmt.Errorf("no video stream found: %w", err)
	}

	codecParams := videoStream.CodecParameters()
	info := &VideoInfo{
		Duration:  inputFormatContext.Duration(),
		Width:     codecParams.Width(),
		Height:    codecParams.Height(),
		Codec:     codecParams.CodecID().String(),
	}

	// Calculate frame rate
	if videoStream.AvgFrameRate().Num() > 0 {
		info.FrameRate = float64(videoStream.AvgFrameRate().Num()) / float64(videoStream.AvgFrameRate().Den())
	}

	return info, nil
}
