# Episode Ordering System

The episode package provides comprehensive parsing and ordering of TV show episode files with support for virtually all common and uncommon naming conventions.

## Supported Naming Formats

### 1. SxxExx Format (Standard)

The most widely used format with full season and episode information.

```
Show.S01E01.mkv          → Season 1, Episode 1
show.s02e15.mkv          → Season 2, Episode 15 (lowercase)
Show.S1E1.mkv            → Season 1, Episode 1 (no leading zeros)
Show.S00E01.mkv          → Season 0 (Special), Episode 1
```

**Variations supported:**
- Uppercase: `S01E01`
- Lowercase: `s01e01`
- Mixed case: `S01e01`, `s01E01`
- With/without leading zeros: `S1E1`, `S01E1`

### 2. xXX Format

Compact season x episode format.

```
Breaking Bad 1x01.mkv    → Season 1, Episode 1
Show 2x15.mkv            → Season 2, Episode 15
Show 01x01.mkv           → Season 1, Episode 1 (zero-padded)
Show 1X01.mkv            → Season 1, Episode 1 (uppercase X)
```

### 3. Dash/Underscore Variations

Formats using separators instead of direct concatenation.

```
Show S01-E01.mkv         → Season 1, Episode 1
Show S01_E01.mkv         → Season 1, Episode 1
Show 1-01.mkv            → Season 1, Episode 1
Show 2_15.mkv            → Season 2, Episode 15
```

### 4. Absolute/Sequential Numbering

Simple episode numbering without explicit season information.

```
Show 001.mkv             → Absolute Episode 1
Show 025.mkv             → Absolute Episode 25
Show ep01.mkv            → Absolute Episode 1 (with prefix)
Show ep 15.mkv           → Absolute Episode 15
```

**Note:** As configured, absolute numbering treats all numbers as episode-only (no season extraction from patterns like 101, 201).

### 5. Date-Based Naming

For daily shows, talk shows, and news programs.

```
The Daily Show 2024-01-15.mkv      → January 15, 2024
Show.2023.12.25.mkv                → December 25, 2023
Show 15-01-2024.mkv                → January 15, 2024 (European format)
```

**Supported date formats:**
- `YYYY-MM-DD` (ISO standard)
- `YYYY.MM.DD` (dot-separated)
- `DD-MM-YYYY` (European format)

### 6. Multi-Episode Files

Files containing multiple episodes in a single file.

#### Range Format
```
Show S01E01-E03.mkv      → Episodes 1, 3 (and implicitly 2)
Show S01E01-03.mkv       → Episodes 1, 3
Show.S02E10-E12.mkv      → Episodes 10, 12
```

#### Stacked Format
```
Show S01E01E02E03.mkv    → Episodes 1, 2, 3
Show.S02E10E11.mkv       → Episodes 10, 11
```

### 7. Split Episodes

Single episodes split across multiple files.

#### Part Numbers
```
Show S01E01 - pt1.mkv    → Season 1, Episode 1, Part 1
Show S01E01 Part 2.mkv   → Season 1, Episode 1, Part 2
Show.S01E01.pt1.mkv      → Season 1, Episode 1, Part 1
```

#### Sub-Parts (for children's cartoons)
```
Show S01E01.1.mkv        → Season 1, Episode 1, Part 1
Show S01E01.2.mkv        → Season 1, Episode 1, Part 2
Show S01E01a.mkv         → Season 1, Episode 1, Part a
Show S01E01b.mkv         → Season 1, Episode 1, Part b
```

## API Usage

### Basic Parsing

```go
import "mpv-launcher/internal/episode"

// Parse a single filename
info := episode.Parse("Show.S01E05.1080p.mkv")

fmt.Println(info.Season)          // 1
fmt.Println(info.Episodes)        // [5]
fmt.Println(info.Type)            // TypeSeasonEpisode
fmt.Println(info.MatchedPattern)  // "SxxExx"
```

### Batch Parsing

```go
filenames := []string{
    "Show S02E01.mkv",
    "Show S01E03.mkv",
    "Show S01E01.mkv",
}

episodes := episode.ParseBatch(filenames)
// Returns: []*EpisodeInfo for each file
```

### Sorting Episodes

```go
filenames := []string{
    "Show S02E01.mkv",
    "Show S01E03.mkv",
    "Show S01E01.mkv",
    "Show S01E02.mkv",
}

episodes := episode.ParseBatch(filenames)
episode.Sort(episodes)

// Episodes are now sorted:
// S01E01, S01E02, S01E03, S02E01
```

### Sorting with Mixed Types

The sorting algorithm handles mixed episode types correctly:

```go
filenames := []string{
    "Show 2024-01-15.mkv",    // Date-based
    "Show 001.mkv",            // Absolute numbering
    "Show S01E01.mkv",         // Season/Episode
}

episodes := episode.ParseBatch(filenames)
episode.Sort(episodes)

// Sorted order:
// 1. S01E01 (TypeSeasonEpisode - highest priority)
// 2. 001    (TypeAbsolute)
// 3. 2024-01-15 (TypeDate - sorted chronologically)
```

### Checking Episode Information

```go
info := episode.Parse("Show S01E01-E03.mkv")

if info.Type == episode.TypeSeasonEpisode {
    fmt.Printf("Season %d\n", info.Season)
    fmt.Printf("Episodes: %v\n", info.Episodes) // [1, 3]

    if len(info.Episodes) > 1 {
        fmt.Println("This is a multi-episode file")
    }
}

if info.Part != "" {
    fmt.Printf("Part: %s\n", info.Part)
}
```

## Episode Types

The parser distinguishes between different types of episode numbering:

```go
const (
    TypeUnknown       = 0  // Could not determine episode info
    TypeSeasonEpisode = 1  // Standard SxxExx or xXX format
    TypeAbsolute      = 2  // Sequential numbering
    TypeDate          = 3  // Date-based episodes
)
```

### Sorting Priority

Episodes are sorted in this order:
1. **TypeSeasonEpisode** - By season, then episode, then part
2. **TypeAbsolute** - By absolute episode number, then part
3. **TypeDate** - Chronologically by date
4. **TypeUnknown** - By filename (lexicographically)

## Complex Filename Handling

The parser is robust against complex filenames with quality tags, group names, and other metadata:

```go
// All of these parse correctly:
episode.Parse("[GroupName] Show - S01E05 - Title [1080p][x264].mkv")
episode.Parse("Show.Name.S02E10.1080p.WEB-DL.DD5.1.H264-GROUP.mkv")
episode.Parse("Show_Name_-_S03E15_-_720p_HDTV_x264.mkv")
```

## Examples

### Complete Sorting Example

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/episode"
)

func main() {
    videos := []string{
        "Game of Thrones S08E06.mkv",
        "Game of Thrones S08E01.mkv",
        "Game of Thrones S07E07.mkv",
        "Game of Thrones S08E03 pt1.mkv",
        "Game of Thrones S08E03 pt2.mkv",
        "Game of Thrones S00E01.mkv",  // Special
    }

    episodes := episode.ParseBatch(videos)
    episode.Sort(episodes)

    fmt.Println("Viewing order:")
    for i, ep := range episodes {
        if ep.Part != "" {
            fmt.Printf("%d. S%02dE%02d Part %s\n",
                i+1, ep.Season, ep.Episodes[0], ep.Part)
        } else {
            fmt.Printf("%d. S%02dE%02d\n",
                i+1, ep.Season, ep.Episodes[0])
        }
    }
}
```

Output:
```
Viewing order:
1. S00E01
2. S07E07
3. S08E01
4. S08E03 Part 1
5. S08E03 Part 2
6. S08E06
```

### Integration with Scanner

```go
package main

import (
    "fmt"
    "mpv-launcher/internal/scanner"
    "mpv-launcher/internal/episode"
)

func main() {
    // Scan directory for video files
    opts := scanner.ScanOptions{Recursive: true}
    videos, _ := scanner.ScanDirectory("/path/to/tv/show", opts)

    // Parse and sort episodes
    episodes := episode.ParseBatch(videos)
    episode.Sort(episodes)

    // Play in order
    for _, ep := range episodes {
        fmt.Printf("Playing: %s\n", ep.Filename)
        // Launch player here
    }
}
```

## Error Handling

Files that don't match any pattern are marked as `TypeUnknown`:

```go
info := episode.Parse("random_file.mkv")

if info.Type == episode.TypeUnknown {
    fmt.Println("Could not determine episode information")
    // File will sort by filename
}
```

## Testing

The package includes comprehensive tests covering all naming conventions:

```bash
go test ./internal/episode -v
```

Test coverage includes:
- All standard formats (SxxExx, xXX, etc.)
- Edge cases (season 0, high episode numbers, etc.)
- Multi-episode files (all variations)
- Split episodes (all part formats)
- Complex filenames with metadata
- Sorting with mixed types
- Date-based episodes

## Performance Considerations

- **Regex compilation**: Patterns are compiled once at package init
- **Linear parsing**: Each file is parsed with O(1) pattern matching
- **Efficient sorting**: Uses Go's standard sort with custom comparator
- **Memory**: Minimal allocations, Episode Info structs are lightweight

## Implementation Notes

### Pattern Priority

Patterns are tried in this order to avoid false matches:

1. Date formats (before dash/underscore to avoid DD-MM-YYYY being parsed as season-episode)
2. Multi-episode patterns (most specific SxxExx variant)
3. Standard SxxExx
4. xXX format
5. Dash/underscore variations
6. Ep-prefix format
7. Absolute numbering (last resort to avoid false matches)

### Equivalence

Files with equivalent semantics but different formatting sort together:

```go
// These are treated as equal:
"Show S1E1.mkv" == "Show S01E01.mkv"  // Same season & episode
```

### Extensibility

To add support for new naming conventions:

1. Add regex pattern to `var` block
2. Create parse function (e.g., `parseNewFormat`)
3. Add to `Parse()` function in appropriate priority order
4. Add tests
