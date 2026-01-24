package main

import (
	"fmt"
	"os"

	"mpv-launcher/internal/tracks"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <video-file>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]

	info, err := tracks.ProbeFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error probing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File: %s\n\n", info.Filename)

	fmt.Printf("Audio tracks (%d):\n", len(info.AudioTracks))
	for _, track := range info.AudioTracks {
		fmt.Printf("  FFmpeg[%d] MPV[aid=%d] %s - %s (%s)\n",
			track.Index, track.MpvIndex, track.Language, track.Title, track.Codec)
	}

	fmt.Printf("\nSubtitle tracks (%d):\n", len(info.SubtitleTracks))
	for _, track := range info.SubtitleTracks {
		fmt.Printf("  FFmpeg[%d] MPV[sid=%d] %s - %s (%s)\n",
			track.Index, track.MpvIndex, track.Language, track.Title, track.Codec)
	}
}
