# High School DxD Hero Test Fixtures

Real ffprobe dumps from the [Judas] release of High School DxD Hero (Season 4).

## Track Structure

- **Audio 1**: Japanese Stereo (mislabeled as `eng`, title contains "Japanese")
- **Audio 2**: English Stereo (mislabeled as `eng`, title has typo "Englsih")
- **Subtitle 1**: English (Full) - ASS format
- **Subtitle 2**: English (Signs/Songs) - PGS format

## Notes

- Both audio tracks have incorrect `lang=eng` metadata
- The English title has a typo: "Englsih" instead of "English"
- Language detection relies on title parsing

## Expected Options

1. English audio (no subtitles)
2. Japanese audio + English subtitles
