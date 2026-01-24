# Track Analysis Test Fixtures

This directory contains ffprobe JSON dumps used as test fixtures for the tracks package.

## Generating Fixtures

To create a new fixture from a real media file:

```bash
ffprobe -v quiet -print_format json -show_format -show_streams /path/to/video.mkv > fixture_name.json
```

## Fixture Format

Each fixture is the raw JSON output from ffprobe with `-show_format -show_streams`.
The testutil package parses this and converts it to `tracks.FileTrackInfo`.

## Available Fixtures

- `anime_dual_audio.json` - Japanese + English audio, English subtitles
- `multi_subtitle.json` - Multiple subtitle tracks (eng, fre, ger)
- `no_language_tags.json` - Edge case: streams with no language metadata
- `single_audio.json` - Simple file with one audio track, no subtitles

## Notes

- Language codes should be ISO 639-2 (3-letter codes like `eng`, `jpn`, `fre`)
- Codec names match ffprobe output (e.g., `aac`, `hevc`, `ass`)
