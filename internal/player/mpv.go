package player

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"mpv-launcher/internal/playlist"
)

// MPVPlayer manages MPV playback
type MPVPlayer struct {
	cmd *exec.Cmd
}

// PlaylistOptions configures playlist playback with track selection
type PlaylistOptions struct {
	// Track selection by index (from FFmpeg stream index)
	AudioTrack    int // Specific audio track index, 0 = auto-select
	SubtitleTrack int // Specific subtitle track index, 0 = auto-select, -1 = disable

	// Track selection by language preference
	AudioLanguage    string // Preferred audio language (e.g., "eng", "jpn")
	SubtitleLanguage string // Preferred subtitle language (e.g., "eng", "fre")

	// Playback options
	Fullscreen bool // Start in fullscreen mode
}

// NewMPVPlayer creates a new MPV player instance
func NewMPVPlayer() (*MPVPlayer, error) {
	return &MPVPlayer{}, nil
}

// Play starts MPV with the given video file
func (p *MPVPlayer) Play(videoPath string) error {
	// Start MPV
	p.cmd = exec.Command("mpv", videoPath)

	// Connect stdout/stderr for output
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	return nil
}

// PlayPlaylist starts MPV with a playlist of files and track selection options
func (p *MPVPlayer) PlayPlaylist(files []string, opts PlaylistOptions) error {
	if len(files) == 0 {
		return fmt.Errorf("playlist is empty")
	}

	// Build MPV command arguments
	args := []string{}

	// Audio track selection
	if opts.AudioTrack > 0 {
		// Use specific audio track index
		args = append(args, "--aid="+strconv.Itoa(opts.AudioTrack))
	} else if opts.AudioLanguage != "" {
		// Use language preference
		args = append(args, "--alang="+opts.AudioLanguage)
	}

	// Subtitle track selection
	if opts.SubtitleTrack == -1 {
		// Disable subtitles
		args = append(args, "--sid=no")
	} else if opts.SubtitleTrack > 0 {
		// Use specific subtitle track index
		args = append(args, "--sid="+strconv.Itoa(opts.SubtitleTrack))
	} else if opts.SubtitleLanguage != "" {
		// Use language preference
		args = append(args, "--slang="+opts.SubtitleLanguage)
	}

	// Fullscreen option
	if opts.Fullscreen {
		args = append(args, "--fs")
	}

	// Add all files to the playlist
	args = append(args, files...)

	// Create and start MPV command
	p.cmd = exec.Command("mpv", args...)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	return nil
}

// Wait waits for MPV to exit
func (p *MPVPlayer) Wait() error {
	if p.cmd == nil {
		return fmt.Errorf("mpv not started")
	}
	return p.cmd.Wait()
}

// Stop stops MPV playback
func (p *MPVPlayer) Stop() error {
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

// PlaylistWithPerFileTracks launches MPV with different tracks per file
// Uses MPV's --{ and --} for per-file options:
// mpv --{ --aid=1 --sid=1 file1.mkv --} --{ --aid=2 --sid=no file2.mkv --}
func (p *MPVPlayer) PlaylistWithPerFileTracks(files []playlist.FilePlayback, fullscreen bool) error {
	if len(files) == 0 {
		return fmt.Errorf("playlist is empty")
	}

	args := []string{}

	// Add fullscreen at the start (global option)
	if fullscreen {
		args = append(args, "--fs")
	}

	// Build per-file arguments using --{ and --}
	for _, file := range files {
		args = append(args, "--{")

		// Audio track
		if file.AudioTrack > 0 {
			args = append(args, "--aid="+strconv.Itoa(file.AudioTrack))
		}

		// Subtitle track
		if file.SubtitleTrack == -1 {
			args = append(args, "--sid=no")
		} else if file.SubtitleTrack > 0 {
			args = append(args, "--sid="+strconv.Itoa(file.SubtitleTrack))
		}

		// Add the file
		args = append(args, file.Filename)

		args = append(args, "--}")
	}

	// Create and start MPV command
	p.cmd = exec.Command("mpv", args...)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	return nil
}
