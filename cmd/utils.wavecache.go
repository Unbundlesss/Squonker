package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"

	"github.com/go-audio/wav"
	"github.com/peterbourgon/diskv"
)

type waveState struct {
	WaveFilename string
	WaveHash     string

	LoopPointsValid bool    // whether the loop points were set
	LoopPointStart  float64 // 0..1 for where in sample we should begin looping
	LoopPointLength float64 // nonlinear madness, see comments later
	SampleCount     int64   // only valid if loopPointsValid

	SampleRate   int
	ChannelCount int // stereo or not
	BitDepth     int
}

type waveCache struct {
	storage *diskv.Diskv
}

func NewWaveStateCache() *waveCache {
	wsc := new(waveCache)

	cacheBase := path.Join(cOutputRoot, "_cache/wav.state")
	if _, err := os.Stat(cacheBase); os.IsNotExist(err) {
		err = os.MkdirAll(cacheBase, 0777)
		if err != nil {
			sqFatal(err)
		}
	}

	flatTransform := func(s string) []string { return []string{} }
	wsc.storage = diskv.New(diskv.Options{
		BasePath:     cacheBase,
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})

	return wsc
}

// look up or generate a WaveState structure for a given .wav file; returns true for
// second result if value came from the cache, false if we generated from scratch
func (wsc *waveCache) GetWavFileState(wavFilename string) (waveState, bool, error) {

	result := waveState{}

	// get a hash for the contents of the WAV to use as a cache key
	wavHash, err := fastFileHash(wavFilename)
	if err != nil {
		return result, false, err
	}

	// try and get the values from the cache, returning them if we got them
	if wsc.storage.Has(wavHash) {
		cachedValue, err := wsc.storage.Read(wavHash)
		if err != nil {
			return result, false, err
		}
		if len(cachedValue) > 0 {
			err = json.Unmarshal(cachedValue, &result)
			if err != nil {
				return result, true, err
			}
			return result, true, nil
		}
		return result, true, fmt.Errorf("cache value empty (H:%s...)", wavHash[:8])
	}

	// .. if not, we have to do it the hard way
	{
		wavFileFs, err := os.Open(wavFilename)
		if err != nil {
			return result, false, err
		}
		defer wavFileFs.Close()

		wavDecoder := wav.NewDecoder(wavFileFs)
		if wavDecoder == nil || !wavDecoder.IsValidFile() {
			return result, false, fmt.Errorf("WAV file is not valid / loadable")
		}
		wavDecoder.ReadInfo()

		// Von cracked Endlesss' mysterious loopLength code; spoilers, it isn't a loop length. it's something else.
		// also! note that loopLength is (presently) dependant on the *PLAYBACK SAMPLE RATE ON THE DEVICE* which is .. something out of our control
		//       .. so the current computation below is tuned to give the right length for a 44.1kHz playback sample rate
		wavDecoder.ReadMetadata()
		if wavDecoder.Metadata != nil && wavDecoder.Metadata.SamplerInfo != nil {
			if wavDecoder.Metadata.SamplerInfo.NumSampleLoops > 0 {
				loop := wavDecoder.Metadata.SamplerInfo.Loops[0]

				// read the full PCM buffer set (inefficiently) to find the number of samples so we can compute loop percentages
				// it's slow but there were problems doing it "properly" and this seems to be fairly bulletproof
				wavFileFs.Seek(0, 0)
				wavDecoder = wav.NewDecoder(wavFileFs)
				audioBuf, err := wavDecoder.FullPCMBuffer()
				if err != nil {
					return result, false, err
				}
				totalSamplesInPCM := audioBuf.NumFrames()

				// loopStart makes sense; it's 0..1 (0%..100%) of where to begin the loop
				loopStart := float64(loop.Start) / float64(totalSamplesInPCM)

				// loopLength is some kind of cruel prank
				loopSamples := float64(loop.End) - float64(loop.Start)
				loopLength := math.Pow(((loopSamples/441.0)-1.0)/500.0, (1.0 / 3.0))

				result.LoopPointsValid = true
				result.LoopPointStart = loopStart
				result.LoopPointLength = loopLength
				result.SampleCount = int64(totalSamplesInPCM)
			}
		}

		result.ChannelCount = int(wavDecoder.NumChans)
		result.SampleRate = int(wavDecoder.SampleRate)
		result.BitDepth = int(wavDecoder.BitDepth)
	}

	result.WaveFilename = wavFilename
	result.WaveHash = wavHash

	// dont forget to write the results into the cache
	cachedValue, err := json.Marshal(result)
	if err != nil {
		return result, false, err
	}
	err = wsc.storage.Write(wavHash, cachedValue)
	if err != nil {
		return result, false, err
	}

	return result, false, nil
}
