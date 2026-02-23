package collect

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"gopkg.in/music-theory.v0/key"

	"math"

	"git.tcp.direct/kayos/keepr/internal/analysis"
	"git.tcp.direct/kayos/keepr/internal/config"
)

var lockMap = make(map[string]*sync.Mutex)

var mapMu = &sync.RWMutex{}

func guessBPM(piece string) (bpm int) {
	// TODO: don't trust this lol?
	if num, numerr := strconv.Atoi(piece); numerr != nil {
		return num
	}
	frg := strings.Split(piece, "bpm")[0]
	m := strings.Split(frg, "")
	var start = 0
	var numfound = false
	for b, p := range m {
		if _, e := strconv.Atoi(p); e != nil {
			start = b
			continue
		}
		numfound = true
		break
	}
	if !numfound {
		return 0
	}
	var fail error
	if bpm, fail = strconv.Atoi(frg[start:]); fail == nil {
		return bpm
	}
	return 0
}

func guessSeperator(name string) (spl []string) {
	var (
		sep  = " "
		seps = []string{"-", "_", " - "}
	)
	for _, s := range seps {
		if strings.Contains(name, s) {
			sep = s
		}
	}
	log.Trace().Msgf("found seperator for %s: %s", name, sep)
	return strings.Split(name, sep)
}

func (s *Sample) getParentDir() string {
	spl := strings.Split(s.Path, "/")
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(spl[len(spl)-2])), " ", "_")
}

func (s *Sample) IsType(st SampleType) bool {
	_, ok := s.Types[st]
	return ok
}

var drumDirMap = map[string]DrumType{
	"snares": DrumSnare, "snare": DrumSnare, "kick": DrumKick, "kicks": DrumKick, "hats": DrumHiHat, "hat": DrumHiHat,
	"hihat": DrumHiHat, "hi-hat": DrumHiHat, "hihats": DrumHiHat, "closed_hihats": DrumHatClosed,
	"open_hihats": DrumHatOpen, "808s": Drum808, "808": Drum808, "toms": DrumTom,
}

var drumToDirMap = map[DrumType]string{
	DrumSnare: "Snares", DrumKick: "Kicks", DrumHiHat: "HiHats", DrumHatClosed: "HiHats/Closed",
	DrumHatOpen: "HiHats/Open", Drum808: "808", DrumTom: "Toms", DrumPercussion: "Other",
}

var (
	rgxSharpIn, _    = regexp.Compile("[♯#]|major")
	rgxFlatIn, _     = regexp.Compile("^F|[♭b]")
	rgxSharpBegin, _ = regexp.Compile("^[♯#]")
	rgxFlatBegin, _  = regexp.Compile("^[♭b]")
	rgxSharpishIn, _ = regexp.Compile("(maj|major|aug)")
	rgxFlattishIn, _ = regexp.Compile("([^a-z]|^)(m|min|minor|dim)")

	mustMatchOne = map[string]*regexp.Regexp{
		"flat": rgxFlatIn, "flat_begin": rgxFlatBegin,
		"sharp_begin": rgxSharpBegin, "sharp": rgxSharpIn,
		"flattish": rgxFlattishIn, "sparpish": rgxSharpishIn}
)

var keySubstrings = map[string]struct{}{
	"C": {}, "D": {}, "E": {}, "F": {}, "G": {}, "A": {}, "B": {},
}

func keyMatch(s string, opiece string) bool {
	var found bool
	if _, ok := keySubstrings[s]; !ok {
		return false
	}
	for desc, rgx := range mustMatchOne {
		if !rgx.MatchString(opiece) {
			continue
		}
		log.Trace().Msgf("matched regex for %s", desc)
		found = true
		break

	}
	return found
}

var melodicKeywords = []string{
	"chord", "synth", "pad", "arp", "piano",
	"organ", "guitar", "bass", "lead", "key",
	"string", "brass", "woodwind", "flute", "trumpet",
	"sax", "horn", "violin", "cello", "harp",
	"vocal", "marimba",
}

func (s *Sample) ParseFilename() {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	slog := log.With().Str("caller", s.Path).Logger()
	if s.Name != "" {
		slog = slog.With().Str("caller", s.Name).Logger()
	}
	pd := s.getParentDir()
	path := strings.ReplaceAll(strings.TrimSpace(strings.ToLower(s.Path)), " ", "_")
	fname := strings.ToLower(filepath.Base(path))
	candidates := make([]string, 0, 1)
	if strings.Contains(pd, "_") {
		candidates = strings.Split(pd, "_")
	} else {
		candidates = []string{pd}
	}
	for _, c := range candidates {
		if drumtype, isdrum := drumDirMap[c]; isdrum {
			slog.Trace().Msgf("found drum type: %s", c)
			s.Types[TypeDrum] = struct{}{}
			go Library.IngestDrum(s, drumtype)
			break
		}
	}

	if strings.Contains(pd, "melod") {
		s.Types[TypeMelodic] = struct{}{}
	}

	if strings.Contains(pd, "loop") || strings.Contains(fname, "bpm") {
		s.Types[TypeLoop] = struct{}{}
	}

	for _, k := range melodicKeywords {
		if strings.Contains(fname, k) {
			s.Types[TypeMelodic] = struct{}{}
			log.Trace().Msgf("found melodic keyword: %s, in %s", k, fname)
			break
		}
	}

	var fallback = ""
	var keyFound = false

	roots := []string{"C", "D", "E", "F", "G", "A", "B"}

	opieces := guessSeperator(s.Name)
	for _, opiece := range opieces {
		opiece = strings.TrimSuffix(opiece, ".wav")
		for _, r := range roots {
			if strings.TrimSpace(opiece) == r {
				fallback = opiece
			}
		}
	}

	for _, opiece := range opieces {
		opiece = strings.TrimSuffix(opiece, ".wav")
		log.Trace().Msgf("parse %s, piece: %s", s.Name, opiece)
		piece := strings.ToLower(opiece)
		if num, numerr := strconv.Atoi(piece); numerr == nil {
			if num > 50 && num != 808 {
				s.Tempo = num
			}
		}
		if strings.Contains(piece, "bpm") {
			s.Tempo = guessBPM(piece)
		}

		if s.Tempo != 0 {
			// go Library.IngestTempo(s)
		}

		spl := strings.Split(opiece, "")
		if len(spl) > 6 || len(spl) == 0 {
			continue
		}

		// if our fragment starts with a known root note, then try to parse the fragment, else dip-set.
		if !keyMatch(spl[0], opiece) {
			continue
		}

		if s.Key.Root != 0 {
			keyFound = true
		}

		if s.Key.Root == 0 {
			s.Key = key.Of(opiece)
			if s.Key.Root != 0 {
				keyFound = true
				// go Library.IngestKey(s)
			}
		}
	}
	if !keyFound && fallback != "" {
		log.Warn().Msgf("using fallback key for %s: %s", s.Name, fallback)
		s.Key = key.Of(fallback)
		// go Library.IngestKey(s)
	}
}

/*
var wavBufs = sync.Pool{
	New: func() interface{} {
		return &audio.IntBuffer{
			Data: make([]int, 0, 4096),
		}
	},
}
*/
/*
	func GetWavBuffer(pcmLen int64) *audio.IntBuffer {
		ib := wavBufs.Get().(*audio.IntBuffer)
		if ib == nil || ib.Data == nil {
			ib = wavBufs.New().(*audio.IntBuffer)
		}
		if ib == nil {
			panic("failed to get wav buffer")
		}
		if ib.Data == nil {
			ib.Data = make([]int, 0, 4096)
		}
		if cap(ib.Data) < int(pcmLen) {
			ib.Data = make([]int, pcmLen)
		}
		ib.Data = ib.Data[:pcmLen]
		return ib
	}

	func PutWavBuffer(b *audio.IntBuffer) {
		if b == nil {
			return
		}

		wavBufs.Put(b)
	}
*/
func readWAV(s *Sample) error {
	f, err := os.Open(s.Path)
	if err != nil {
		return fmt.Errorf("couldn't open %s: %s", s.Path, err.Error())
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)

	decoder.ReadMetadata()
	if decoder.Err() != nil {
		return decoder.Err()
	}

	if s.Metadata == nil {
		s.Metadata = decoder.Metadata
	}

	s.Duration, err = decoder.Duration()
	if err != nil {
		log.Warn().Caller().Str("caller", s.Name).Err(err).Msg("failed to get duration")
	}

	log.Debug().Caller().Str("caller", s.Name).Msgf("duration: %s", s.Duration.String())

	isLoop := false

	if s.Duration != 0 && s.Duration > 1500*time.Millisecond {
		s.Types[TypeLoop] = struct{}{}
		delete(s.Types, TypeOneShot)
		isLoop = true
	}

	if s.Duration != 0 && s.Duration < 1*time.Second && !isLoop {
		s.Types[TypeOneShot] = struct{}{}
		delete(s.Types, TypeLoop)
	}

	if s.Metadata == nil {
		log.Debug().Caller().Str("caller", s.Name).Msg("no metadata found")
		return nil
	}

	if s.Metadata != nil && s.Metadata.SamplerInfo != nil && len(s.Metadata.SamplerInfo.Loops) > 0 {
		isLoop = true
		s.Types[TypeLoop] = struct{}{}
	}

	log.Trace().Msg(fmt.Sprintf("metadata: %v", s.Metadata))

	// Acoustic verification: override filename guesses with measured audio data
	if !config.SkipWavDecode {
		if buf, pcmErr := decoder.FullPCMBuffer(); pcmErr == nil && buf != nil {
			mono := toMonoFloat32(buf)
			sr := int(buf.Format.SampleRate)
			// BPM
			bpm := analysis.DetectBPM(mono, sr)
			if bpm >= 50 && bpm <= 250 {
				acousticTempo := int(math.Round(bpm))
				if s.Tempo == 0 {
					s.Tempo = acousticTempo
				} else if s.Tempo != acousticTempo {
					log.Warn().Str("caller", s.Name).Msgf("BPM mismatch: filename=%d acoustic=%d, trusting acoustic", s.Tempo, acousticTempo)
					s.Tempo = acousticTempo
				}
			}
			// Key — skip one-shots, too short for reliable detection
			if _, isOneShot := s.Types[TypeOneShot]; !isOneShot {
				detectedKey, _ := analysis.DetectKey(mono, sr, float64(config.AnalyzeSeconds))
				if detectedKey.Root != 0 || detectedKey.Mode != 0 {
					if s.Key.Root == 0 {
						s.Key = detectedKey
					} else if s.Key != detectedKey {
						log.Warn().Str("caller", s.Name).Msgf("key mismatch: filename=%s acoustic=%s, trusting acoustic",
							s.Key.Root.String(s.Key.AdjSymbol), detectedKey.Root.String(detectedKey.AdjSymbol))
						s.Key = detectedKey
					}
				}
			}
		}
	}

	decoder = nil // avoid memory leak

	return nil
}

func Process(entry fs.DirEntry, dir string) (*Sample, error) {
	log.Trace().Str("caller", entry.Name()).Msg("Processing")
	var finfo os.FileInfo
	var err error
	finfo, err = entry.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to Process %s: %s", entry.Name(), err.Error())
	}

	spl := strings.Split(entry.Name(), ".")
	ext := spl[len(spl)-1]

	s := &Sample{
		Name:    entry.Name(),
		Path:    dir,
		ModTime: finfo.ModTime(),
		Types:   make(map[SampleType]struct{}),
	}

	s.ParseFilename()
	defer Library.IngestSample(s)

	switch ext {
	case "midi", "mid":
		if !config.NoMIDI {
			s.Types[TypeMIDI] = struct{}{}
			// Parse MIDI meta events for tempo and key
			if midiTempo, midiKey, midiErr := parseMIDI(s.Path); midiErr == nil {
				if midiTempo > 0 && s.Tempo == 0 {
					s.Tempo = midiTempo
				}
				if midiKey.Root != 0 && s.Key.Root == 0 {
					s.Key = midiKey
				}
			} else {
				log.Debug().Str("caller", s.Name).Err(midiErr).Msg("failed to parse MIDI meta events")
			}
			Library.IngestMIDI(s)
		}

	case "wav":
		if config.SkipWavDecode {
			break
		}
		wavErr := readWAV(s)
		if wavErr != nil {
			log.Debug().Caller().Str("caller", s.Name).Msgf("failed to parse wav data: %s", wavErr.Error())
			return nil, nil
		}
		if s.Metadata == nil {
			break
		}

	default:
		return nil, nil
	}

	return s, err
}

// toMonoFloat32 converts a PCM buffer to normalized mono float32 in [-1,1].
func toMonoFloat32(buf *audio.IntBuffer) []float32 {
	numChannels := buf.Format.NumChannels
	frames := len(buf.Data) / numChannels
	mono := make([]float32, frames)
	var maxVal float64
	switch buf.SourceBitDepth {
	case 8:
		maxVal = 128.0
	case 16:
		maxVal = 32768.0
	case 24:
		maxVal = 8388608.0
	case 32:
		maxVal = 2147483648.0
	default:
		maxVal = 32768.0
	}
	idx := 0
	for i := 0; i < frames; i++ {
		var sum float64
		for ch := 0; ch < numChannels; ch++ {
			sum += float64(buf.Data[idx])
			idx++
		}
		mono[i] = float32((sum / float64(numChannels)) / maxVal)
	}
	return mono
}
