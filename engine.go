package pigo8

import (
	"log"
	"os"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Settings ---

// Settings defines configurable parameters for the PIGO8 console.
type Settings struct {
	ScaleFactor       int    // Integer scaling factor for the window (Default: 4).
	WindowTitle       string // Title displayed on the window bar (Default: "PIGO-8 Game").
	TargetFPS         int    // Target ticks per second (Default: 30).
	ScreenWidth       int    // Custom screen width (Default: 128 for PICO-8 compatibility).
	ScreenHeight      int    // Custom screen height (Default: 128 for PICO-8 compatibility).
}

// NewSettings creates a new Settings object with default values.
func NewSettings() *Settings {
	return &Settings{
		ScaleFactor:  4,
		WindowTitle:  "PIGO-8 Game",
		TargetFPS:    30,
		ScreenWidth:  128, // Default PICO-8 width
		ScreenHeight: 128, // Default PICO-8 height
	}
}

// --- Configuration (Internal Constants) ---

const (
	// DefaultWidth is the default PICO-8 screen width.
	DefaultWidth = 128
	// DefaultHeight is the default PICO-8 screen height.
	DefaultHeight = 128
)

// These variables hold the current logical screen dimensions.
// They can be modified through Settings when calling PlayGameWith.
var (
	// ScreenWidth is the current screen width.
	ScreenWidth = DefaultWidth
	// ScreenHeight is the current screen height.
	ScreenHeight = DefaultHeight
)

// --- Internal State for Drawing ---

// These hold the current context for the user's draw callback.
// They are set by the engine's Draw method each frame.
var (
	currentScreen    *ebiten.Image // Internal: Current screen image
	currentSprites   []SpriteInfo  // Internal: Loaded sprites
	currentDrawColor int           // Internal: Current draw color (0-15)
	elapsedTime      float64       // Internal: Time elapsed since game start (in seconds)
	timeIncrement    float64       // Internal: Amount to increment time each update
)

// --- Cartridge Definition and Loading ---

// Cartridge defines the functions a user game must implement.
// This is the user's "game cartridge" code.
type Cartridge interface {
	Init()   // Called once at the start.
	Update() // Called every frame for logic.
	Draw()   // Called every frame for drawing.
}

// Internal default implementation.
type emptyCartridge struct{}

func (d *emptyCartridge) Init()   {}
func (d *emptyCartridge) Update() {}
func (d *emptyCartridge) Draw()   {}

// loadedCartridge holds the registered implementation provided by the user.
var loadedCartridge Cartridge = &emptyCartridge{}

// InsertGame sets the user's game implementation (the cartridge).
// Pass an instance of a struct that implements the Cartridge interface.
func InsertGame(cartridge Cartridge) {
	if cartridge == nil {
		loadedCartridge = &emptyCartridge{} // Reset to default if nil
	} else {
		loadedCartridge = cartridge
	}
}

// CurrentCartridge returns the currently loaded cartridge.
// This is useful for accessing the game state from network callbacks.
func CurrentCartridge() Cartridge {
	return loadedCartridge
}

// --- Internal Ebiten Game Implementation ---

// game is the internal struct that satisfies ebiten.Game interface.
type game struct {
	initialized     bool
	firstFrameDrawn bool // Track if the first frame has been drawn
}

// Layout implements ebiten.Game.
func (g *game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

// Update implements ebiten.Game.
func (g *game) Update() error {
	if !g.initialized {
		log.Println("Cartridge Initializing...")
		// Log initial memory usage
		logInitialMemory()
		loadedCartridge.Init()
		g.initialized = true
		// Don't call Update on the first frame, wait for Draw to be called first
		return nil
	}

	// Only call Update after the first frame has been drawn
	if g.firstFrameDrawn {
		UpdateConnectedGamepads()
		UpdateMouseState()
		loadedCartridge.Update()

		// Update elapsed time
		elapsedTime += timeIncrement
	}
	return nil
}

// Draw implements ebiten.Game.
func (g *game) Draw(screen *ebiten.Image) {
	// Set the current screen for drawing
	currentScreen = screen

	// Clear the screen
	screen.Clear()

	// Call the user's Draw function
	loadedCartridge.Draw()

	// Mark that the first frame has been drawn
	if !g.firstFrameDrawn {
		g.firstFrameDrawn = true
	}
}

// --- Helper for User Code ---

// CurrentSprites returns the currently loaded sprite data slice.
// Useful if user code needs to check len(sprites) or access flags.
func CurrentSprites() []SpriteInfo {
	return currentSprites
}

// CurrentScreen returns the current Ebiten screen image for direct drawing.
// This allows advanced users to use Ebiten's drawing functions directly.
func CurrentScreen() *ebiten.Image {
	return currentScreen
}

// --- Time Functions ---

// Time returns the number of seconds (as a decimal) since the game started.
// This is calculated by counting the number of times the Update method is called.
// Multiple calls to Time() in the same frame will return the same value.
func Time() float64 {
	return elapsedTime
}

// T is an alias for Time() for PICO-8 compatibility.
func T() float64 {
	return Time()
}

// --- Play Functions ---

// logInitialMemory logs the initial memory usage of the PIGO-8 console
func logInitialMemory() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("PIGO-8 Memory: %.2f MB", float64(m.Alloc)/1024/1024)
}

// PlayGameWith initializes Ebitengine, loads resources, and starts the main game loop.
// It uses the provided Settings and runs the Cartridge previously loaded via Insert.
func PlayGameWith(settings *Settings) {
	// Use default settings if nil is passed
	cfg := settings
	if cfg == nil {
		log.Println("Warning: pigo8.PlayGameWith called with nil Settings, using defaults.")
		cfg = NewSettings()
	}

	// Check for network configuration from command line arguments
	networkConfig := ParseNetworkArgs()
	
	// Set game name from window title if not specified
	if networkConfig.GameName == "PIGO8 Game" {
		networkConfig.GameName = cfg.WindowTitle
	}
	
	// Initialize networking if enabled (determined by ParseNetworkArgs)
	if IsNetworkInitialized() || networkConfig != nil {
		if err := InitNetwork(networkConfig); err != nil {
			log.Printf("Warning: Failed to initialize network: %v", err)
		}
		defer ShutdownNetwork()
	}

	// Reset time tracking variables
	elapsedTime = 0.0

	// Update logical screen dimensions if custom values are provided
	if cfg.ScreenWidth > 0 {
		ScreenWidth = cfg.ScreenWidth
	} else {
		ScreenWidth = DefaultWidth
		cfg.ScreenWidth = DefaultWidth
	}

	if cfg.ScreenHeight > 0 {
		ScreenHeight = cfg.ScreenHeight
	} else {
		ScreenHeight = DefaultHeight
		cfg.ScreenHeight = DefaultHeight
	}

	// Try to load custom palette from palette.hex if it exists
	loadPaletteFromHexFile()

	// Configure Ebitengine window using Settings object
	ebiten.SetWindowTitle(cfg.WindowTitle)
	winWidth := ScreenWidth * cfg.ScaleFactor
	winHeight := ScreenHeight * cfg.ScaleFactor
	if winWidth <= 0 || winHeight <= 0 {
		log.Printf("Warning: Calculated window size (%dx%d based on ScaleFactor %d) is non-positive. Using default %dx%d.", winWidth, winHeight, cfg.ScaleFactor, DefaultWidth, DefaultHeight)
		winWidth, winHeight = DefaultWidth, DefaultHeight
	}
	ebiten.SetWindowSize(winWidth, winHeight)
	ebiten.SetTPS(cfg.TargetFPS)

	// Calculate time increment based on target FPS
	timeIncrement = 1.0 / float64(cfg.TargetFPS)

	internalGame := &game{
		initialized: false,
	}

	log.Println("Booting PIGO8 console...")
	err := ebiten.RunGame(internalGame)
	if err != nil {
		log.Panicf("pico8.PlayGameWith: Ebitengine loop failed: %v", err)
	}
	log.Println("PIGO8 console shutdown.")
	os.Exit(0)
}

// Play initializes and runs the PIGO8 console with default settings.
// It runs the Cartridge previously loaded via Insert.
// This is a convenience wrapper around PlayGameWith(NewSettings()).
func Play() {
	PlayGameWith(NewSettings())
}
