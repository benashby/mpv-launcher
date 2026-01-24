# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build           # Build to build/mpv-launcher
make test            # Run all tests
make test-coverage   # Generate coverage.html
make install         # Install to ~/.local/bin
make deps            # Download dependencies
make tidy            # Tidy go.mod
```

Run a single test:
```bash
go test -v ./internal/scanner -run TestScanDirectory
```

CGO is required (FFmpeg bindings). System FFmpeg dev libraries must be installed.

## Architecture

Go application that scans directories for video files, orders episodes intelligently, analyzes audio/subtitle tracks via FFmpeg, and launches MPV.

**Main entry point**: `cmd/mpv-launcher/main.go` - CLI with user interaction

**Core packages** (all in `internal/`):
- **scanner** - Discovers video files (40+ formats), handles symlinks, supports recursive scanning
- **episode** - Parses filenames to extract season/episode info, supports SxxExx, xXX, absolute, date-based formats. Key types: `EpisodeInfo`, `Type` enum
- **tracks** - FFmpeg-based audio/subtitle track analysis. Aggregates common languages across files. Uses `go-astiav` CGO bindings
- **priority** - Defines audio/subtitle `Permutation` combinations (e.g., jpn+eng, eng+none) with ordering logic
- **playlist** - Builds per-file track selections based on priority order, handles fallbacks
- **ui** - Interactive terminal UI: `PrioritySelector` for reordering items with vim-like keys (j/k/J/K)
- **player** - Launches MPV with playlist and track selection (by language or index)
- **ffmpeg** - Low-level video probing (duration, resolution, codec)
- **testutil** - Test helpers: fixture loading, ffprobe JSON parsing, comparison utilities

**Flow**: scan → parse episodes → sort → analyze tracks → select priorities (UI) → build playlist → launch MPV

## Key Constraints

- Max 100 video files per scan (hardcoded safety limit)
- Uses absolute paths throughout (`filepath.Abs()`)
- Language codes are ISO 639-2 (3-letter: eng, jpn, fre)
- Module imports use `mpv-launcher/internal/...`
