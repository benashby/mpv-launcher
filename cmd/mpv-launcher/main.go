package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mpv-launcher/internal/episode"
	"mpv-launcher/internal/hyprland"
	"mpv-launcher/internal/player"
	"mpv-launcher/internal/playlist"
	"mpv-launcher/internal/priority"
	"mpv-launcher/internal/scanner"
	"mpv-launcher/internal/tracks"
	"mpv-launcher/internal/ui"
)

const maxFiles = 100

func main() {
	// Parse command-line flags
	dirPath := flag.String("dir", ".", "Directory to scan for videos (default: current directory)")
	recursive := flag.Bool("r", false, "Recursively scan subdirectories")
	flag.Parse()

	// Step 1: Scan directory for video files
	fmt.Printf("Scanning directory: %s\n", *dirPath)
	opts := scanner.ScanOptions{
		Recursive:      *recursive,
		FollowSymlinks: false,
	}

	videos, err := scanner.ScanDirectory(*dirPath, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(videos) == 0 {
		fmt.Println("No video files found in directory")
		os.Exit(1)
	}

	// Apply hard cap on file count
	if len(videos) > maxFiles {
		fmt.Printf("Warning: Found %d files, limiting to %d\n", len(videos), maxFiles)
		videos = videos[:maxFiles]
	}

	fmt.Printf("Found %d video files\n", len(videos))

	// Step 2: Parse and sort episodes (specials now sort to end)
	episodes := episode.ParseBatch(videos)
	episode.Sort(episodes)

	// Build ordered playlist
	filenames := make([]string, len(episodes))
	for i, ep := range episodes {
		filenames[i] = ep.Filename
	}

	// Step 3: Analyze tracks for all files
	fmt.Println("Analyzing audio and subtitle tracks...")
	analysis := tracks.AnalyzeFiles(filenames)

	// Check for any valid files
	validFiles := []tracks.FileTrackInfo{}
	for _, file := range analysis.Files {
		if file.Error == nil {
			validFiles = append(validFiles, file)
		}
	}

	if len(validFiles) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No valid files to analyze\n")
		os.Exit(1)
	}

	// Step 3.5: Monitor selection (Hyprland only)
	playOpts := player.PlaylistOptions{}

	if hyprland.IsAvailable() {
		monitors, err := hyprland.GetMonitors()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not query Hyprland monitors: %v\n", err)
		} else {
			items := make([]string, 1+len(monitors))
			items[0] = "Normal window"
			for i, m := range monitors {
				label := fmt.Sprintf("%s - %dx%d @ %.0fHz", m.Name, m.Width, m.Height, m.RefreshRate)
				if m.Focused {
					label += " [focused]"
				}
				items[i+1] = label
			}
			fmt.Println()
			choice, err := ui.CursorSelect("Select monitor:", items)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error in monitor selector: %v\n", err)
				os.Exit(1)
			}
			if choice == -1 {
				fmt.Println("Selection canceled")
				os.Exit(0)
			}
			if choice > 0 {
				playOpts.ScreenName = monitors[choice-1].Name
			}
		}
	}

	// Step 4: Get default permutations and show priority UI
	perms := priority.DefaultPermutations()

	// Build labels for the UI
	labels := make([]string, len(perms))
	for i, p := range perms {
		labels[i] = p.Label
	}

	// Show priority selector
	fmt.Println()
	selector := ui.NewPrioritySelector(labels)
	order, err := selector.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in priority selector: %v\n", err)
		os.Exit(1)
	}

	if selector.Canceled() {
		fmt.Println("Selection canceled")
		os.Exit(0)
	}

	// Reorder permutations based on user selection
	orderedPerms := make([]priority.Permutation, len(order))
	for i, idx := range order {
		orderedPerms[i] = perms[idx]
	}

	// Step 5: Build per-file playlist
	builder := playlist.NewBuilder(orderedPerms, validFiles)
	playbackList := builder.Build()

	// Show summary
	summary := builder.Summary(playbackList)
	fmt.Println("Track selection summary:")
	for label, count := range summary {
		fmt.Printf("  %s: %d files\n", label, count)
	}
	fmt.Println()

	// Step 6: Launch MPV with per-file tracks
	fmt.Printf("Launching MPV with %d files...\n\n", len(playbackList))

	mpvPlayer, err := player.NewMPVPlayer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating player: %v\n", err)
		os.Exit(1)
	}

	if err := mpvPlayer.PlaylistWithPerFileTracks(playbackList, playOpts); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting playback: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nStopping player...")
		mpvPlayer.Stop()
		os.Exit(0)
	}()

	// Wait for playback to complete
	if err := mpvPlayer.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Playback error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nPlayback finished")
}
