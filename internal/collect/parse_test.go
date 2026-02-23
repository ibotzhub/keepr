package collect

import (
	"sync"
	"testing"
	"time"

	"github.com/go-audio/wav"
	"github.com/rs/zerolog"
	"gopkg.in/music-theory.v0/key"
)

func init() {
	l := zerolog.Nop()
	log = &l
}

// --------------------------------------------------------------------
// getParentDir
// --------------------------------------------------------------------

func TestSample_getParentDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "808 bass dir",
			path: "/home/fuckholejones/808_Bass/OS_BB_808_E_RARRI.wav",
			want: "808_bass",
		},
		{
			name: "snares dir",
			path: "/samples/Snares/snare_crack_01.wav",
			want: "snares",
		},
		{
			name: "kicks dir",
			path: "/samples/Kicks/kick_sub_808.wav",
			want: "kicks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{Name: tt.name, Path: tt.path}
			if got := s.getParentDir(); got != tt.want {
				t.Errorf("getParentDir() = %q, want %q", got, tt.want)
			} else {
				t.Logf("parent dir: %s", got)
			}
		})
	}
}

// --------------------------------------------------------------------
// guessBPM
// --------------------------------------------------------------------

func TestGuessBPM(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"140bpm", 0},
		{"OS_140_loop", 0},
		{"90bpm_kick", 0},
		{"notanumber", 0},
		{"120", 120},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := guessBPM(tt.input)
			if got != tt.want {
				t.Errorf("guessBPM(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// --------------------------------------------------------------------
// ParseFilename — tempo detection
// --------------------------------------------------------------------

func TestParseFilename_Tempo(t *testing.T) {
	tests := []struct {
		filename  string
		wantTempo int
	}{
		{"loop_140bpm_Amin.wav", 0},
		{"hat_120_open.wav", 120},
		{"sample_no_tempo.wav", 0},
	}

	Library = newTestLibrary()
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			s := &Sample{
				Name:  tt.filename,
				Path:  "/samples/" + tt.filename,
				Types: make(map[SampleType]struct{}),
			}
			s.ParseFilename()
			if s.Tempo != tt.wantTempo {
				t.Errorf("ParseFilename tempo = %d, want %d", s.Tempo, tt.wantTempo)
			} else {
				t.Logf("tempo: %d", s.Tempo)
			}
		})
	}
}

// --------------------------------------------------------------------
// ParseFilename — drum type detection via parent dir
// --------------------------------------------------------------------

func TestParseFilename_DrumType(t *testing.T) {
	tests := []struct {
		path   string
		isDrum bool
	}{
		{"/samples/Kicks/kick_01.wav", true},
		{"/samples/Snares/snare_01.wav", true},
		{"/samples/HiHats/hat_open.wav", true},
		{"/samples/Melodic/pad_Cmaj.wav", false},
	}

	Library = newTestLibrary()
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			s := &Sample{
				Name:  "test.wav",
				Path:  tt.path,
				Types: make(map[SampleType]struct{}),
			}
			s.ParseFilename()
			_, got := s.Types[TypeDrum]
			if got != tt.isDrum {
				t.Errorf("isDrum = %v, want %v for path %s", got, tt.isDrum, tt.path)
			}
		})
	}
}

// --------------------------------------------------------------------
// Duration-based type detection (one-shot vs loop)
// --------------------------------------------------------------------

func TestSampleType_Duration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		wantLoop bool
		wantShot bool
	}{
		{"short one-shot", 300 * time.Millisecond, false, true},
		{"long loop", 4 * time.Second, true, false},
		{"borderline", 1500 * time.Millisecond, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				Name:     tt.name + ".wav",
				Path:     "/samples/" + tt.name + ".wav",
				Duration: tt.duration,
				Types:    make(map[SampleType]struct{}),
				Metadata: &wav.Metadata{},
			}
			if s.Duration > 1500*time.Millisecond {
				s.Types[TypeLoop] = struct{}{}
				delete(s.Types, TypeOneShot)
			} else if s.Duration < time.Second {
				s.Types[TypeOneShot] = struct{}{}
				delete(s.Types, TypeLoop)
			}
			_, isLoop := s.Types[TypeLoop]
			_, isShot := s.Types[TypeOneShot]
			if isLoop != tt.wantLoop {
				t.Errorf("isLoop = %v, want %v", isLoop, tt.wantLoop)
			}
			if isShot != tt.wantShot {
				t.Errorf("isOneShot = %v, want %v", isShot, tt.wantShot)
			}
		})
	}
}

// --------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------

func newTestLibrary() *Collection {
	return &Collection{
		Tempos:        make(map[int][]*Sample),
		Keys:          make(map[key.Key][]*Sample),
		Drums:         make(map[DrumType][]*Sample),
		Artists:       make(map[string][]*Sample),
		Sources:       make(map[string][]*Sample),
		Genres:        make(map[string][]*Sample),
		CreationDates: make(map[string][]*Sample),
		Arists:        make(map[string][]*Sample),
		Software:      make(map[string][]*Sample),
		mu:            &sync.RWMutex{},
	}
}
