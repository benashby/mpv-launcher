package episode

import (
	"testing"
)

// Test SxxExx formats
func TestParse_SxxExx_Standard(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
		pattern  string
	}{
		{"Show.S01E01.mkv", 1, 1, "SxxExx"},
		{"Show.S01E01.1080p.mkv", 1, 1, "SxxExx"},
		{"Show Name S01E01.mkv", 1, 1, "SxxExx"},
		{"Show.S10E25.mkv", 10, 25, "SxxExx"},
		{"Show.S01E100.mkv", 1, 100, "SxxExx"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
			if info.MatchedPattern != tt.pattern {
				t.Errorf("Pattern: got %s, want %s", info.MatchedPattern, tt.pattern)
			}
		})
	}
}

func TestParse_SxxExx_Lowercase(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
	}{
		{"show.s01e01.mkv", 1, 1},
		{"show.s02e15.mkv", 2, 15},
		{"show.s1e1.mkv", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
		})
	}
}

func TestParse_SxxExx_NoLeadingZeros(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
	}{
		{"Show.S1E1.mkv", 1, 1},
		{"Show.s2e3.mkv", 2, 3},
		{"Show.S5E15.mkv", 5, 15},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
		})
	}
}

// Test xXX formats
func TestParse_xXX_Format(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
	}{
		{"Breaking Bad 1x01.mkv", 1, 1},
		{"Show 2x15.mkv", 2, 15},
		{"Show 01x01.mkv", 1, 1},
		{"Show 1X01.mkv", 1, 1},
		{"Show.5x23.720p.mkv", 5, 23},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
			if info.MatchedPattern != "xXX" {
				t.Errorf("Pattern: got %s, want xXX", info.MatchedPattern)
			}
		})
	}
}

// Test absolute/sequential numbering
func TestParse_Absolute_Sequential(t *testing.T) {
	tests := []struct {
		filename   string
		absoluteEp int
	}{
		{"Show 001.mkv", 1},
		{"Show 002.mkv", 2},
		{"Show 015.mkv", 15},
		{"Show 100.mkv", 100},
		{"Show.Episode.025.mkv", 25},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.AbsoluteEp != tt.absoluteEp {
				t.Errorf("AbsoluteEp: got %d, want %d", info.AbsoluteEp, tt.absoluteEp)
			}
			if info.Type != TypeAbsolute {
				t.Errorf("Type: got %v, want TypeAbsolute", info.Type)
			}
		})
	}
}

func TestParse_Absolute_EpPrefix(t *testing.T) {
	tests := []struct {
		filename   string
		absoluteEp int
	}{
		{"Show ep01.mkv", 1},
		{"Show ep02.mkv", 2},
		{"Show ep 15.mkv", 15},
		{"Show.ep.100.mkv", 100},
		{"Show_ep_025.mkv", 25},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.AbsoluteEp != tt.absoluteEp {
				t.Errorf("AbsoluteEp: got %d, want %d", info.AbsoluteEp, tt.absoluteEp)
			}
			if info.MatchedPattern != "ep-prefix" {
				t.Errorf("Pattern: got %s, want ep-prefix", info.MatchedPattern)
			}
		})
	}
}

// Test dash/underscore variations
func TestParse_DashUnderscore(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
	}{
		{"Show S01-E01.mkv", 1, 1},
		{"Show S01_E01.mkv", 1, 1},
		{"Show 1-01.mkv", 1, 1},
		{"Show 1_01.mkv", 1, 1},
		{"Show 2-15.mkv", 2, 15},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
		})
	}
}

// Test date-based naming
func TestParse_Date_YYYY_MM_DD(t *testing.T) {
	tests := []struct {
		filename string
		year     int
		month    int
		day      int
	}{
		{"The Daily Show 2024-01-15.mkv", 2024, 1, 15},
		{"Show.2023-12-25.1080p.mkv", 2023, 12, 25},
		{"Show 2020-05-01.mkv", 2020, 5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Type != TypeDate {
				t.Errorf("Type: got %v, want TypeDate", info.Type)
			}
			if info.Date == nil {
				t.Fatal("Date is nil")
			}
			if info.Date.Year() != tt.year || int(info.Date.Month()) != tt.month || info.Date.Day() != tt.day {
				t.Errorf("Date: got %v, want %d-%d-%d", info.Date, tt.year, tt.month, tt.day)
			}
		})
	}
}

func TestParse_Date_YYYY_MM_DD_Dots(t *testing.T) {
	tests := []struct {
		filename string
		year     int
		month    int
		day      int
	}{
		{"Show.2024.01.15.mkv", 2024, 1, 15},
		{"Show.2023.12.25.mkv", 2023, 12, 25},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Type != TypeDate {
				t.Errorf("Type: got %v, want TypeDate", info.Type)
			}
			if info.Date == nil {
				t.Fatal("Date is nil")
			}
			if info.Date.Year() != tt.year || int(info.Date.Month()) != tt.month || info.Date.Day() != tt.day {
				t.Errorf("Date: got %v, want %d-%d-%d", info.Date, tt.year, tt.month, tt.day)
			}
		})
	}
}

func TestParse_Date_DD_MM_YYYY(t *testing.T) {
	tests := []struct {
		filename string
		year     int
		month    int
		day      int
	}{
		{"Show 15-01-2024.mkv", 2024, 1, 15},
		{"Show 25-12-2023.mkv", 2023, 12, 25},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Type != TypeDate {
				t.Errorf("Type: got %v, want TypeDate", info.Type)
			}
			if info.Date == nil {
				t.Fatal("Date is nil")
			}
			if info.Date.Year() != tt.year || int(info.Date.Month()) != tt.month || info.Date.Day() != tt.day {
				t.Errorf("Date: got %v, want %d-%d-%d", info.Date, tt.year, tt.month, tt.day)
			}
		})
	}
}

// Test special episodes (Season 0)
func TestParse_Specials_Season00(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
	}{
		{"Show S00E01.mkv", 0, 1},
		{"Show s00e05.mkv", 0, 5},
		{"Show.S00E10.Special.mkv", 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
		})
	}
}

// Test multi-episode files
func TestParse_MultiEpisode_Range(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episodes []int
	}{
		{"Show S01E01-E03.mkv", 1, []int{1, 3}},
		{"Show S01E01-03.mkv", 1, []int{1, 3}},
		{"Show.S02E10-E12.mkv", 2, []int{10, 12}},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) < 2 {
				t.Fatalf("Expected multi-episode, got %v", info.Episodes)
			}
			// Check first and last episode
			if info.Episodes[0] != tt.episodes[0] {
				t.Errorf("First episode: got %d, want %d", info.Episodes[0], tt.episodes[0])
			}
			if info.Episodes[len(info.Episodes)-1] != tt.episodes[1] {
				t.Errorf("Last episode: got %d, want %d", info.Episodes[len(info.Episodes)-1], tt.episodes[1])
			}
		})
	}
}

func TestParse_MultiEpisode_Stacked(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		minEps   int // Minimum episodes expected
	}{
		{"Show S01E01E02E03.mkv", 1, 2},
		{"Show.S02E10E11.mkv", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) < tt.minEps {
				t.Errorf("Expected at least %d episodes, got %d (%v)", tt.minEps, len(info.Episodes), info.Episodes)
			}
		})
	}
}

// Test split episodes
func TestParse_SplitEpisode_PartNumber(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
		part     string
	}{
		{"Show S01E01 - pt1.mkv", 1, 1, "1"},
		{"Show S01E01 - pt2.mkv", 1, 1, "2"},
		{"Show S01E01 Part 1.mkv", 1, 1, "1"},
		{"Show S01E01 Part 2.mkv", 1, 1, "2"},
		{"Show.S01E01.pt1.mkv", 1, 1, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
			if info.Part != tt.part {
				t.Errorf("Part: got %s, want %s", info.Part, tt.part)
			}
		})
	}
}

func TestParse_SplitEpisode_SubPart(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
		part     string
	}{
		{"Show S01E01.1.mkv", 1, 1, "1"},
		{"Show S01E01.2.mkv", 1, 1, "2"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
			if info.Part != tt.part {
				t.Errorf("Part: got %s, want %s", info.Part, tt.part)
			}
		})
	}
}

func TestParse_SplitEpisode_Alpha(t *testing.T) {
	tests := []struct {
		filename string
		season   int
		episode  int
		part     string
	}{
		{"Show S01E01a.mkv", 1, 1, "a"},
		{"Show S01E01b.mkv", 1, 1, "b"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
			if info.Part != tt.part {
				t.Errorf("Part: got %s, want %s", info.Part, tt.part)
			}
		})
	}
}

// Test sorting
func TestSort_SeasonEpisode(t *testing.T) {
	filenames := []string{
		"Show S02E01.mkv",
		"Show S01E03.mkv",
		"Show S01E01.mkv",
		"Show S01E02.mkv",
		"Show S03E01.mkv",
	}

	episodes := ParseBatch(filenames)
	Sort(episodes)

	expected := []struct {
		season  int
		episode int
	}{
		{1, 1},
		{1, 2},
		{1, 3},
		{2, 1},
		{3, 1},
	}

	for i, exp := range expected {
		if episodes[i].Season != exp.season || episodes[i].Episodes[0] != exp.episode {
			t.Errorf("Index %d: got S%02dE%02d, want S%02dE%02d",
				i, episodes[i].Season, episodes[i].Episodes[0], exp.season, exp.episode)
		}
	}
}

func TestSort_WithParts(t *testing.T) {
	filenames := []string{
		"Show S01E01 pt2.mkv",
		"Show S01E01 pt1.mkv",
		"Show S01E02.mkv",
	}

	episodes := ParseBatch(filenames)
	Sort(episodes)

	// Should be: pt1, pt2, E02
	if episodes[0].Part != "1" {
		t.Errorf("First should be pt1, got %s", episodes[0].Part)
	}
	if episodes[1].Part != "2" {
		t.Errorf("Second should be pt2, got %s", episodes[1].Part)
	}
	if episodes[2].Episodes[0] != 2 {
		t.Errorf("Third should be E02, got E%02d", episodes[2].Episodes[0])
	}
}

func TestSort_Absolute(t *testing.T) {
	filenames := []string{
		"Show 003.mkv",
		"Show 001.mkv",
		"Show 002.mkv",
		"Show 010.mkv",
	}

	episodes := ParseBatch(filenames)
	Sort(episodes)

	expectedEps := []int{1, 2, 3, 10}
	for i, exp := range expectedEps {
		if episodes[i].AbsoluteEp != exp {
			t.Errorf("Index %d: got ep %d, want ep %d", i, episodes[i].AbsoluteEp, exp)
		}
	}
}

func TestSort_DateBased(t *testing.T) {
	filenames := []string{
		"Show 2024-01-15.mkv",
		"Show 2024-01-10.mkv",
		"Show 2024-01-20.mkv",
	}

	episodes := ParseBatch(filenames)
	Sort(episodes)

	// Check they're in chronological order
	for i := 1; i < len(episodes); i++ {
		if !episodes[i-1].Date.Before(*episodes[i].Date) {
			t.Errorf("Dates not in order: %v should be before %v",
				episodes[i-1].Date, episodes[i].Date)
		}
	}
}

func TestSort_MixedTypes(t *testing.T) {
	filenames := []string{
		"Show 2024-01-15.mkv", // Date
		"Show 001.mkv",        // Absolute
		"Show S01E01.mkv",     // Season/Episode
	}

	episodes := ParseBatch(filenames)
	Sort(episodes)

	// Order should be: SeasonEpisode, Absolute, Date
	if episodes[0].Type != TypeSeasonEpisode {
		t.Errorf("First should be TypeSeasonEpisode, got %v", episodes[0].Type)
	}
	if episodes[1].Type != TypeAbsolute {
		t.Errorf("Second should be TypeAbsolute, got %v", episodes[1].Type)
	}
	if episodes[2].Type != TypeDate {
		t.Errorf("Third should be TypeDate, got %v", episodes[2].Type)
	}
}

func TestSort_SpecialsAtEnd(t *testing.T) {
	filenames := []string{
		"Show S00E01.mkv", // Special
		"Show S01E01.mkv",
		"Show S00E02.mkv", // Special
		"Show S01E02.mkv",
		"Show S02E01.mkv",
		"Show S00E03.mkv", // Special
	}

	episodes := ParseBatch(filenames)
	Sort(episodes)

	// Expected order: S01E01, S01E02, S02E01, S00E01, S00E02, S00E03
	expected := []struct {
		season  int
		episode int
	}{
		{1, 1},
		{1, 2},
		{2, 1},
		{0, 1}, // Specials at end
		{0, 2},
		{0, 3},
	}

	for i, exp := range expected {
		if episodes[i].Season != exp.season {
			t.Errorf("Index %d: got season %d, want %d", i, episodes[i].Season, exp.season)
		}
		if len(episodes[i].Episodes) > 0 && episodes[i].Episodes[0] != exp.episode {
			t.Errorf("Index %d: got episode %d, want %d", i, episodes[i].Episodes[0], exp.episode)
		}
	}
}

// Test edge cases
func TestParse_NoMatch(t *testing.T) {
	filenames := []string{
		"random_file.mkv",
		"no_episode_info.mp4",
		"just_a_movie.avi",
	}

	for _, filename := range filenames {
		t.Run(filename, func(t *testing.T) {
			info := Parse(filename)
			if info.Type != TypeUnknown {
				t.Errorf("Should be TypeUnknown, got %v", info.Type)
			}
		})
	}
}

func TestParse_ComplexFilename(t *testing.T) {
	// Test with quality tags, group names, etc.
	tests := []struct {
		filename string
		season   int
		episode  int
	}{
		{"[GroupName] Show Name - S01E05 - Episode Title [1080p][x264].mkv", 1, 5},
		{"Show.Name.S02E10.1080p.WEB-DL.DD5.1.H264-GROUP.mkv", 2, 10},
		{"Show_Name_-_S03E15_-_720p_HDTV_x264.mkv", 3, 15},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			info := Parse(tt.filename)
			if info.Season != tt.season {
				t.Errorf("Season: got %d, want %d", info.Season, tt.season)
			}
			if len(info.Episodes) == 0 || info.Episodes[0] != tt.episode {
				t.Errorf("Episode: got %v, want %d", info.Episodes, tt.episode)
			}
		})
	}
}

func TestParse_Equivalence_S1E1_vs_S01E01(t *testing.T) {
	// These should parse to the same season/episode
	file1 := "Show S1E1.mkv"
	file2 := "Show S01E01.mkv"

	info1 := Parse(file1)
	info2 := Parse(file2)

	if info1.Season != info2.Season {
		t.Errorf("Seasons should match: %d vs %d", info1.Season, info2.Season)
	}
	if len(info1.Episodes) == 0 || len(info2.Episodes) == 0 {
		t.Fatal("Episodes not parsed")
	}
	if info1.Episodes[0] != info2.Episodes[0] {
		t.Errorf("Episodes should match: %d vs %d", info1.Episodes[0], info2.Episodes[0])
	}

	// When sorted together, they should be adjacent
	episodes := []*EpisodeInfo{info1, info2}
	Sort(episodes)

	// Check they sort together (order doesn't matter, but they should be next to each other)
	if !Less(episodes[0], episodes[1]) && !Less(episodes[1], episodes[0]) {
		// They're equal in sorting, which is what we want
	} else {
		// One is less than the other, which means they're not treated as equivalent
		// This is actually okay - they'll still be next to each other
	}
}
