# Audio and Subtitle Track Analysis

The tracks package provides comprehensive analysis of audio and subtitle tracks in media files using FFmpeg's libavformat library.

## Features

- **Single-file probing** - Extract audio and subtitle track information from individual media files
- **Multi-file analysis** - Aggregate track information across multiple files
- **Language detection** - Identify all audio and subtitle languages
- **Common track identification** - Determine which tracks are present in ALL files
- **Detailed metadata** - Extract codec, language, and title information for each track

## Use Cases

- **Quality checking** - Verify all episodes have consistent audio/subtitle tracks
- **Language availability** - Determine which languages are available across a series
- **Track selection** - Choose appropriate audio/subtitle tracks for playback
- **Batch processing** - Identify files missing specific language tracks

## Data Structures

### Track

Represents a single audio or subtitle track.

```go
type Track struct {
    Index    int       // Stream index in the file (0, 1, 2, ...)
    Type     TrackType // "audio" or "subtitle"
    Codec    string    // Codec name (e.g., "aac", "ass", "subrip")
    Language string    // ISO 639-2 language code (e.g., "eng", "jpn", "fre")
    Title    string    // Human-readable title (e.g., "English", "English[CC]")
}
```

### FileTrackInfo

Contains track information for a single media file.

```go
type FileTrackInfo struct {
    Filename       string   // Path to the file
    AudioTracks    []Track  // All audio tracks found
    SubtitleTracks []Track  // All subtitle tracks found
    Error          error    // Error if probing failed (nil on success)
}
```

### TrackSummary

Represents a unique track configuration (used for aggregation).

```go
type TrackSummary struct {
    Type     TrackType  // "audio" or "subtitle"
    Language string     // Language code
    Title    string     // Track title
}
```

### TrackAnalysis

Aggregated analysis across multiple files.

```go
type TrackAnalysis struct {
    Files []FileTrackInfo  // Individual file track info

    // Language aggregation
    AllAudioLanguages       []string  // All unique audio languages found
    AllSubtitleLanguages    []string  // All unique subtitle languages found
    CommonAudioLanguages    []string  // Audio languages in ALL files
    CommonSubtitleLanguages []string  // Subtitle languages in ALL files

    // Detailed track aggregation
    AllAudioTracks       []TrackSummary  // All unique audio tracks
    AllSubtitleTracks    []TrackSummary  // All unique subtitle tracks
    CommonAudioTracks    []TrackSummary  // Audio tracks in ALL files
    CommonSubtitleTracks []TrackSummary  // Subtitle tracks in ALL files
}
```

## API Usage

### Probe a Single File

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

    fmt.Printf("File: %s\n", info.Filename)
    fmt.Printf("Audio tracks: %d\n", len(info.AudioTracks))
    fmt.Printf("Subtitle tracks: %d\n", len(info.SubtitleTracks))

    for _, track := range info.AudioTracks {
        fmt.Printf("  [%d] %s - %s (%s)\n",
            track.Index, track.Language, track.Title, track.Codec)
    }
}
```

**Output:**
```
File: movie.mkv
Audio tracks: 2
Subtitle tracks: 11
  [1] jpn - Japanese (aac)
  [2] eng - English (aac)
```

### Analyze Multiple Files

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/tracks"
)

func main() {
    files := []string{
        "episode01.mkv",
        "episode02.mkv",
        "episode03.mkv",
    }

    analysis := tracks.AnalyzeFiles(files)

    fmt.Printf("Analyzed %d files\n", len(analysis.Files))
    fmt.Printf("Common audio languages: %v\n", analysis.CommonAudioLanguages)
    fmt.Printf("Common subtitle languages: %v\n", analysis.CommonSubtitleLanguages)

    // Check for consistency
    if len(analysis.CommonAudioTracks) < len(analysis.AllAudioTracks) {
        fmt.Println("Warning: Not all files have the same audio tracks!")
    }
}
```

**Output:**
```
Analyzed 3 files
Common audio languages: [jpn eng]
Common subtitle languages: [eng fre ger ita spa por rus ara]
```

### Check Language Availability

```go
info, _ := tracks.ProbeFile("movie.mkv")

// Check if file has English audio
if info.HasLanguage(tracks.TrackTypeAudio, "eng") {
    fmt.Println("English audio available")
}

// Get all English subtitle tracks
engSubs := info.GetTracksByLanguage(tracks.TrackTypeSubtitle, "eng")
fmt.Printf("Found %d English subtitle tracks\n", len(engSubs))
for _, track := range engSubs {
    fmt.Printf("  - %s\n", track.Title)
}
```

**Output:**
```
English audio available
Found 3 English subtitle tracks
  - English
  - English[CC]
  - English[Signs]
```

### Integration with Scanner

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/scanner"
    "mpv-launcher/internal/tracks"
)

func main() {
    // Scan for video files
    opts := scanner.ScanOptions{Recursive: true}
    videos, _ := scanner.ScanDirectory("/path/to/series", opts)

    // Analyze all videos
    analysis := tracks.AnalyzeFiles(videos)

    // Report on common tracks
    fmt.Println("Common Audio Languages:")
    for _, lang := range analysis.CommonAudioLanguages {
        fmt.Printf("  - %s\n", lang)
    }

    // Find files missing English audio
    for _, fileInfo := range analysis.Files {
        if !fileInfo.HasLanguage(tracks.TrackTypeAudio, "eng") {
            fmt.Printf("Missing English audio: %s\n", fileInfo.Filename)
        }
    }
}
```

## Real-World Example

### Verify Series Consistency

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/scanner"
    "mpv-launcher/internal/tracks"
)

func verifySeriesConsistency(seriesPath string) {
    // Find all episodes
    opts := scanner.ScanOptions{Recursive: true}
    episodes, err := scanner.ScanDirectory(seriesPath, opts)
    if err != nil {
        panic(err)
    }

    // Analyze tracks
    analysis := tracks.AnalyzeFiles(episodes)

    // Report summary
    fmt.Printf("Analyzed %d episodes\n\n", len(episodes))

    fmt.Println("Audio Languages:")
    fmt.Printf("  Available: %v\n", analysis.AllAudioLanguages)
    fmt.Printf("  In all files: %v\n", analysis.CommonAudioLanguages)

    fmt.Println("\nSubtitle Languages:")
    fmt.Printf("  Available: %v\n", analysis.AllSubtitleLanguages)
    fmt.Printf("  In all files: %v\n", analysis.CommonSubtitleLanguages)

    // Check for inconsistencies
    if len(analysis.CommonAudioLanguages) < len(analysis.AllAudioLanguages) {
        fmt.Println("\n⚠️  Warning: Audio tracks are inconsistent across files")
        printInconsistencies(analysis)
    }

    if len(analysis.CommonSubtitleLanguages) < len(analysis.AllSubtitleLanguages) {
        fmt.Println("\n⚠️  Warning: Subtitle tracks are inconsistent across files")
    }
}

func printInconsistencies(analysis *tracks.TrackAnalysis) {
    // Find files with different track configurations
    firstFile := analysis.Files[0]

    for i := 1; i < len(analysis.Files); i++ {
        file := analysis.Files[i]

        // Compare audio track counts
        if len(file.AudioTracks) != len(firstFile.AudioTracks) {
            fmt.Printf("  %s has %d audio tracks (expected %d)\n",
                file.Filename, len(file.AudioTracks), len(firstFile.AudioTracks))
        }
    }
}
```

**Output:**
```
Analyzed 23 episodes

Audio Languages:
  Available: [jpn eng]
  In all files: [jpn eng]

Subtitle Languages:
  Available: [eng fre ger ita spa por rus ara]
  In all files: [eng fre ger ita spa por rus ara]
```

## Language Codes

The package uses ISO 639-2 three-letter language codes:

| Code | Language   |
|------|------------|
| eng  | English    |
| jpn  | Japanese   |
| fre  | French     |
| ger  | German     |
| spa  | Spanish    |
| ita  | Italian    |
| por  | Portuguese |
| rus  | Russian    |
| ara  | Arabic     |
| chi  | Chinese    |
| kor  | Korean     |

## Track Variants

Some files may have multiple tracks of the same language with different purposes:

### English Subtitle Variants
- `English` - Standard subtitles
- `English[CC]` - Closed captions (includes sound descriptions)
- `English[Signs]` - Signs and text translation only
- `English[Forced]` - Forced subtitles (foreign language only)

### Spanish Audio Variants
- `Spanish[ESP]` - European Spanish
- `Spanish[LAT]` - Latin American Spanish

## Error Handling

```go
analysis := tracks.AnalyzeFiles(files)

for _, fileInfo := range analysis.Files {
    if fileInfo.Error != nil {
        fmt.Printf("Error analyzing %s: %v\n", fileInfo.Filename, fileInfo.Error)
        continue
    }

    // Process successful files
    fmt.Printf("✓ %s: %d audio, %d subtitle tracks\n",
        fileInfo.Filename,
        len(fileInfo.AudioTracks),
        len(fileInfo.SubtitleTracks))
}
```

## Performance Considerations

- **FFmpeg probing**: Each file requires opening and reading metadata (~10-50ms per file)
- **Parallel processing**: Consider using goroutines for large batches
- **Memory usage**: Minimal - track metadata is small (~100 bytes per track)

### Parallel Analysis Example

```go
import "sync"

func analyzeFilesParallel(files []string) *tracks.TrackAnalysis {
    results := make([]tracks.FileTrackInfo, len(files))
    var wg sync.WaitGroup

    for i, file := range files {
        wg.Add(1)
        go func(index int, path string) {
            defer wg.Done()
            info, err := tracks.ProbeFile(path)
            if err != nil {
                results[index] = tracks.FileTrackInfo{
                    Filename: path,
                    Error:    err,
                }
                return
            }
            results[index] = *info
        }(i, file)
    }

    wg.Wait()

    // Build analysis from results
    // ... (implementation omitted for brevity)
}
```

## Testing

The package includes comprehensive tests:

```bash
go test ./internal/tracks -v
```

Tests cover:
- Single-file probing with real media files
- Multi-file analysis
- Language detection
- Common track identification
- Error handling
- Helper functions (HasLanguage, GetTracksByLanguage)

## Implementation Notes

### FFmpeg Integration

The package uses `go-astiav` (FFmpeg Go bindings) to:
1. Open media files with `OpenInput()`
2. Find stream info with `FindStreamInfo()`
3. Iterate through streams
4. Extract metadata (language, title) from stream tags

### Track Aggregation Algorithm

For multi-file analysis:
1. Probe each file individually
2. Count occurrences of each language/track across files
3. Languages/tracks present in ALL valid files become "common"
4. All unique languages/tracks are collected separately

### Language vs Track Granularity

- **Languages**: Aggregated at language code level (e.g., "eng")
- **Tracks**: Include both language and title (e.g., "eng - English[CC]")

This allows detecting:
- Files with English audio (language level)
- Files with specific English subtitle variants (track level)

## Limitations

- Only extracts metadata (doesn't validate actual stream content)
- Relies on proper metadata in source files
- Attachment streams (fonts) may generate warnings (safe to ignore)
- Requires FFmpeg libraries to be installed

## Future Enhancements

Potential additions:
- Channel layout detection (stereo, 5.1, 7.1)
- Audio bitrate analysis
- Subtitle format detection (SRT, ASS, PGS)
- Default track identification
- Forced subtitle detection
