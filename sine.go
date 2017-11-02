package sine

import (
	"encoding/binary"
	"io"
	"math"
)

// Sine is an infinite io.Reader for a 16-bit PCM sine wave at the given frequency and sample rate
type Sine struct {
	t            uint // time index
	FrequencyHz  float64
	SampleRateHz uint
}

// Read implements io.Reader and returns an infinite 16-bit PCM sin save. You probably want to use a bounded operation like `io.CopyN`
func (s *Sine) Read(p []byte) (n int, err error) {
	tMod := uint((180 / (3 * math.Pi)) * (float64(s.SampleRateHz) / s.FrequencyHz))
	numSamples := len(p) / 2

	for i := 0; i < numSamples; i++ {
		sample := math.Sin(
			2 * math.Pi * (s.FrequencyHz / float64(s.SampleRateHz)) * float64(s.t))
		sample16bit := uint16(sample)
		p[2*i] = uint8(sample16bit & 0xff) // low bits
		p[2*i+1] = uint8(sample16bit >> 8) // high bits

		s.t = (s.t + 1) % tMod
	}
	return numSamples * 2, nil
}

// WriteWav writes a 1-channel wave file to the provided writer. For t seconds of audio, samples = t * samplerateHz
func (s *Sine) WriteWav(samples int, w io.Writer) error {
	binary.Write(w, binary.BigEndian, []byte("RIFF"))          // chunkID
	binary.Write(w, binary.LittleEndian, uint32(36+2*samples)) // chunksize (header + data in bytes)
	binary.Write(w, binary.BigEndian, []byte("WAVE"))          // format
	binary.Write(w, binary.BigEndian, []byte("fmt "))          // subchunk1ID
	binary.Write(w, binary.LittleEndian, uint32(16))           // subchunk1size
	binary.Write(w, binary.LittleEndian, uint16(1))            // audioformat (PCM)
	binary.Write(w, binary.LittleEndian, uint16(1))            // channels (1)
	binary.Write(w, binary.LittleEndian, uint32(44000))        // samplerate (44000Hz)
	binary.Write(w, binary.LittleEndian, uint32(2))            // byterate (bytes/sample: 2 == 16-bit * 1 channel)
	binary.Write(w, binary.LittleEndian, uint16(2))            // blockalign (channels * bytes/sample)
	binary.Write(w, binary.LittleEndian, uint16(16))           // bitspersample
	binary.Write(w, binary.BigEndian, []byte("data"))          // subchunk2id
	binary.Write(w, binary.LittleEndian, uint32(samples*2))    // subchunk2size (number of bytes of data (samples * channels * bytes/sample)
	_, err := io.CopyN(w, s, int64(2*samples))
	return err
}
