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

// audioPlayerF32 manages the playback of audio files using 32-bit float format
// This is the new recommended approach for Ebitengine v2.8+
type audioPlayerF32 struct {
	audioContext *audio.Context
	musicPlayers map[int]*audio.Player
	musicData    map[int][]byte
	mutex        sync.Mutex
}

// Global audio player instance for 32-bit float audio
var audioPlayerF32Instance *audioPlayerF32
var audioPlayerF32Once sync.Once

// getAudioPlayerF32 returns the singleton AudioPlayerF32 instance
func getAudioPlayerF32() *audioPlayerF32 {
	audioPlayerF32Once.Do(func() {
		audioContext := audio.NewContext(sampleRate)
		audioPlayerF32Instance = &audioPlayerF32{
			audioContext: audioContext,
			musicPlayers: make(map[int]*audio.Player),
			musicData:    make(map[int][]byte),
			mutex:        sync.Mutex{},
		}
		// Load all audio files at initialization
		audioPlayerF32Instance.loadAudioFiles()
	})
	return audioPlayerF32Instance
}

// loadAudioFiles loads all music*.wav files from the embedded resources
func (ap *audioPlayerF32) loadAudioFiles() {
	// Skip if no custom resources are registered
	if customResources == nil {
		log.Println("No custom resources registered, skipping audio file loading")
		return
	}

	// Walk through the embedded filesystem to find audio files
	walkErr := fs.WalkDir(customResources.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if the file is a music*.wav file
		if strings.HasPrefix(filepath.Base(path), "music") && strings.HasSuffix(path, ".wav") {
			filename := filepath.Base(path)
			var audioNumber int

			// Extract the number from the filename (e.g., "music1.wav" -> 1)
			_, err := fmt.Sscanf(filename, "music%d.wav", &audioNumber)
			if err != nil {
				log.Printf("Warning: Could not parse audio number from %s: %v", filename, err)
				return nil
			}

			// Read the audio file
			data, err := fs.ReadFile(customResources.FS, path)
			if err != nil {
				log.Printf("Warning: Could not read audio file %s: %v", path, err)
				return nil
			}

			// Store the audio data
			ap.musicData[audioNumber] = data
			log.Printf("Loaded 32-bit float audio file: %s (ID: %d)", path, audioNumber)
		}

		return nil
	})

	if walkErr != nil {
		log.Printf("Error walking through embedded filesystem: %v", walkErr)
	}
	log.Printf("Loaded %d 32-bit float audio files", len(ap.musicData))
}

// MusicF32 plays the audio file with the given ID using 32-bit float format.
// This is the new recommended approach for Ebitengine v2.8+ as it provides
// better performance and easier audio processing.
// If n is -1, it stops all currently playing audio.
// If n is a valid audio ID, it plays that audio file.
// If exclusive is true, it stops all other audio files before playing.
func MusicF32(n int, exclusive ...bool) {
	if n == -1 {
		// Special case: stop all music
		StopMusicF32(-1)
		return
	}

	// Default exclusive to false
	shouldBeExclusive := len(exclusive) > 0 && exclusive[0]

	ap := getAudioPlayerF32()
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	// Stop other audio if requested
	if shouldBeExclusive {
		for _, player := range ap.musicPlayers {
			if player != nil {
				player.Pause()
				if err := player.Rewind(); err != nil {
					log.Printf("Error rewinding player: %v", err)
				}
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
		if err := player.Rewind(); err != nil {
			log.Printf("Error rewinding player: %v", err)
		}
		player.Play()
		return
	}

	// Create a new player for this audio using 32-bit float format
	reader := bytes.NewReader(audioData)
	wavReader, err := wav.DecodeF32(reader)
	if err != nil {
		log.Printf("Error decoding WAV file to 32-bit float (ID: %d): %v", n, err)
		return
	}

	player, err = ap.audioContext.NewPlayerF32(wavReader)
	if err != nil {
		log.Printf("Error creating 32-bit float audio player (ID: %d): %v", n, err)
		return
	}

	// Store the player and play
	ap.musicPlayers[n] = player
	player.Play()
}

// StopMusicF32 stops the audio file with the given ID using 32-bit float format
// If id is -1, it stops all audio files
func StopMusicF32(id int) {
	ap := getAudioPlayerF32()
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if id == -1 {
		// Stop all audio
		for _, player := range ap.musicPlayers {
			if player != nil {
				player.Pause()
				if err := player.Rewind(); err != nil {
					log.Printf("Error rewinding player: %v", err)
				}
			}
		}
		return
	}

	// Stop specific audio
	player, exists := ap.musicPlayers[id]
	if exists && player != nil {
		player.Pause()
		if err := player.Rewind(); err != nil {
			log.Printf("Error rewinding player: %v", err)
		}
	}
}

// IsAudioF32Available returns true if 32-bit float audio is supported
// This can be used to check if the new audio features are available
func IsAudioF32Available() bool {
	// In Ebitengine v2.8+, 32-bit float audio is always available
	return true
}
