package audio

import (
"encoding/binary"
"fmt"
"io"
"os"
"sync"

"github.com/gordonklaus/portaudio"
)

type Player struct {
	stream   *portaudio.Stream
	buffer   []int16
	position int
	playing  bool
	looping  bool
	mu       sync.Mutex
	channels int
	sampleRate int
	volume   float32
}

func NewPlayer() (*Player, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize portaudio: %w", err)
	}

	return &Player{
		volume: 1.0,
	}, nil
}

func (p *Player) LoadWAV(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var riff [4]byte
	var size int32
	var wave [4]byte

	binary.Read(file, binary.LittleEndian, &riff)
	binary.Read(file, binary.LittleEndian, &size)
	binary.Read(file, binary.LittleEndian, &wave)

	if string(riff[:]) != "RIFF" || string(wave[:]) != "WAVE" {
		return fmt.Errorf("invalid WAV file")
	}

	for {
		var chunkID [4]byte
		var chunkSize int32

		if err := binary.Read(file, binary.LittleEndian, &chunkID); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := binary.Read(file, binary.LittleEndian, &chunkSize); err != nil {
			return err
		}

		switch string(chunkID[:]) {
		case "fmt ":
			var format int16
			var channels int16
			var sampleRate int32

			binary.Read(file, binary.LittleEndian, &format)
			binary.Read(file, binary.LittleEndian, &channels)
			binary.Read(file, binary.LittleEndian, &sampleRate)

			p.channels = int(channels)
			p.sampleRate = int(sampleRate)

			file.Seek(int64(chunkSize-6), io.SeekCurrent)

		case "data":
			numSamples := int(chunkSize) / 2
			p.buffer = make([]int16, numSamples)
			binary.Read(file, binary.LittleEndian, p.buffer)

		default:
			file.Seek(int64(chunkSize), io.SeekCurrent)
		}
	}

	p.position = 0
	return nil
}

func (p *Player) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.playing {
		return nil
	}

	if len(p.buffer) == 0 {
		return fmt.Errorf("no audio data loaded")
	}

	stream, err := portaudio.OpenDefaultStream(
0,
p.channels,
float64(p.sampleRate),
0,
p.playCallback,
)
	if err != nil {
		return fmt.Errorf("failed to open audio stream: %w", err)
	}

	if err := stream.Start(); err != nil {
		stream.Close()
		return fmt.Errorf("failed to start audio stream: %w", err)
	}

	p.stream = stream
	p.playing = true

	return nil
}

func (p *Player) playCallback(out []int16) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range out {
		if p.position >= len(p.buffer) {
			if p.looping {
				p.position = 0
			} else {
				out[i] = 0
				continue
			}
		}

		sample := float32(p.buffer[p.position]) * p.volume
		if sample > 32767 {
			sample = 32767
		} else if sample < -32768 {
			sample = -32768
		}
		out[i] = int16(sample)
		p.position++
	}

	if p.position >= len(p.buffer) && !p.looping {
		if p.stream != nil {
			p.stream.Stop()
			p.stream.Close()
			p.stream = nil
			p.playing = false
		}
	}
}

func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stream != nil {
		if err := p.stream.Stop(); err != nil {
			return err
		}
		if err := p.stream.Close(); err != nil {
			return err
		}
		p.stream = nil
	}

	p.playing = false
	p.position = 0

	return nil
}

func (p *Player) SetLoop(loop bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.looping = loop
}

func (p *Player) IsLooping() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.looping
}

func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}

func (p *Player) SetVolume(volume float32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if volume < 0 {
		volume = 0
	} else if volume > 2 {
		volume = 2
	}
	p.volume = volume
}

func (p *Player) GetVolume() float32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

func (p *Player) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stream != nil {
		if p.playing {
			p.stream.Stop()
		}
		p.stream.Close()
		p.stream = nil
	}

	portaudio.Terminate()
	return nil
}
