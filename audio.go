package pigo8

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	// SampleRate is the audio sample rate used for all audio playback
	SampleRate = 44100
	// AudioChannels is the number of audio channels (stereo)
	AudioChannels = 2
)

// AudioPlayer manages the playback of audio files
type AudioPlayer struct {
	audioContext *audio.Context
	musicPlayers map[int]*audio.Player
	musicData    map[int][]byte
	mutex        sync.Mutex
}

// Global audio player instance
var audioPlayerInstance *AudioPlayer
var audioPlayerOnce sync.Once

// GetAudioPlayer returns the singleton AudioPlayer instance
func GetAudioPlayer() *AudioPlayer {
	audioPlayerOnce.Do(func() {
		audioContext := audio.NewContext(SampleRate)
		audioPlayerInstance = &AudioPlayer{
			audioContext: audioContext,
			musicPlayers: make(map[int]*audio.Player),
			musicData:    make(map[int][]byte),
			mutex:        sync.Mutex{},
		}
		// Load all audio files at initialization
		audioPlayerInstance.loadAudioFiles()
	})
	return audioPlayerInstance
}

// loadAudioFiles loads all audio*.wav files from the embedded resources
func (ap *AudioPlayer) loadAudioFiles() {
	// Skip if no custom resources are registered
	if CustomResources == nil {
		log.Println("No custom resources registered, skipping audio file loading")
		return
	}

	// Walk through the embedded filesystem to find audio files
	fs.WalkDir(CustomResources.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if the file is an audio*.wav file
		if strings.HasPrefix(filepath.Base(path), "audio") && strings.HasSuffix(path, ".wav") {
			// Extract the number from the filename (e.g., "audio1.wav" -> 1)
			filename := filepath.Base(path)
			var audioNumber int
			_, err := fmt.Sscanf(filename, "audio%d.wav", &audioNumber)
			if err != nil {
				log.Printf("Warning: Could not parse audio number from %s: %v", filename, err)
				return nil
			}

			// Read the audio file
			data, err := fs.ReadFile(CustomResources.FS, path)
			if err != nil {
				log.Printf("Warning: Could not read audio file %s: %v", path, err)
				return nil
			}

			// Store the audio data
			ap.musicData[audioNumber] = data
			log.Printf("Loaded audio file: %s (ID: %d)", path, audioNumber)
		}

		return nil
	})

	log.Printf("Loaded %d audio files", len(ap.musicData))
}

// music plays the audio file with the given ID
// If n is -1, it stops all currently playing audio
// If n is a valid audio ID, it plays that audio file
// If exclusive is true, it stops all other audio files before playing
func Music(n int, exclusive ...bool) {
	if n == -1 {
		// Special case: stop all music
		StopMusic(-1)
		return
	}

	// Default exclusive to false
	shouldBeExclusive := false
	if len(exclusive) > 0 && exclusive[0] {
		shouldBeExclusive = true
	}

	ap := GetAudioPlayer()
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	// Stop other audio if requested
	if shouldBeExclusive {
		for _, player := range ap.musicPlayers {
			if player != nil {
				player.Pause()
				player.Rewind()
			}
		}
	}

	// Check if the audio file exists
	audioData, exists := ap.musicData[n]
	if !exists {
		log.Printf("Warning: Audio file with ID %d not found", n)
		return
	}

	// Check if this audio is already playing
	player, exists := ap.musicPlayers[n]
	if exists && player != nil {
		if player.IsPlaying() {
			// Already playing, do nothing
			return
		}
		// Player exists but is not playing, rewind and play
		player.Rewind()
		player.Play()
		return
	}

	// Create a new player for this audio
	reader := bytes.NewReader(audioData)
	wavReader, err := wav.Decode(ap.audioContext, reader)
	if err != nil {
		log.Printf("Error decoding WAV file (ID: %d): %v", n, err)
		return
	}

	player, err = ap.audioContext.NewPlayer(wavReader)
	if err != nil {
		log.Printf("Error creating audio player (ID: %d): %v", n, err)
		return
	}

	// Store the player and play
	ap.musicPlayers[n] = player
	player.Play()
}

// StopMusic stops the audio file with the given ID
// If id is -1, it stops all audio files
func StopMusic(id int) {
	ap := GetAudioPlayer()
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if id == -1 {
		// Stop all audio
		for _, player := range ap.musicPlayers {
			if player != nil {
				player.Pause()
				player.Rewind()
			}
		}
		return
	}

	// Stop specific audio
	player, exists := ap.musicPlayers[id]
	if exists && player != nil {
		player.Pause()
		player.Rewind()
	}
}
