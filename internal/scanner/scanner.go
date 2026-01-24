package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// VideoExtensions contains all common and uncommon video file extensions
var VideoExtensions = map[string]bool{
	// Common formats
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
	".m4v":  true,
	".mpg":  true,
	".mpeg": true,
	".m2v":  true,
	".3gp":  true,
	".3g2":  true,
	".ogv":  true,
	".ts":   true,
	".mts":  true,
	".m2ts": true,

	// Uncommon/specialized formats
	".vob":  true,
	".asf":  true,
	".rm":   true,
	".rmvb": true,
	".divx": true,
	".dv":   true,
	".f4v":  true,
	".mxf":  true,
	".roq":  true,
	".yuv":  true,
	".nsv":  true,
	".gxf":  true,
	".qt":   true,
	".xvid": true,

	// Additional formats
	".mp2":  true,
	".mpe":  true,
	".mpv":  true,
	".m4p":  true,
	".m4b":  true,
	".dat":  true, // VCD format
	".vcd":  true,
	".svcd": true,
	".drc":  true,
	".gif":  true, // Animated GIFs
	".gifv": true,
	".mng":  true,
	".viv":  true,
	".amv":  true,
	".m2p":  true,
	".m2t":  true,
	".m4s":  true,
	".tod":  true,
	".vro":  true,
	".wtv":  true,
}

// ScanOptions configures how the scanner should operate
type ScanOptions struct {
	// Recursive determines whether to scan subdirectories
	Recursive bool
	// FollowSymlinks determines whether to follow symbolic links
	FollowSymlinks bool
}

// IsVideoFile checks if a file has a video extension
func IsVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return VideoExtensions[ext]
}

// ScanDirectory scans a directory for video files
// Returns a slice of absolute paths to video files
func ScanDirectory(dirPath string, opts ScanOptions) ([]string, error) {
	// Verify directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dirPath)
	}

	var videoFiles []string

	if opts.Recursive {
		videoFiles, err = scanRecursive(dirPath, opts)
	} else {
		videoFiles, err = scanNonRecursive(dirPath, opts)
	}

	if err != nil {
		return nil, err
	}

	return videoFiles, nil
}

// scanRecursive performs a recursive directory scan for video files
func scanRecursive(dirPath string, opts ScanOptions) ([]string, error) {
	var videoFiles []string

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't access
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle symlinks
		if d.Type()&fs.ModeSymlink != 0 {
			if !opts.FollowSymlinks {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Follow symlink and check if it's a directory
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil // Skip broken symlinks
			}

			info, err := os.Stat(realPath)
			if err != nil {
				return nil
			}

			// If symlink points to a directory, skip it to avoid cycles
			if info.IsDir() {
				return filepath.SkipDir
			}
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if it's a video file
		if IsVideoFile(d.Name()) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil
			}
			videoFiles = append(videoFiles, absPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return videoFiles, nil
}

// scanNonRecursive performs a non-recursive directory scan for video files
func scanNonRecursive(dirPath string, opts ScanOptions) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var videoFiles []string

	for _, entry := range entries {
		// Skip subdirectories
		if entry.IsDir() {
			continue
		}

		// Handle symlinks
		if entry.Type()&fs.ModeSymlink != 0 {
			if !opts.FollowSymlinks {
				continue
			}

			// Follow symlink and check if it points to a file
			fullPath := filepath.Join(dirPath, entry.Name())
			realPath, err := filepath.EvalSymlinks(fullPath)
			if err != nil {
				continue // Skip broken symlinks
			}

			info, err := os.Stat(realPath)
			if err != nil || info.IsDir() {
				continue // Skip if it's a directory or can't stat
			}
		}

		// Check if it's a video file
		if IsVideoFile(entry.Name()) {
			absPath, err := filepath.Abs(filepath.Join(dirPath, entry.Name()))
			if err != nil {
				continue
			}
			videoFiles = append(videoFiles, absPath)
		}
	}

	return videoFiles, nil
}

// ScanMultipleDirectories scans multiple directories and returns combined results
func ScanMultipleDirectories(dirPaths []string, opts ScanOptions) ([]string, error) {
	var allVideoFiles []string
	seen := make(map[string]bool)

	for _, dirPath := range dirPaths {
		files, err := ScanDirectory(dirPath, opts)
		if err != nil {
			return nil, fmt.Errorf("error scanning %s: %w", dirPath, err)
		}

		// Deduplicate files (in case of overlapping paths)
		for _, file := range files {
			if !seen[file] {
				seen[file] = true
				allVideoFiles = append(allVideoFiles, file)
			}
		}
	}

	return allVideoFiles, nil
}
