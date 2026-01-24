package player

import (
	"os"
	"path/filepath"
	"testing"
)

const testMediaDir = "/mnt/media/family/tv/Anne Shirley/Season 1"

func TestNewMPVPlayer(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create MPV player: %v", err)
	}

	if player == nil {
		t.Error("Expected non-nil player")
	}
}

func TestStop_WithoutStarting(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Stop should not panic when called without starting
	err = player.Stop()
	if err != nil {
		t.Errorf("Stop() returned error when no process running: %v", err)
	}
}

func TestWait_WithoutStarting(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Wait should return error when MPV hasn't been started
	err = player.Wait()
	if err == nil {
		t.Error("Wait() should return error when MPV not started")
	}
}

func TestPlayPlaylist_EmptyPlaylist(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Empty playlist should return error
	err = player.PlayPlaylist([]string{}, PlaylistOptions{})
	if err == nil {
		t.Error("PlayPlaylist should return error for empty playlist")
	}
}

func TestPlayPlaylist_WithAudioTrackIndex(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Test that specifying audio track index doesn't cause errors
	opts := PlaylistOptions{
		AudioTrack: 2, // Select second audio track
	}

	testFile := filepath.Join(testMediaDir, "Anne Shirley - S01E05 - Let Us Look on the Bright Side of Things HDTV-720p Proper.mkv")

	// Check if test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	files := []string{testFile}

	// This will start MPV - we just verify it doesn't error on startup
	err = player.PlayPlaylist(files, opts)
	if err != nil {
		t.Fatalf("PlayPlaylist failed: %v", err)
	}

	// Clean up - stop the player immediately
	player.Stop()
}

func TestPlayPlaylist_WithLanguagePreference(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	opts := PlaylistOptions{
		AudioLanguage:    "eng",
		SubtitleLanguage: "eng",
	}

	testFile := filepath.Join(testMediaDir, "Anne Shirley - S01E05 - Let Us Look on the Bright Side of Things HDTV-720p Proper.mkv")

	// Check if test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	files := []string{testFile}

	err = player.PlayPlaylist(files, opts)
	if err != nil {
		t.Fatalf("PlayPlaylist failed: %v", err)
	}

	// Clean up
	player.Stop()
}

func TestPlayPlaylist_DisableSubtitles(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	opts := PlaylistOptions{
		SubtitleTrack: -1, // Disable subtitles
	}

	testFile := filepath.Join(testMediaDir, "Anne Shirley - S01E05 - Let Us Look on the Bright Side of Things HDTV-720p Proper.mkv")

	// Check if test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	files := []string{testFile}

	err = player.PlayPlaylist(files, opts)
	if err != nil {
		t.Fatalf("PlayPlaylist failed: %v", err)
	}

	// Clean up
	player.Stop()
}

func TestPlayPlaylist_MultipleFiles(t *testing.T) {
	player, err := NewMPVPlayer()
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	testFiles := []string{
		filepath.Join(testMediaDir, "Anne Shirley - S01E05 - Let Us Look on the Bright Side of Things HDTV-720p Proper.mkv"),
		filepath.Join(testMediaDir, "Anne Shirley - S01E09 - Next to Trying and Winning, the Best Thing Is Trying and Failing HDTV-720p Proper.mkv"),
	}

	// Check if test files exist
	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Skipf("Test files not found in: %s", testMediaDir)
		}
	}

	opts := PlaylistOptions{
		AudioLanguage: "jpn",
	}

	err = player.PlayPlaylist(testFiles, opts)
	if err != nil {
		t.Fatalf("PlayPlaylist failed: %v", err)
	}

	// Clean up
	player.Stop()
}

// Note: Testing Play() and Wait() with actual video requires MPV to be installed
// and would require integration tests rather than unit tests
