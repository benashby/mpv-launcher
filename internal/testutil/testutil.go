package testutil

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestdataPath returns the absolute path to a file in the testdata directory
// relative to the calling test file's package
func TestdataPath(t *testing.T, filename string) string {
	t.Helper()

	// Get the caller's file path
	_, callerFile, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("failed to get caller information")
	}

	// Get the directory containing the test file
	callerDir := filepath.Dir(callerFile)

	// Construct path to testdata
	return filepath.Join(callerDir, "testdata", filename)
}

// MustLoadFixture loads an ffprobe JSON fixture or fails the test
func MustLoadFixture(t *testing.T, path string) *FFProbeOutput {
	t.Helper()

	output, err := LoadFFProbeJSON(path)
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", path, err)
	}
	return output
}

// CompareStringSlices compares two string slices (order-independent)
func CompareStringSlices(t *testing.T, name string, expected, actual []string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("%s: count mismatch: expected %d, got %d", name, len(expected), len(actual))
		t.Errorf("  expected: %v", expected)
		t.Errorf("  actual: %v", actual)
		return
	}

	expectedSet := make(map[string]bool)
	for _, s := range expected {
		expectedSet[s] = true
	}

	for _, s := range actual {
		if !expectedSet[s] {
			t.Errorf("%s: unexpected value: %s", name, s)
		}
		delete(expectedSet, s)
	}

	for s := range expectedSet {
		t.Errorf("%s: missing value: %s", name, s)
	}
}
