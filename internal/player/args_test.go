package player

import (
	"slices"
	"testing"

	"mpv-launcher/internal/playlist"
)

var testFiles = []playlist.FilePlayback{
	{Filename: "ep01.mkv", AudioTrack: 2, SubtitleTrack: 1},
	{Filename: "ep02.mkv", AudioTrack: 2, SubtitleTrack: -1},
}

func TestBuildPerFileArgs_NormalWindow(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{})

	if slices.Contains(args, "--fs") {
		t.Error("normal window should not include --fs")
	}
	if slices.Contains(args, "--keepaspect=yes") {
		t.Error("normal window should not force --keepaspect=yes")
	}
	if !slices.Contains(args, "ep01.mkv") || !slices.Contains(args, "ep02.mkv") {
		t.Error("args should contain all filenames")
	}
}

func TestBuildPerFileArgs_Fullscreen(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{Fullscreen: true})

	if !slices.Contains(args, "--fs") {
		t.Error("fullscreen mode should include --fs")
	}
	if slices.Contains(args, "--keepaspect=yes") {
		t.Error("plain fullscreen should not force --keepaspect (mpv default handles it)")
	}
}

func TestBuildPerFileArgs_ScreenName(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{ScreenName: "HDMI-A-2"})

	if !slices.Contains(args, "--fs") {
		t.Error("monitor targeting should include --fs")
	}
	if !slices.Contains(args, "--fs-screen-name=HDMI-A-2") {
		t.Error("monitor targeting should include --fs-screen-name=HDMI-A-2")
	}
	if !slices.Contains(args, "--keepaspect=yes") {
		t.Error("monitor targeting should force --keepaspect=yes")
	}
}


func TestBuildPerFileArgs_PerFileScoping(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{})

	// Each file must be wrapped in --{ ... --}
	openBraces := 0
	closeBraces := 0
	for _, a := range args {
		if a == "--{" {
			openBraces++
		}
		if a == "--}" {
			closeBraces++
		}
	}
	if openBraces != len(testFiles) || closeBraces != len(testFiles) {
		t.Errorf("expected %d --{ and --} pairs, got %d/%d", len(testFiles), openBraces, closeBraces)
	}
}

func TestBuildPerFileArgs_TrackSelection(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{})

	if !slices.Contains(args, "--aid=2") {
		t.Error("should include audio track selection --aid=2")
	}
	if !slices.Contains(args, "--sid=1") {
		t.Error("should include subtitle track selection --sid=1")
	}
	if !slices.Contains(args, "--sid=no") {
		t.Error("should include --sid=no for disabled subtitles")
	}
}

