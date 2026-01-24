# High School DxD Born Test Fixtures

These fixtures are real ffprobe dumps from the [Judas] release of High School DxD Born.

## Track Structure

Each episode has:
- 1 video stream (HEVC 10-bit 1080p)
- 2 audio streams (both labeled "eng" but one is Japanese, one is English)
  - Stream 1: Japanese Stereo (mislabeled as "eng", title says "Japanese")
  - Stream 2: English Stereo (correctly labeled "eng", title says "English")
- 2 subtitle streams
  - Stream 3: English (Full) - ASS format
  - Stream 4: English (Signs/Songs) - PGS format
- Multiple attachment streams (fonts)

## Expected Behavior

The `GetEffectiveLanguage` function should:
1. Detect Stream 1 as "jpn" from the title "Japanese Stereo"
2. Detect Stream 2 as "eng" from the title "English Stereo"

This allows the user to choose:
1. English audio (no subtitles needed)
2. Japanese audio with English subtitles

## Generation

```bash
ffprobe -v quiet -print_format json -show_format -show_streams "/path/to/episode.mkv" > epXX.json
```
