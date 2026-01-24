package episode

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// EpisodeType represents the type of episode numbering detected
type EpisodeType int

const (
	TypeUnknown EpisodeType = iota
	TypeSeasonEpisode
	TypeAbsolute
	TypeDate
)

// EpisodeInfo contains parsed information about a video file's episode
type EpisodeInfo struct {
	Filename       string      // Original filename
	Season         int         // Season number (0 for specials, -1 for unknown)
	Episodes       []int       // Episode number(s) - multiple for multi-episode files
	Part           string      // Part identifier for split episodes ("1", "2", "a", "b", etc.)
	Date           *time.Time  // For date-based episodes
	AbsoluteEp     int         // For absolute numbering (-1 if not applicable)
	Type           EpisodeType // Type of episode numbering detected
	MatchedPattern string      // Which pattern matched (for debugging)
}

// Pattern definitions for various naming conventions
var (
	// SxxExx formats (S01E01, s01e01, S1E1)
	patternSeasonEpisode = regexp.MustCompile(`(?i)s(\d{1,2})e(\d{1,3})`)

	// Multi-episode SxxExx (S01E01-E03, S01E01-03, S01E01E02E03)
	patternMultiEpisode = regexp.MustCompile(`(?i)s\d{1,2}e\d{1,3}(?:(?:[-_]?e?\d{1,3})+|(?:e\d{1,3})+)`)

	// xXX format (1x01, 01x01)
	patternSeasonX = regexp.MustCompile(`(?i)(\d{1,2})x(\d{1,3})`)

	// Dash/underscore variations (S01-E01, S01_E01, 1-01, 1_01)
	patternDashUnderscore = regexp.MustCompile(`(?i)s?(\d{1,2})[-_]e?(\d{1,3})`)

	// Absolute/sequential numbering (001, 002, ep01, ep02)
	patternAbsolute = regexp.MustCompile(`(?i)(?:^|[^\d])(\d{2,3})(?:[^\d]|$)`)
	patternEpPrefix = regexp.MustCompile(`(?i)ep[._\s-]?(\d{1,3})`)

	// Date-based naming (YYYY-MM-DD, YYYY.MM.DD)
	patternDateDash = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	patternDateDot  = regexp.MustCompile(`(\d{4})\.(\d{2})\.(\d{2})`)
	patternDateEU   = regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`) // DD-MM-YYYY

	// Split episode patterns (pt1, part 1, .1, a, b)
	patternPartNum   = regexp.MustCompile(`(?i)(?:pt|part)[\s._-]?(\d+)`)
	patternPartAlpha = regexp.MustCompile(`(?i)(?:pt|part)[\s._-]?([a-z])`)
	patternSubPart   = regexp.MustCompile(`(?i)e\d+\.(\d+)`)      // S01E01.1
	patternSubAlpha  = regexp.MustCompile(`(?i)e\d+([a-z])\b`)     // S01E01a
)

// Parse attempts to extract episode information from a filename
func Parse(filename string) *EpisodeInfo {
	info := &EpisodeInfo{
		Filename:   filename,
		Season:     -1,
		Episodes:   []int{},
		AbsoluteEp: -1,
		Type:       TypeUnknown,
	}

	base := filepath.Base(filename)
	baseNoExt := strings.TrimSuffix(base, filepath.Ext(base))

	// Try patterns in order of specificity (most specific first)
	// 1. Try date-based formats first (before dash/underscore to avoid false matches)
	if parseDate(baseNoExt, info) {
		return info
	}

	// 2. Try multi-episode patterns
	if parseMultiEpisode(baseNoExt, info) {
		parsePart(baseNoExt, info)
		return info
	}

	// 3. Try SxxExx format
	if parseSeasonEpisode(baseNoExt, info) {
		parsePart(baseNoExt, info)
		return info
	}

	// 4. Try xXX format
	if parseSeasonX(baseNoExt, info) {
		parsePart(baseNoExt, info)
		return info
	}

	// 5. Try dash/underscore variations
	if parseDashUnderscore(baseNoExt, info) {
		parsePart(baseNoExt, info)
		return info
	}

	// 6. Try ep-prefix format
	if parseEpPrefix(baseNoExt, info) {
		parsePart(baseNoExt, info)
		return info
	}

	// 7. Try absolute numbering (last resort for numbered episodes)
	if parseAbsolute(baseNoExt, info) {
		parsePart(baseNoExt, info)
		return info
	}

	return info
}

// parseSeasonEpisode parses SxxExx format
func parseSeasonEpisode(s string, info *EpisodeInfo) bool {
	matches := patternSeasonEpisode.FindStringSubmatch(s)
	if len(matches) >= 3 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		info.Season = season
		info.Episodes = []int{episode}
		info.Type = TypeSeasonEpisode
		info.MatchedPattern = "SxxExx"
		return true
	}
	return false
}

// parseMultiEpisode parses multi-episode formats
func parseMultiEpisode(s string, info *EpisodeInfo) bool {
	// First check if this is a multi-episode file
	if !patternMultiEpisode.MatchString(s) {
		return false
	}

	// Extract season and first episode using standard pattern
	matches := patternSeasonEpisode.FindStringSubmatch(s)
	if len(matches) < 3 {
		return false
	}

	season, _ := strconv.Atoi(matches[1])
	firstEp, _ := strconv.Atoi(matches[2])

	info.Season = season
	info.Episodes = []int{firstEp}
	info.Type = TypeSeasonEpisode
	info.MatchedPattern = "SxxExx-Multi"

	// Extract all episode numbers after the first one
	// Look for: E02, E03, E02E03, etc.
	// Find all "E##" patterns
	epPattern := regexp.MustCompile(`(?i)e(\d{1,3})`)
	allEps := epPattern.FindAllStringSubmatch(s, -1)

	seen := make(map[int]bool)
	seen[firstEp] = true

	// Process all E## matches, skipping duplicates
	for _, match := range allEps {
		if len(match) >= 2 {
			ep, _ := strconv.Atoi(match[1])
			if ep > 0 && !seen[ep] {
				info.Episodes = append(info.Episodes, ep)
				seen[ep] = true
			}
		}
	}

	// Also look for dash-number format like "-03" (without the E)
	dashNumPattern := regexp.MustCompile(`(?i)e\d{1,3}[-_](\d{1,3})`)
	dashMatches := dashNumPattern.FindAllStringSubmatch(s, -1)
	for _, match := range dashMatches {
		if len(match) >= 2 {
			ep, _ := strconv.Atoi(match[1])
			if ep > 0 && !seen[ep] {
				info.Episodes = append(info.Episodes, ep)
				seen[ep] = true
			}
		}
	}

	return len(info.Episodes) > 1
}

// parseSeasonX parses 1x01 format
func parseSeasonX(s string, info *EpisodeInfo) bool {
	matches := patternSeasonX.FindStringSubmatch(s)
	if len(matches) >= 3 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		info.Season = season
		info.Episodes = []int{episode}
		info.Type = TypeSeasonEpisode
		info.MatchedPattern = "xXX"
		return true
	}
	return false
}

// parseDashUnderscore parses S01-E01, 1-01, etc.
func parseDashUnderscore(s string, info *EpisodeInfo) bool {
	matches := patternDashUnderscore.FindStringSubmatch(s)
	if len(matches) >= 3 {
		season, _ := strconv.Atoi(matches[1])
		episode, _ := strconv.Atoi(matches[2])
		info.Season = season
		info.Episodes = []int{episode}
		info.Type = TypeSeasonEpisode
		info.MatchedPattern = "Dash/Underscore"
		return true
	}
	return false
}

// parseEpPrefix parses ep01, ep02, etc.
func parseEpPrefix(s string, info *EpisodeInfo) bool {
	matches := patternEpPrefix.FindStringSubmatch(s)
	if len(matches) >= 2 {
		episode, _ := strconv.Atoi(matches[1])
		info.AbsoluteEp = episode
		info.Episodes = []int{episode}
		info.Type = TypeAbsolute
		info.MatchedPattern = "ep-prefix"
		return true
	}
	return false
}

// parseAbsolute parses 001, 002, etc.
func parseAbsolute(s string, info *EpisodeInfo) bool {
	matches := patternAbsolute.FindStringSubmatch(s)
	if len(matches) >= 2 {
		episode, _ := strconv.Atoi(matches[1])
		if episode > 0 {
			info.AbsoluteEp = episode
			info.Episodes = []int{episode}
			info.Type = TypeAbsolute
			info.MatchedPattern = "Absolute"
			return true
		}
	}
	return false
}

// parseDate parses date-based naming
func parseDate(s string, info *EpisodeInfo) bool {
	// Try YYYY-MM-DD
	matches := patternDateDash.FindStringSubmatch(s)
	if len(matches) >= 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		info.Date = &date
		info.Type = TypeDate
		info.MatchedPattern = "Date-YYYY-MM-DD"
		return true
	}

	// Try YYYY.MM.DD
	matches = patternDateDot.FindStringSubmatch(s)
	if len(matches) >= 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		info.Date = &date
		info.Type = TypeDate
		info.MatchedPattern = "Date-YYYY.MM.DD"
		return true
	}

	// Try DD-MM-YYYY (European)
	matches = patternDateEU.FindStringSubmatch(s)
	if len(matches) >= 4 {
		day, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		year, _ := strconv.Atoi(matches[3])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		info.Date = &date
		info.Type = TypeDate
		info.MatchedPattern = "Date-DD-MM-YYYY"
		return true
	}

	return false
}

// parsePart extracts part information for split episodes
func parsePart(s string, info *EpisodeInfo) {
	// Try numeric part (pt1, part 1)
	matches := patternPartNum.FindStringSubmatch(s)
	if len(matches) >= 2 {
		info.Part = matches[1]
		return
	}

	// Try alpha part (pta, part a)
	matches = patternPartAlpha.FindStringSubmatch(s)
	if len(matches) >= 2 {
		info.Part = matches[1]
		return
	}

	// Try sub-part number (S01E01.1)
	matches = patternSubPart.FindStringSubmatch(s)
	if len(matches) >= 2 {
		info.Part = matches[1]
		return
	}

	// Try sub-part alpha (S01E01a)
	matches = patternSubAlpha.FindStringSubmatch(s)
	if len(matches) >= 2 {
		info.Part = matches[1]
		return
	}
}

// ParseBatch parses multiple filenames and returns their episode information
func ParseBatch(filenames []string) []*EpisodeInfo {
	episodes := make([]*EpisodeInfo, len(filenames))
	for i, filename := range filenames {
		episodes[i] = Parse(filename)
	}
	return episodes
}

// Sort sorts a slice of EpisodeInfo in the correct viewing order
func Sort(episodes []*EpisodeInfo) {
	sort.SliceStable(episodes, func(i, j int) bool {
		return Less(episodes[i], episodes[j])
	})
}

// Less compares two episodes for sorting
func Less(a, b *EpisodeInfo) bool {
	// Different types sort in this order: SeasonEpisode, Absolute, Date, Unknown
	if a.Type != b.Type {
		return a.Type < b.Type
	}

	switch a.Type {
	case TypeSeasonEpisode:
		// Sort by season, then episode, then part
		// Treat Season 0 (specials) as infinity so they sort to the end
		if a.Season != b.Season {
			if a.Season == 0 {
				return false // a is special, goes after b
			}
			if b.Season == 0 {
				return true // b is special, a goes before
			}
			return a.Season < b.Season
		}
		if len(a.Episodes) > 0 && len(b.Episodes) > 0 {
			if a.Episodes[0] != b.Episodes[0] {
				return a.Episodes[0] < b.Episodes[0]
			}
		}
		return a.Part < b.Part

	case TypeAbsolute:
		// Sort by absolute episode number, then part
		if a.AbsoluteEp != b.AbsoluteEp {
			return a.AbsoluteEp < b.AbsoluteEp
		}
		return a.Part < b.Part

	case TypeDate:
		// Sort by date
		if a.Date != nil && b.Date != nil {
			return a.Date.Before(*b.Date)
		}
		return false

	default:
		// Unknown types sort by filename
		return a.Filename < b.Filename
	}
}
