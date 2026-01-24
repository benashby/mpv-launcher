# Playlist Playback with Track Selection

The player package provides comprehensive playlist support with audio and subtitle track selection capabilities.

## Features

- **Multiple file playback** - Play entire series or seasons in one session
- **Track selection by index** - Use specific FFmpeg stream indices
- **Track selection by language** - Prefer specific audio/subtitle languages
- **Subtitle control** - Enable, disable, or auto-select subtitles
- **Fullscreen mode** - Start playback in fullscreen

## API Reference

### PlaylistOptions

```go
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
```

### PlayPlaylist

```go
func (p *MPVPlayer) PlayPlaylist(files []string, opts PlaylistOptions) error
```

Launches MPV with a playlist of files and the specified playback options.

## Usage Examples

### Basic Playlist

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
        AudioLanguage: "eng",
    }

    mpvPlayer, _ := player.NewMPVPlayer()
    mpvPlayer.PlayPlaylist(files, opts)
    mpvPlayer.Wait()
}
```

### Language Preference

```go
// Prefer Japanese audio with English subtitles
opts := player.PlaylistOptions{
    AudioLanguage:    "jpn",
    SubtitleLanguage: "eng",
}
```

### Specific Track Indices

```go
// Use track indices from tracks.Track.Index
opts := player.PlaylistOptions{
    AudioTrack:    2, // Second audio track
    SubtitleTrack: 4, // Fourth subtitle track
}
```

### Disable Subtitles

```go
opts := player.PlaylistOptions{
    SubtitleTrack: -1, // Disable subtitles
}
```

### Fullscreen Playback

```go
opts := player.PlaylistOptions{
    AudioLanguage: "eng",
    Fullscreen:    true,
}
```

## Integration with Other Modules

The player module integrates seamlessly with other mpv-launcher modules:

### Complete Example

```go
package main

import (
    "mpv-launcher/internal/episode"
    "mpv-launcher/internal/player"
    "mpv-launcher/internal/scanner"
    "mpv-launcher/internal/tracks"
)

func main() {
    // 1. Scan directory for videos
    opts := scanner.ScanOptions{Recursive: true}
    videos, _ := scanner.ScanDirectory("/path/to/series", opts)

    // 2. Parse and sort episodes
    episodes := episode.ParseBatch(videos)
    episode.Sort(episodes)

    // 3. Build ordered playlist
    playlist := make([]string, len(episodes))
    for i, ep := range episodes {
        playlist[i] = ep.Filename
    }

    // 4. Analyze tracks (optional)
    analysis := tracks.AnalyzeFiles(videos)

    // 5. Play with preferred language
    playlistOpts := player.PlaylistOptions{
        AudioLanguage: analysis.CommonAudioLanguages[0],
    }

    mpvPlayer, _ := player.NewMPVPlayer()
    mpvPlayer.PlayPlaylist(playlist, playlistOpts)
    mpvPlayer.Wait()
}
```

## MPV Command-Line Mapping

The PlaylistOptions are translated to MPV command-line arguments:

| Option | MPV Argument | Example |
|--------|--------------|---------|
| `AudioTrack: 2` | `--aid=2` | Select audio track index 2 |
| `SubtitleTrack: 4` | `--sid=4` | Select subtitle track index 4 |
| `SubtitleTrack: -1` | `--sid=no` | Disable subtitles |
| `AudioLanguage: "eng"` | `--alang=eng` | Prefer English audio |
| `SubtitleLanguage: "fre"` | `--slang=fre` | Prefer French subtitles |
| `Fullscreen: true` | `--fs` | Start in fullscreen |

## Testing

The player package includes comprehensive tests:

```bash
go test ./internal/player -v
```

Tests cover:
- Empty playlist validation
- Audio track selection by index
- Subtitle track selection by index
- Language preference selection
- Subtitle disabling
- Multiple file playlists

## Example Programs

### examples/play-series.go

A complete integration example that demonstrates:
1. Directory scanning
2. Episode parsing and sorting
3. Track analysis
4. Playlist playback with language selection

```bash
# Play a TV series with Japanese audio
go run examples/play-series.go -dir "/path/to/series" -audio jpn

# Play with English audio and subtitles
go run examples/play-series.go -dir "/path/to/series" -audio eng -sub eng

# Play without subtitles in fullscreen
go run examples/play-series.go -dir "/path/to/series" -no-sub -fullscreen
```

## Notes

- MPV must be installed and available in PATH
- Track indices come from FFmpeg stream indices (use tracks.ProbeFile to discover them)
- Language codes follow ISO 639-2 (eng, jpn, fre, ger, etc.)
- When both track index and language are specified, track index takes precedence
- Empty playlists return an error
- The player automatically advances to the next file when one finishes
