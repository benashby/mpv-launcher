# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

All build operations require `CGO_ENABLED=1` (FFmpeg bindings are CGO-based). The Makefile exports this automatically.

```bash
make build          # outputs build/mpv-launcher
make install        # installs to ~/.local/bin
make test           # go test -v ./cmd/... ./internal/...
make test-coverage  # produces coverage.html
make tidy           # go mod tidy
make clean
```

There is no `make lint`. The Nix dev shell (`flake.nix`) provides `go-tools` (staticcheck). To enter it: `nix develop`.

System prerequisites: Go 1.21+, FFmpeg dev libraries, MPV, and a C toolchain.

## Architecture

The program is a linear 6-step pipeline:

```
scanner → episode → tracks → [ui: monitor + priority] → playlist → player
```

1. **scanner** — scans a directory for video files by extension
2. **episode** — parses filenames with a chain of regexes into `EpisodeInfo`, then sorts (season specials sorted to end)
3. **tracks** — uses `go-astiav` (CGO FFmpeg bindings) to probe each file's audio/subtitle streams into `Track` structs with both FFmpeg stream index and MPV's 1-based `--aid`/`--sid` index
4. **ui** — two Bubble Tea TUI components: `CursorSelect` (simple list) and `PrioritySelector` (reorderable list with `j`/`k` to move cursor, `J`/`K` to swap items)
5. **playlist** — `Builder` maps each file to specific audio/subtitle MpvIndex values based on the ordered priority list, with fallback to first available track
6. **player** — launches MPV via `os/exec` using per-file options syntax (`--{ --aid=N --sid=M file.mkv --}`) for different tracks per file

Optional: **hyprland** — if `HYPRLAND_INSTANCE_SIGNATURE` is set, queries the Hyprland socket for monitor list and passes `--fs-screen-name=<name>` to MPV.

## Key Design Details

**Track language detection** (`internal/tracks/tracks.go`): `GetEffectiveLanguage()` checks the track *title* field for language words (e.g. "japanese", "english") before falling back to the metadata language tag. This handles mis-tagged fan-encoded files.

**Priority permutations** (`internal/priority/priority.go`): 4 hardcoded audio+subtitle pairings. Subtitle disabled is represented as `subIdx = -1`. `CanSatisfy()` checks if a file has the required tracks.

**MPV is subprocess-only**: there is no IPC or socket control. The `go-mpv` package is not used.

**`ffmpeg` package vs `tracks` package**: both wrap go-astiav but serve different purposes. `ffmpeg` extracts video-stream info (duration, width, height, fps); `tracks` handles audio/subtitle probing for playback selection.

## Testing

Tests use fixture injection to avoid real media files:

- `internal/testutil/` — `FFProbeOutput` structs and `LoadFFProbeJSON()` for loading `.json` fixture files
- `internal/testutil/probers/` — `FixtureFileProber` implements `tracks.FileProber`, returning track info from fixture JSON instead of calling FFmpeg
- Fixture JSON files live in each package's `testdata/` directory; `testutil.TestdataPath()` resolves them via `runtime.Caller`

Real-world fixture data from an anime series (High School DxD) exists in `tracks/` as representative test cases for the full analysis pipeline.

`player.buildPerFileArgs()` is exported package-internally so `args_test.go` can test MPV argument construction without launching a process.
