# MPV Launcher

A Go application that uses FFmpeg for video analysis and preprocessing, then launches MPV for playback.

## Features

- Video analysis using FFmpeg (via go-astiav CGO bindings)
- MPV playback control
- XDG-compatible installation to `~/.local/bin`
- Graceful shutdown handling
- **Video file scanner** - Recursively or non-recursively scan directories for video files
  - Supports 40+ common and uncommon video file extensions
  - Configurable recursive scanning
  - Symlink handling options
- **Episode ordering system** - Parse and sort TV show episodes intelligently
  - Supports all major naming conventions (SxxExx, xXX, absolute, date-based)
  - Handles multi-episode files and split episodes
  - Smart sorting by season, episode, and part
  - Works with complex filenames containing quality tags and metadata
- **Audio and subtitle track analysis** - Analyze media file tracks using FFmpeg
  - Extract audio and subtitle language information
  - Identify tracks common to all files in a series
  - Support for multiple track variants (CC, Signs, regional variants)
  - Batch analysis across multiple files

## Prerequisites

### Required Dependencies

1. **Go 1.21+** - For building the application
2. **FFmpeg development libraries** - Required for go-astiav
3. **MPV media player** - Required for playback
4. **CGO toolchain** - Required for building (gcc/clang)

### Installing Dependencies

#### Arch Linux
```bash
sudo pacman -S go ffmpeg mpv base-devel
```

#### Ubuntu/Debian
```bash
sudo apt install golang ffmpeg mpv libavcodec-dev libavformat-dev libavutil-dev libswscale-dev build-essential
```

#### Fedora
```bash
sudo dnf install golang ffmpeg mpv ffmpeg-devel gcc
```

#### macOS
```bash
brew install go ffmpeg mpv
```

## Building

### Build the binary
```bash
make build
```

This creates the binary at `build/mpv-launcher`.

### Install to ~/.local/bin
```bash
make install
```

Make sure `~/.local/bin` is in your PATH:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Add this to your `~/.bashrc` or `~/.zshrc` to make it permanent.

## Usage

### Interactive Mode (Default)

Run `mpv-launcher` in a directory containing video files:

```bash
cd /path/to/tv/show
mpv-launcher
```

The launcher will:
1. Scan the current directory for video files (non-recursively)
2. Sort episodes in correct viewing order
3. Analyze audio and subtitle tracks
4. Present playback options based on available tracks
5. Launch MPV with your selected configuration

**Available options** (based on detected tracks):
- `English - No Subtitles` - If English audio is available
- `Japanese - Eng Subtitles` - If Japanese audio and English subtitles are available
- `Japanese - No Subtitles` - If Japanese audio is available but no English subtitles

### Command-Line Flags

```bash
# Scan a specific directory
mpv-launcher -dir /path/to/videos

# Recursively scan subdirectories (up to 100 files)
mpv-launcher -dir /path/to/videos -r

# Scan current directory recursively
mpv-launcher -r
```

### Example Session

```
$ cd "/path/to/Anne of Green Gables/Season 1"
$ mpv-launcher

Scanning directory: .
Found 24 video files
Analyzing audio and subtitle tracks...

Playback options:
1. English - No Subtitles
2. Japanese - Eng Subtitles

Select option: 2

Selected: Japanese - Eng Subtitles
Launching MPV with 24 files...

[MPV plays the entire season in order with Japanese audio and English subtitles]
```

Press `Ctrl+C` to stop playback at any time.

## Video File Scanner API

The scanner package provides functionality to find video files in directories.

### Supported Video Extensions

The scanner recognizes 40+ video file extensions including:
- **Common**: `.mp4`, `.mkv`, `.avi`, `.mov`, `.wmv`, `.flv`, `.webm`, `.m4v`, `.mpg`, `.mpeg`
- **Uncommon**: `.vob`, `.rmvb`, `.mxf`, `.yuv`, `.divx`, `.dv`, `.f4v`, `.roq`
- And many more...

### Usage Example

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/scanner"
)

func main() {
    // Configure scan options
    opts := scanner.ScanOptions{
        Recursive:      true,  // Scan subdirectories
        FollowSymlinks: false, // Don't follow symbolic links
    }

    // Scan directory for video files
    videos, err := scanner.ScanDirectory("/path/to/videos", opts)
    if err != nil {
        panic(err)
    }

    // Print found videos
    for _, video := range videos {
        fmt.Println(video)
    }
}
```

### Scan Multiple Directories

```go
dirs := []string{"/movies", "/tv-shows", "/clips"}
opts := scanner.ScanOptions{Recursive: true}

videos, err := scanner.ScanMultipleDirectories(dirs, opts)
```

### Check If File Is a Video

```go
if scanner.IsVideoFile("movie.mp4") {
    fmt.Println("This is a video file!")
}
```

### Example Program

See `examples/scan-directory.go` for a complete example:

```bash
go run examples/scan-directory.go ~/Videos
go run examples/scan-directory.go ~/Videos --recursive
```

## Complete Integration Example

The `examples/play-series.go` demonstrates full integration of all modules:

```bash
# Play a TV series with auto-detected tracks
go run examples/play-series.go -dir "/path/to/series"

# Specify audio and subtitle languages
go run examples/play-series.go -dir "/path/to/series" -audio jpn -sub eng

# Disable subtitles
go run examples/play-series.go -dir "/path/to/series" -no-sub

# Start in fullscreen
go run examples/play-series.go -dir "/path/to/series" -fullscreen
```

This example:
1. Scans the directory for video files
2. Parses and sorts episodes in correct viewing order
3. Analyzes audio/subtitle tracks (if languages not specified)
4. Launches MPV with the complete playlist and selected tracks

## Episode Ordering API

The episode package provides intelligent parsing and sorting of TV show episodes.

### Supported Formats

- **SxxExx**: `S01E01`, `s02e15`, `S1E1` (with/without leading zeros)
- **xXX**: `1x01`, `2x15`, `1X01` (compact format)
- **Absolute**: `001`, `ep01`, `025` (sequential numbering)
- **Date-based**: `2024-01-15`, `2023.12.25`, `15-01-2024`
- **Multi-episode**: `S01E01-E03`, `S01E01E02E03`
- **Split episodes**: `S01E01 pt1`, `S01E01.1`, `S01E01a`

### Usage Example

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/episode"
    "mpv-launcher/internal/scanner"
)

func main() {
    // Scan for video files
    opts := scanner.ScanOptions{Recursive: true}
    videos, _ := scanner.ScanDirectory("/path/to/show", opts)

    // Parse and sort episodes
    episodes := episode.ParseBatch(videos)
    episode.Sort(episodes)

    // Play in correct order
    for _, ep := range episodes {
        fmt.Printf("S%02dE%02d: %s\n",
            ep.Season, ep.Episodes[0], ep.Filename)
    }
}
```

### Parse Individual File

```go
info := episode.Parse("Show.S02E15.1080p.mkv")

fmt.Println(info.Season)    // 2
fmt.Println(info.Episodes)  // [15]
fmt.Println(info.Type)      // TypeSeasonEpisode
```

### Documentation

See [EPISODE_ORDERING.md](EPISODE_ORDERING.md) for comprehensive documentation including:
- All supported naming conventions
- Multi-episode and split episode handling
- Sorting logic and priorities
- Complex filename examples
- API reference

## Playlist Playback with Track Selection

The player package supports launching MPV with playlists and audio/subtitle track selection.

### Basic Playlist Playback

```go
package main

import (
    "mpv-launcher/internal/player"
)

func main() {
    files := []string{
        "episode01.mkv",
        "episode02.mkv",
        "episode03.mkv",
    }

    opts := player.PlaylistOptions{
        AudioLanguage:    "eng",  // Prefer English audio
        SubtitleLanguage: "eng",  // Prefer English subtitles
        Fullscreen:       true,   // Start in fullscreen
    }

    mpvPlayer, _ := player.NewMPVPlayer()
    mpvPlayer.PlayPlaylist(files, opts)
    mpvPlayer.Wait() // Wait for playback to complete
}
```

### Track Selection Options

```go
// Select by language preference
opts := player.PlaylistOptions{
    AudioLanguage:    "jpn", // Prefer Japanese audio
    SubtitleLanguage: "eng", // Prefer English subtitles
}

// Select by specific track index (from tracks.Track.Index)
opts := player.PlaylistOptions{
    AudioTrack:    2, // Use audio track index 2
    SubtitleTrack: 4, // Use subtitle track index 4
}

// Disable subtitles
opts := player.PlaylistOptions{
    SubtitleTrack: -1, // Disable subtitles
}
```

## Audio and Subtitle Track Analysis

The tracks package analyzes audio and subtitle tracks in media files.

### Single File Analysis

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/tracks"
)

func main() {
    info, err := tracks.ProbeFile("movie.mkv")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Audio tracks: %d\n", len(info.AudioTracks))
    for _, track := range info.AudioTracks {
        fmt.Printf("  [%d] %s - %s\n", track.Index, track.Language, track.Title)
    }

    fmt.Printf("Subtitle tracks: %d\n", len(info.SubtitleTracks))
    for _, track := range info.SubtitleTracks {
        fmt.Printf("  [%d] %s - %s\n", track.Index, track.Language, track.Title)
    }
}
```

### Multi-File Analysis

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/tracks"
    "mpv-launcher/internal/scanner"
)

func main() {
    // Scan for videos
    opts := scanner.ScanOptions{Recursive: true}
    videos, _ := scanner.ScanDirectory("/path/to/series", opts)

    // Analyze all files
    analysis := tracks.AnalyzeFiles(videos)

    fmt.Println("Common audio languages:", analysis.CommonAudioLanguages)
    fmt.Println("Common subtitle languages:", analysis.CommonSubtitleLanguages)

    // Find files missing English audio
    for _, fileInfo := range analysis.Files {
        if !fileInfo.HasLanguage(tracks.TrackTypeAudio, "eng") {
            fmt.Printf("Missing English: %s\n", fileInfo.Filename)
        }
    }
}
```

### Documentation

See [TRACKS.md](TRACKS.md) for comprehensive documentation including:
- Data structures and API reference
- Language detection and aggregation
- Track variants (CC, Signs, regional)
- Integration examples
- Performance considerations

## Development

### Project Structure

```
mpv-launcher/
├── cmd/
│   └── mpv-launcher/     # Main application entry point
│       └── main.go
├── internal/             # Internal packages
│   ├── ffmpeg/          # FFmpeg video probing
│   │   ├── probe.go
│   │   └── probe_test.go
│   ├── player/          # MPV player control
│   │   ├── mpv.go
│   │   └── mpv_test.go
│   ├── scanner/         # Video file scanner
│   │   ├── scanner.go
│   │   └── scanner_test.go
│   ├── episode/         # Episode parsing and ordering
│   │   ├── episode.go
│   │   └── episode_test.go
│   └── tracks/          # Audio/subtitle track analysis
│       ├── tracks.go
│       └── tracks_test.go
├── examples/            # Example programs
│   └── scan-directory.go
├── go.mod              # Go module definition
├── Makefile            # Build and installation tasks
├── README.md           # This file
├── SCANNER.md          # Scanner API documentation
├── EPISODE_ORDERING.md # Episode ordering documentation
└── TRACKS.md           # Track analysis documentation
```

### Running Tests

```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

This generates `coverage.html` for viewing coverage details.

### Managing Dependencies

Download dependencies:
```bash
make deps
```

Tidy dependencies:
```bash
make tidy
```

### Makefile Targets

- `make build` - Build the binary to `build/mpv-launcher`
- `make install` - Install to `~/.local/bin`
- `make uninstall` - Remove from `~/.local/bin`
- `make clean` - Clean build artifacts
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make deps` - Download and verify dependencies
- `make tidy` - Tidy up go.mod and go.sum

## Uninstalling

```bash
make uninstall
```

## Libraries Used

- [go-astiav](https://github.com/asticode/go-astiav) - FFmpeg CGO bindings
- [go-mpv](https://github.com/nbr23/go-mpv) - MPV IPC control library

## License

This is a personal project for direct installation and use.

## Notes

- CGO is required for building due to go-astiav FFmpeg bindings
- MPV must be installed and available in PATH for playback
- The application uses XDG-compatible directories for runtime files
