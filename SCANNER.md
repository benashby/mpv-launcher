# Video File Scanner

The scanner package provides robust video file detection and directory scanning capabilities.

## Features

- **Comprehensive format support**: 40+ video file extensions (mp4, mkv, avi, mov, webm, vob, rmvb, mxf, and more)
- **Recursive scanning**: Optional deep directory traversal
- **Symlink handling**: Configurable symlink following with cycle prevention
- **Multiple directory support**: Scan multiple directories with deduplication
- **Case-insensitive**: Matches extensions regardless of case (MP4, Mp4, mp4)
- **Absolute paths**: Always returns absolute paths to found videos

## API Overview

### Types

```go
type ScanOptions struct {
    Recursive      bool  // Scan subdirectories
    FollowSymlinks bool  // Follow symbolic links
}
```

### Functions

#### `ScanDirectory(dirPath string, opts ScanOptions) ([]string, error)`
Scans a single directory for video files.

**Parameters:**
- `dirPath`: Path to directory to scan
- `opts`: Scan configuration options

**Returns:**
- `[]string`: Slice of absolute paths to video files
- `error`: Error if directory doesn't exist or can't be read

#### `ScanMultipleDirectories(dirPaths []string, opts ScanOptions) ([]string, error)`
Scans multiple directories and returns combined, deduplicated results.

**Parameters:**
- `dirPaths`: Slice of directory paths to scan
- `opts`: Scan configuration options

**Returns:**
- `[]string`: Deduplicated slice of absolute paths to video files
- `error`: Error if any directory fails to scan

#### `IsVideoFile(filename string) bool`
Checks if a filename has a recognized video extension.

**Parameters:**
- `filename`: Filename or path to check

**Returns:**
- `bool`: true if file has a video extension

### Supported Extensions

#### Common Formats
`.mp4`, `.mkv`, `.avi`, `.mov`, `.wmv`, `.flv`, `.webm`, `.m4v`, `.mpg`, `.mpeg`, `.m2v`, `.3gp`, `.3g2`, `.ogv`, `.ts`, `.mts`, `.m2ts`

#### Uncommon/Specialized Formats
`.vob`, `.asf`, `.rm`, `.rmvb`, `.divx`, `.dv`, `.f4v`, `.mxf`, `.roq`, `.yuv`, `.nsv`, `.gxf`, `.qt`, `.xvid`

#### Additional Formats
`.mp2`, `.mpe`, `.mpv`, `.m4p`, `.m4b`, `.dat`, `.vcd`, `.svcd`, `.drc`, `.gif`, `.gifv`, `.mng`, `.viv`, `.amv`, `.m2p`, `.m2t`, `.m4s`, `.tod`, `.vro`, `.wtv`

## Examples

### Basic Non-Recursive Scan

```go
opts := scanner.ScanOptions{
    Recursive:      false,
    FollowSymlinks: false,
}

videos, err := scanner.ScanDirectory("/path/to/videos", opts)
if err != nil {
    log.Fatal(err)
}

for _, video := range videos {
    fmt.Println(video)
}
```

### Recursive Scan with Symlink Following

```go
opts := scanner.ScanOptions{
    Recursive:      true,
    FollowSymlinks: true,
}

videos, err := scanner.ScanDirectory("/media/videos", opts)
```

### Scan Multiple Directories

```go
directories := []string{
    "/home/user/Movies",
    "/media/external/Videos",
    "/mnt/nas/Films",
}

opts := scanner.ScanOptions{Recursive: true}
allVideos, err := scanner.ScanMultipleDirectories(directories, opts)
```

### Check Individual Files

```go
files := []string{"movie.mp4", "document.pdf", "clip.MKV"}

for _, file := range files {
    if scanner.IsVideoFile(file) {
        fmt.Printf("%s is a video\n", file)
    }
}
```

## CLI Example

Run the included example program:

```bash
# Non-recursive scan
go run examples/scan-directory.go ~/Videos

# Recursive scan
go run examples/scan-directory.go ~/Videos --recursive
```

## Error Handling

The scanner handles several error conditions gracefully:

- **Directory doesn't exist**: Returns error
- **Path is a file, not directory**: Returns error
- **Permission denied**: Skips directory and continues
- **Broken symlinks**: Skips and continues
- **Symlink cycles**: Prevented when FollowSymlinks is enabled

## Performance Considerations

- **Non-recursive scans** use `os.ReadDir()` for optimal performance
- **Recursive scans** use `filepath.WalkDir()` which is memory-efficient
- **Extension matching** uses a map lookup (O(1)) for fast filtering
- **Symlink cycle prevention** avoids infinite loops in recursive scans

## Testing

The package includes comprehensive tests:

```bash
go test ./internal/scanner -v
```

Tests cover:
- Video extension detection (common and uncommon formats)
- Non-recursive directory scanning
- Recursive directory scanning
- Multiple directory scanning
- Error conditions
- Edge cases (empty directories, permission errors, etc.)
