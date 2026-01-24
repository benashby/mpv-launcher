package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		// Common formats
		{"video.mp4", true},
		{"movie.mkv", true},
		{"clip.avi", true},
		{"film.mov", true},
		{"video.webm", true},
		{"UPPERCASE.MP4", true},
		{"MixedCase.MkV", true},

		// Uncommon formats
		{"old.vob", true},
		{"stream.rmvb", true},
		{"raw.yuv", true},
		{"professional.mxf", true},

		// Non-video files
		{"audio.mp3", false},
		{"document.pdf", false},
		{"image.jpg", false},
		{"text.txt", false},
		{"archive.zip", false},
		{"", false},
		{"no-extension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsVideoFile(tt.filename)
			if result != tt.expected {
				t.Errorf("IsVideoFile(%q) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestScanDirectory_NonExistent(t *testing.T) {
	opts := ScanOptions{Recursive: false}
	_, err := ScanDirectory("/nonexistent/path/to/directory", opts)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestScanDirectory_NotADirectory(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	opts := ScanOptions{Recursive: false}
	_, err = ScanDirectory(tmpFile.Name(), opts)
	if err == nil {
		t.Error("Expected error when scanning a file instead of directory")
	}
}

func TestScanNonRecursive(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "scanner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []struct {
		name      string
		isVideo   bool
		createDir bool
	}{
		{"video1.mp4", true, false},
		{"video2.mkv", true, false},
		{"audio.mp3", false, false},
		{"document.txt", false, false},
		{"subdir", false, true},
	}

	expectedVideos := 0
	for _, tf := range testFiles {
		path := filepath.Join(tmpDir, tf.name)
		if tf.createDir {
			os.Mkdir(path, 0755)
			// Create a video in the subdirectory (should not be found in non-recursive scan)
			subVideoPath := filepath.Join(path, "nested.mp4")
			os.WriteFile(subVideoPath, []byte("test"), 0644)
		} else {
			os.WriteFile(path, []byte("test"), 0644)
			if tf.isVideo {
				expectedVideos++
			}
		}
	}

	opts := ScanOptions{Recursive: false}
	videos, err := ScanDirectory(tmpDir, opts)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(videos) != expectedVideos {
		t.Errorf("Expected %d videos, got %d", expectedVideos, len(videos))
	}

	// Verify all returned files are video files
	for _, video := range videos {
		if !IsVideoFile(video) {
			t.Errorf("Non-video file returned: %s", video)
		}
	}
}

func TestScanRecursive(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "scanner-test-recursive-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directory structure
	subDir1 := filepath.Join(tmpDir, "subdir1")
	subDir2 := filepath.Join(tmpDir, "subdir1", "subdir2")
	os.MkdirAll(subDir2, 0755)

	// Create test files at various levels
	testFiles := map[string]bool{
		filepath.Join(tmpDir, "root.mp4"):          true,
		filepath.Join(tmpDir, "root.txt"):          false,
		filepath.Join(subDir1, "level1.mkv"):       true,
		filepath.Join(subDir1, "level1.jpg"):       false,
		filepath.Join(subDir2, "level2.avi"):       true,
		filepath.Join(subDir2, "level2.mp3"):       false,
		filepath.Join(tmpDir, "another.webm"):      true,
		filepath.Join(subDir1, "uncommon.rmvb"):    true,
	}

	expectedVideos := 0
	for path, isVideo := range testFiles {
		os.WriteFile(path, []byte("test"), 0644)
		if isVideo {
			expectedVideos++
		}
	}

	opts := ScanOptions{Recursive: true}
	videos, err := ScanDirectory(tmpDir, opts)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(videos) != expectedVideos {
		t.Errorf("Expected %d videos, got %d", expectedVideos, len(videos))
		t.Logf("Found videos: %v", videos)
	}

	// Verify all returned files exist and are video files
	for _, video := range videos {
		if !IsVideoFile(video) {
			t.Errorf("Non-video file returned: %s", video)
		}
		if _, err := os.Stat(video); err != nil {
			t.Errorf("Returned file doesn't exist: %s", video)
		}
	}
}

func TestScanMultipleDirectories(t *testing.T) {
	// Create two temporary directories
	tmpDir1, err := os.MkdirTemp("", "scanner-multi-1-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "scanner-multi-2-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	// Create files in first directory
	os.WriteFile(filepath.Join(tmpDir1, "video1.mp4"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir1, "video2.mkv"), []byte("test"), 0644)

	// Create files in second directory
	os.WriteFile(filepath.Join(tmpDir2, "video3.avi"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir2, "audio.mp3"), []byte("test"), 0644)

	opts := ScanOptions{Recursive: false}
	videos, err := ScanMultipleDirectories([]string{tmpDir1, tmpDir2}, opts)
	if err != nil {
		t.Fatalf("ScanMultipleDirectories failed: %v", err)
	}

	expectedVideos := 3
	if len(videos) != expectedVideos {
		t.Errorf("Expected %d videos from both directories, got %d", expectedVideos, len(videos))
	}
}

func TestVideoExtensions(t *testing.T) {
	// Verify we have a good set of extensions
	commonExtensions := []string{".mp4", ".mkv", ".avi", ".mov", ".webm"}
	for _, ext := range commonExtensions {
		if !VideoExtensions[ext] {
			t.Errorf("Common video extension %s not found in VideoExtensions", ext)
		}
	}

	uncommonExtensions := []string{".rmvb", ".vob", ".mxf", ".yuv"}
	for _, ext := range uncommonExtensions {
		if !VideoExtensions[ext] {
			t.Errorf("Uncommon video extension %s not found in VideoExtensions", ext)
		}
	}

	// Verify we have a decent number of extensions
	if len(VideoExtensions) < 20 {
		t.Errorf("Expected at least 20 video extensions, got %d", len(VideoExtensions))
	}
}
