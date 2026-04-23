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
	Fullscreen bool   // Start in fullscreen mode
	ScreenName string // Hyprland wl_output name; sets --fs --fs-screen-name=<name>

	// mpv.conf named profile to activate (e.g. "anime", "music")
	Profile string

	// Pass --deband=no to override mpv.conf deband setting
	NoDeband bool

	// Hardware decode mode override; passed as --hwdec=<mode> when non-empty
	HwdecMode string

	// Unix socket path for mpv IPC; passed as --input-ipc-server=<path> when non-empty
	IPCSocket string
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

// buildPerFileArgs constructs the mpv argument list for a per-file playlist.
// Exported as an internal helper so it can be tested without launching a process.
func buildPerFileArgs(files []playlist.FilePlayback, opts PlaylistOptions) []string {
	args := []string{}

	// Global options that apply to the whole session.
	if opts.Profile != "" {
		args = append(args, "--profile="+opts.Profile)
	}
	if opts.NoDeband {
		args = append(args, "--deband=no")
	}
	if opts.HwdecMode != "" {
		args = append(args, "--hwdec="+opts.HwdecMode)
	}
	if opts.IPCSocket != "" {
		args = append(args, "--input-ipc-server="+opts.IPCSocket)
	}

	// Monitor targeting: fullscreen on a specific output, aspect ratio always preserved.
	// Plain fullscreen is a fallback when no monitor is named.
	if opts.ScreenName != "" {
		args = append(args, "--fs", "--fs-screen-name="+opts.ScreenName, "--keepaspect=yes")
	} else if opts.Fullscreen {
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

		args = append(args, file.Filename)
		args = append(args, "--}")
	}

	return args
}

// PlaylistWithPerFileTracks launches MPV with different tracks per file.
// Uses MPV's --{ and --} for per-file options:
// mpv --{ --aid=1 --sid=1 file1.mkv --} --{ --aid=2 --sid=no file2.mkv --}
func (p *MPVPlayer) PlaylistWithPerFileTracks(files []playlist.FilePlayback, opts PlaylistOptions) error {
	if len(files) == 0 {
		return fmt.Errorf("playlist is empty")
	}

	args := buildPerFileArgs(files, opts)

	p.cmd = exec.Command("mpv", args...)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	return nil
}
