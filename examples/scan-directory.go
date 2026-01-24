package main

import (
	"fmt"
	"os"

	"mpv-launcher/internal/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <directory> [--recursive]\n", os.Args[0])
		os.Exit(1)
	}

	directory := os.Args[1]
	recursive := false

	// Check for --recursive flag
	if len(os.Args) > 2 && os.Args[2] == "--recursive" {
		recursive = true
	}

	// Configure scan options
	opts := scanner.ScanOptions{
		Recursive:      recursive,
		FollowSymlinks: false,
	}

	// Scan for video files
	fmt.Printf("Scanning %s for video files", directory)
	if recursive {
		fmt.Print(" (recursive)")
	}
	fmt.Println("...")

	videos, err := scanner.ScanDirectory(directory, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Printf("\nFound %d video file(s):\n\n", len(videos))
	for i, video := range videos {
		fmt.Printf("%d. %s\n", i+1, video)
	}
}
