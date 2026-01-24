# Video Probe Test Fixtures

This directory contains ffprobe JSON dumps used as test fixtures for the ffmpeg package.

## Generating Fixtures

To create a new fixture from a real media file:

```bash
ffprobe -v quiet -print_format json -show_format -show_streams /path/to/video.mkv > fixture_name.json
```

## Fixture Format

Each fixture is the raw JSON output from ffprobe with `-show_format -show_streams`.
The testutil package parses this and converts it to `ffmpeg.VideoInfo`.

## Available Fixtures

- `video_1080p_hevc.json` - Standard 1080p HEVC video
- `video_4k_hdr.json` - 4K HDR content
- `video_720p_h264.json` - 720p H.264 video

## Notes

- Duration is in seconds in the format section
- Frame rate is in "num/den" format (e.g., "24000/1001" for 23.976fps)
- Resolution is in the video stream's width/height fields
