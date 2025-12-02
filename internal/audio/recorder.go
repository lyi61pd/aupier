package audio

import (
"encoding/binary"
"fmt"
"os"
"sync"
"time"

"github.com/gordonklaus/portaudio"
)

type Recorder struct {
	stream     *portaudio.Stream
	buffer     []int16
	recording  bool
	mu         sync.Mutex
	sampleRate int
	channels   int
}

func NewRecorder(sampleRate, channels int) (*Recorder, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize portaudio: %w", err)
	}

	return &Recorder{
		sampleRate: sampleRate,
		channels:   channels,
		buffer:     make([]int16, 0),
	}, nil
}

func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.recording {
		return fmt.Errorf("already recording")
	}

	r.buffer = make([]int16, 0)

	stream, err := portaudio.OpenDefaultStream(
r.channels,
0,
float64(r.sampleRate),
0,
r.recordCallback,
)
	if err != nil {
		return fmt.Errorf("failed to open audio stream: %w", err)
	}

	if err := stream.Start(); err != nil {
		stream.Close()
		return fmt.Errorf("failed to start audio stream: %w", err)
	}

	r.stream = stream
	r.recording = true

	return nil
}

func (r *Recorder) recordCallback(in []int16) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buffer = append(r.buffer, in...)
}

func (r *Recorder) Stop(filename string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.recording {
		return fmt.Errorf("not recording")
	}

	if r.stream != nil {
		if err := r.stream.Stop(); err != nil {
			return fmt.Errorf("failed to stop audio stream: %w", err)
		}
		if err := r.stream.Close(); err != nil {
			return fmt.Errorf("failed to close audio stream: %w", err)
		}
		r.stream = nil
	}

	r.recording = false

	if err := r.saveWAV(filename); err != nil {
		return fmt.Errorf("failed to save WAV file: %w", err)
	}

	return nil
}

func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

func (r *Recorder) saveWAV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bitsPerSample := 16
	byteRate := r.sampleRate * r.channels * bitsPerSample / 8
	blockAlign := r.channels * bitsPerSample / 8
	dataSize := len(r.buffer) * 2
	fileSize := 36 + dataSize

	file.Write([]byte("RIFF"))
	binary.Write(file, binary.LittleEndian, int32(fileSize))
	file.Write([]byte("WAVE"))

	file.Write([]byte("fmt "))
	binary.Write(file, binary.LittleEndian, int32(16))
	binary.Write(file, binary.LittleEndian, int16(1))
	binary.Write(file, binary.LittleEndian, int16(r.channels))
	binary.Write(file, binary.LittleEndian, int32(r.sampleRate))
	binary.Write(file, binary.LittleEndian, int32(byteRate))
	binary.Write(file, binary.LittleEndian, int16(blockAlign))
	binary.Write(file, binary.LittleEndian, int16(bitsPerSample))

	file.Write([]byte("data"))
	binary.Write(file, binary.LittleEndian, int32(dataSize))
	binary.Write(file, binary.LittleEndian, r.buffer)

	return nil
}

func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream != nil {
		if r.recording {
			r.stream.Stop()
		}
		r.stream.Close()
		r.stream = nil
	}

	portaudio.Terminate()
	return nil
}

func GenerateFilename(outputDir string) string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("%s/clip_%s.wav", outputDir, timestamp)
}
