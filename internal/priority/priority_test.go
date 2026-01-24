package priority

import (
	"testing"

	"mpv-launcher/internal/tracks"
)

func TestDefaultPermutations(t *testing.T) {
	perms := DefaultPermutations()

	if len(perms) != 4 {
		t.Errorf("Expected 4 permutations, got %d", len(perms))
	}

	// Check first permutation is Japanese + Eng Subtitles
	if perms[0].ID != "jpn_eng" {
		t.Errorf("Expected first permutation ID to be 'jpn_eng', got '%s'", perms[0].ID)
	}
	if perms[0].AudioLang != "jpn" {
		t.Errorf("Expected first permutation audio to be 'jpn', got '%s'", perms[0].AudioLang)
	}
	if perms[0].SubtitleLang != "eng" {
		t.Errorf("Expected first permutation subtitle to be 'eng', got '%s'", perms[0].SubtitleLang)
	}

	// Check second permutation is English - No Subtitles
	if perms[1].ID != "eng_none" {
		t.Errorf("Expected second permutation ID to be 'eng_none', got '%s'", perms[1].ID)
	}
	if perms[1].SubtitleLang != "" {
		t.Errorf("Expected second permutation subtitle to be empty, got '%s'", perms[1].SubtitleLang)
	}
}

func TestPermutation_CanSatisfy(t *testing.T) {
	tests := []struct {
		name     string
		perm     Permutation
		fileInfo tracks.FileTrackInfo
		want     bool
	}{
		{
			name: "jpn_eng satisfied by dual audio with eng subs",
			perm: Permutation{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng"},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "jpn"},
					{MpvIndex: 2, Language: "eng"},
				},
				SubtitleTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
			},
			want: true,
		},
		{
			name: "jpn_eng not satisfied by missing jpn audio",
			perm: Permutation{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng"},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
				SubtitleTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
			},
			want: false,
		},
		{
			name: "jpn_eng not satisfied by missing eng subs",
			perm: Permutation{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng"},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "jpn"},
				},
				SubtitleTracks: []tracks.Track{
					{MpvIndex: 1, Language: "fre"},
				},
			},
			want: false,
		},
		{
			name: "eng_none satisfied by eng audio only",
			perm: Permutation{ID: "eng_none", AudioLang: "eng", SubtitleLang: ""},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
				SubtitleTracks: []tracks.Track{},
			},
			want: true,
		},
		{
			name: "eng_none satisfied even with subs available",
			perm: Permutation{ID: "eng_none", AudioLang: "eng", SubtitleLang: ""},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
				SubtitleTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.perm.CanSatisfy(&tt.fileInfo)
			if got != tt.want {
				t.Errorf("CanSatisfy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermutation_GetTracks(t *testing.T) {
	tests := []struct {
		name       string
		perm       Permutation
		fileInfo   tracks.FileTrackInfo
		wantAudio  int
		wantSub    int
		wantOk     bool
	}{
		{
			name: "get jpn audio and eng sub",
			perm: Permutation{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng"},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "jpn"},
					{MpvIndex: 2, Language: "eng"},
				},
				SubtitleTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
			},
			wantAudio: 1,
			wantSub:   1,
			wantOk:    true,
		},
		{
			name: "get eng audio with no sub",
			perm: Permutation{ID: "eng_none", AudioLang: "eng", SubtitleLang: ""},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "jpn"},
					{MpvIndex: 2, Language: "eng"},
				},
				SubtitleTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
			},
			wantAudio: 2,
			wantSub:   -1,
			wantOk:    true,
		},
		{
			name: "fail when audio not found",
			perm: Permutation{ID: "jpn_eng", AudioLang: "jpn", SubtitleLang: "eng"},
			fileInfo: tracks.FileTrackInfo{
				AudioTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
				SubtitleTracks: []tracks.Track{
					{MpvIndex: 1, Language: "eng"},
				},
			},
			wantAudio: 0,
			wantSub:   0,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audio, sub, ok := tt.perm.GetTracks(&tt.fileInfo)
			if ok != tt.wantOk {
				t.Errorf("GetTracks() ok = %v, want %v", ok, tt.wantOk)
			}
			if audio != tt.wantAudio {
				t.Errorf("GetTracks() audio = %v, want %v", audio, tt.wantAudio)
			}
			if sub != tt.wantSub {
				t.Errorf("GetTracks() sub = %v, want %v", sub, tt.wantSub)
			}
		})
	}
}
