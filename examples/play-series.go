package main

import (
	"flag"
	"fmt"
	"os"

	"mpv-launcher/internal/episode"
	"mpv-launcher/internal/player"
	"mpv-launcher/internal/scanner"
	"mpv-launcher/internal/tracks"
)

func main() {
	// Command-line flags
	dirPath := flag.String("dir", "", "Directory containing video files")
	recursive := flag.Bool("recursive", true, "Scan subdirectories recursively")
	audioLang := flag.String("audio", "", "Preferred audio language (e.g., eng, jpn)")
	subLang := flag.String("sub", "", "Preferred subtitle language (e.g., eng, fre)")
	noSub := flag.Bool("no-sub", false, "Disable subtitles")
	fullscreen := flag.Bool("fullscreen", false, "Start in fullscreen mode")
	flag.Parse()

	if *dirPath == "" {
		fmt.Println("Usage: play-series -dir /path/to/series [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

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
		fmt.Println("No video files found")
		os.Exit(0)
	}

	fmt.Printf("Found %d video files\n", len(videos))

	// Step 2: Parse and sort episodes
	fmt.Println("\nParsing and sorting episodes...")
	episodes := episode.ParseBatch(videos)
	episode.Sort(episodes)

	// Display episode order
	fmt.Println("\nPlaylist order:")
	for i, ep := range episodes {
		if ep.Type == episode.TypeSeasonEpisode {
			if ep.Part != "" {
				fmt.Printf("  %d. S%02dE%02d Part %s\n", i+1, ep.Season, ep.Episodes[0], ep.Part)
			} else if len(ep.Episodes) > 1 {
				fmt.Printf("  %d. S%02dE%02d-%02d\n", i+1, ep.Season, ep.Episodes[0], ep.Episodes[len(ep.Episodes)-1])
			} else {
				fmt.Printf("  %d. S%02dE%02d\n", i+1, ep.Season, ep.Episodes[0])
			}
		} else if ep.Type == episode.TypeAbsolute {
			fmt.Printf("  %d. Episode %d\n", i+1, ep.AbsoluteEp)
		} else if ep.Type == episode.TypeDate {
			fmt.Printf("  %d. %s\n", i+1, ep.Date.Format("2006-01-02"))
		} else {
			fmt.Printf("  %d. %s\n", i+1, ep.Filename)
		}
	}

	// Step 3: Analyze tracks (if language preferences not specified)
	var analysis *tracks.TrackAnalysis
	if *audioLang == "" && *subLang == "" && !*noSub {
		fmt.Println("\nAnalyzing audio and subtitle tracks...")
		analysis = tracks.AnalyzeFiles(videos)

		if len(analysis.CommonAudioLanguages) > 0 {
			fmt.Printf("\nCommon audio languages: %v\n", analysis.CommonAudioLanguages)
		}
		if len(analysis.CommonSubtitleLanguages) > 0 {
			fmt.Printf("Common subtitle languages: %v\n", analysis.CommonSubtitleLanguages)
		}

		// Auto-select first common language if available
		if *audioLang == "" && len(analysis.CommonAudioLanguages) > 0 {
			*audioLang = analysis.CommonAudioLanguages[0]
			fmt.Printf("\nAuto-selected audio language: %s\n", *audioLang)
		}
	}

	// Step 4: Build playlist with ordered filenames
	playlist := make([]string, len(episodes))
	for i, ep := range episodes {
		playlist[i] = ep.Filename
	}

	// Step 5: Configure playback options
	playlistOpts := player.PlaylistOptions{
		AudioLanguage:    *audioLang,
		SubtitleLanguage: *subLang,
		Fullscreen:       *fullscreen,
	}

	if *noSub {
		playlistOpts.SubtitleTrack = -1
	}

	// Step 6: Launch MPV
	fmt.Printf("\nLaunching MPV with %d files...\n", len(playlist))
	if playlistOpts.AudioLanguage != "" {
		fmt.Printf("Audio language: %s\n", playlistOpts.AudioLanguage)
	}
	if playlistOpts.SubtitleTrack == -1 {
		fmt.Println("Subtitles: disabled")
	} else if playlistOpts.SubtitleLanguage != "" {
		fmt.Printf("Subtitle language: %s\n", playlistOpts.SubtitleLanguage)
	}

	mpvPlayer, err := player.NewMPVPlayer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating player: %v\n", err)
		os.Exit(1)
	}

	if err := mpvPlayer.PlayPlaylist(playlist, playlistOpts); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting playback: %v\n", err)
		os.Exit(1)
	}

	// Wait for playback to complete
	if err := mpvPlayer.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Playback error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nPlayback completed")
}
