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

func TestBuildPerFileArgs_Profile(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{Profile: "anime"})

	if !slices.Contains(args, "--profile=anime") {
		t.Error("should include --profile=anime")
	}
}

func TestBuildPerFileArgs_NoProfile(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{})

	for _, a := range args {
		if len(a) > 9 && a[:9] == "--profile" {
			t.Errorf("no profile set but got %q", a)
		}
	}
}

func TestBuildPerFileArgs_NoDeband(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{NoDeband: true})

	if !slices.Contains(args, "--deband=no") {
		t.Error("NoDeband should include --deband=no")
	}
}

func TestBuildPerFileArgs_DeBandDefault(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{})

	if slices.Contains(args, "--deband=no") {
		t.Error("deband should not be overridden when NoDeband is false")
	}
}

func TestBuildPerFileArgs_HwdecMode(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{HwdecMode: "vaapi"})

	if !slices.Contains(args, "--hwdec=vaapi") {
		t.Error("should include --hwdec=vaapi")
	}
}

func TestBuildPerFileArgs_HwdecModeEmpty(t *testing.T) {
	args := buildPerFileArgs(testFiles, PlaylistOptions{})

	for _, a := range args {
		if len(a) > 7 && a[:7] == "--hwdec" {
			t.Errorf("no hwdec set but got %q", a)
		}
	}
}

func TestBuildPerFileArgs_IPCSocket(t *testing.T) {
	sock := "/run/user/1000/mpv-launcher.sock"
	args := buildPerFileArgs(testFiles, PlaylistOptions{IPCSocket: sock})

	if !slices.Contains(args, "--input-ipc-server="+sock) {
		t.Error("should include --input-ipc-server with socket path")
	}
}

func TestBuildPerFileArgs_GlobalOptsBeforePerFile(t *testing.T) {
	// Global options must appear before the first --{ so mpv applies them session-wide.
	args := buildPerFileArgs(testFiles, PlaylistOptions{
		Profile:   "anime",
		NoDeband:  true,
		HwdecMode: "vaapi",
		IPCSocket: "/run/user/1000/mpv-launcher.sock",
	})

	firstBrace := -1
	for i, a := range args {
		if a == "--{" {
			firstBrace = i
			break
		}
	}
	if firstBrace < 0 {
		t.Fatal("no --{ found in args")
	}

	globalArgs := args[:firstBrace]
	for _, want := range []string{"--profile=anime", "--deband=no", "--hwdec=vaapi", "--input-ipc-server=/run/user/1000/mpv-launcher.sock"} {
		if !slices.Contains(globalArgs, want) {
			t.Errorf("global option %q should appear before first --{", want)
		}
	}
}

