// Package analysis provides acoustic BPM and musical key detection
// from raw PCM audio samples (mono float32, normalized to [-1,1]).
//
// started by yunginnanet/kayos. finished by lifelessai/ibot. kayos+ibot 5evr.
package analysis

import (
	"math"
	"sort"

	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
	"gopkg.in/music-theory.v0/key"
)

func DetectBPM(samples []float32, sampleRate int) float64 {
	envelope := make([]float64, len(samples))
	for i, s := range samples {
		envelope[i] = math.Abs(float64(s))
	}
	targetRate := 200
	downsampleFactor := sampleRate / targetRate
	if downsampleFactor < 1 {
		downsampleFactor = 1
	}
	envDown := make([]float64, 0, len(envelope)/downsampleFactor)
	for i := 0; i < len(envelope); i += downsampleFactor {
		var sum float64
		count := 0
		for j := i; j < i+downsampleFactor && j < len(envelope); j++ {
			sum += envelope[j]
			count++
		}
		envDown = append(envDown, sum/float64(count))
	}
	newFs := sampleRate / downsampleFactor
	minBPM, maxBPM := 60.0, 200.0
	minLag := int(float64(newFs) * 60.0 / maxBPM)
	maxLag := int(float64(newFs) * 60.0 / minBPM)
	autocorr := make([]float64, maxLag-minLag)
	for lag := minLag; lag < maxLag; lag++ {
		var sum float64
		for i := 0; i < len(envDown)-lag; i++ {
			sum += envDown[i] * envDown[i+lag]
		}
		autocorr[lag-minLag] = sum / float64(len(envDown)-lag)
	}
	bestIdx, bestVal := 0, autocorr[0]
	for i, v := range autocorr {
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	lag := minLag + bestIdx
	if lag == 0 {
		return 0
	}
	return 60.0 * float64(newFs) / float64(lag)
}

var (
	majorProfile = []float64{6.35, 2.23, 3.48, 2.33, 4.38, 4.09, 2.52, 5.19, 2.39, 3.66, 2.29, 2.88}
	minorProfile = []float64{6.33, 2.68, 3.52, 5.38, 2.60, 3.53, 2.54, 4.75, 3.98, 2.69, 3.34, 3.17}
	noteNamesForKey = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
)

func freqToMIDI(freq float64) float64 {
	return 12*math.Log2(freq/440.0) + 69
}

func whitenSpectrum(mag []float64, windowSize int) []float64 {
	whitened := make([]float64, len(mag))
	half := windowSize / 2
	for i := range mag {
		start := i - half
		if start < 0 { start = 0 }
		end := i + half + 1
		if end > len(mag) { end = len(mag) }
		var sum float64
		for j := start; j < end; j++ { sum += mag[j] }
		avg := sum / float64(end-start)
		if avg > 1e-9 {
			whitened[i] = mag[i] / avg
		} else {
			whitened[i] = mag[i]
		}
	}
	return whitened
}

func EstimateTuning(magnitudes, freqs []float64) float64 {
	var peaks []float64
	for i := 1; i < len(magnitudes)-1; i++ {
		if magnitudes[i] > magnitudes[i-1] && magnitudes[i] > magnitudes[i+1] && magnitudes[i] > 1e-3 {
			peaks = append(peaks, freqs[i])
		}
	}
	if len(peaks) == 0 { return 0 }
	const binWidth = 5.0
	hist := make(map[int]float64)
	for _, f := range peaks {
		midi := 12*math.Log2(f/440.0) + 69
		note := math.Round(midi)
		expectedFreq := 440.0 * math.Pow(2, (note-69)/12)
		cents := 1200 * math.Log2(f/expectedFreq)
		hist[int(math.Round(cents/binWidth))]++
	}
	var bestBin int
	var bestCount float64
	for bin, count := range hist {
		if count > bestCount { bestCount = count; bestBin = bin }
	}
	return float64(bestBin) * binWidth
}

func ComputeChroma(samples []float32, sampleRate int, maxSeconds float64) []float64 {
	maxSamples := int(float64(sampleRate) * maxSeconds)
	if len(samples) > maxSamples { samples = samples[:maxSamples] }
	if len(samples) == 0 { return make([]float64, 12) }
	frameSize := 4096
	hopSize := 2048
	if frameSize > len(samples) {
		frameSize = len(samples)
		hopSize = frameSize / 2
		if hopSize < 1 { hopSize = 1 }
	}
	binFreqs := make([]float64, frameSize/2+1)
	for bin := range binFreqs {
		binFreqs[bin] = float64(bin) * float64(sampleRate) / float64(frameSize)
	}
	chromaSum := make([]float64, 12)
	frames := 0
	var tuningOffset float64
	tuningEstimated := false
	for pos := 0; pos+frameSize <= len(samples); pos += hopSize {
		frame := make([]float64, frameSize)
		for i := 0; i < frameSize; i++ { frame[i] = float64(samples[pos+i]) }
		window.Apply(frame, window.Hann)
		fftVals := fft.FFTReal(frame)
		mag := make([]float64, frameSize/2+1)
		for i := range mag {
			re := real(fftVals[i])
			im := imag(fftVals[i])
			mag[i] = math.Sqrt(re*re + im*im)
		}
		if !tuningEstimated {
			tuningOffset = EstimateTuning(mag, binFreqs)
			tuningEstimated = true
		}
		mag = whitenSpectrum(mag, 15)
		chromaFrame := make([]float64, 12)
		for bin, freq := range binFreqs {
			if freq < 65 || freq > 2100 { continue }
			adjustedFreq := freq * math.Pow(2, -tuningOffset/1200)
			midi := freqToMIDI(adjustedFreq)
			note := int(math.Round(midi)) % 12
			if note < 0 { note += 12 }
			chromaFrame[note] += mag[bin]
		}
		for i := range chromaSum { chromaSum[i] += chromaFrame[i] }
		frames++
	}
	if frames == 0 { return make([]float64, 12) }
	chromaAvg := make([]float64, 12)
	for i := range chromaAvg { chromaAvg[i] = chromaSum[i] / float64(frames) }
	return chromaAvg
}

type KeyGuess struct {
	Key   key.Key
	Score float64
}

type KeyCandidates struct {
	Best       KeyGuess
	Candidates []KeyGuess
}

func shiftProfile(profile []float64, shift int) []float64 {
	shifted := make([]float64, 12)
	for i := 0; i < 12; i++ {
		shifted[i] = profile[(i-shift+12)%12]
	}
	return shifted
}

func correlate(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 { return 0 }
	var meanA, meanB float64
	for i := range a { meanA += a[i]; meanB += b[i] }
	meanA /= float64(len(a))
	meanB /= float64(len(b))
	var cov, varA, varB float64
	for i := range a {
		da := a[i] - meanA
		db := b[i] - meanB
		cov += da * db
		varA += da * da
		varB += db * db
	}
	if varA == 0 || varB == 0 { return 0 }
	return cov / math.Sqrt(varA*varB)
}

func EstimateKeyFromChroma(chroma []float64) KeyCandidates {
	var guesses []KeyGuess
	for root := 0; root < 12; root++ {
		noteName := noteNamesForKey[root]
		guesses = append(guesses, KeyGuess{
			Key:   key.Of(noteName + " major"),
			Score: correlate(chroma, shiftProfile(majorProfile, root)),
		})
		guesses = append(guesses, KeyGuess{
			Key:   key.Of(noteName + " minor"),
			Score: correlate(chroma, shiftProfile(minorProfile, root)),
		})
	}
	sort.Slice(guesses, func(i, j int) bool { return guesses[i].Score > guesses[j].Score })
	topN := 3
	if len(guesses) < topN { topN = len(guesses) }
	return KeyCandidates{Best: guesses[0], Candidates: guesses[:topN]}
}

func DetectKey(samples []float32, sampleRate int, maxSeconds float64) (key.Key, []KeyGuess) {
	chroma := ComputeChroma(samples, sampleRate, maxSeconds)
	candidates := EstimateKeyFromChroma(chroma)
	return candidates.Best.Key, candidates.Candidates
}
