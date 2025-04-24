package pigo8

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Settings ---

// Settings defines configurable parameters for the PIGO8 console.
type Settings struct {
	ScaleFactor int    // Integer scaling factor for the window (Default: 5).
	WindowTitle string // Title displayed on the window bar (Default: "PIGO-8 Game").
	TargetFPS   int    // Target ticks per second (Default: 30).
}

// NewSettings creates a new Settings object with default values.
func NewSettings() *Settings {
	return &Settings{
		ScaleFactor: 4,
		WindowTitle: "PIGO-8 Game",
		TargetFPS:   30,
	}
}

// --- Configuration (Internal Constants) ---

const (
	// LogicalWidth is the fixed PICO-8 screen width.
	LogicalWidth = 128
	// LogicalHeight is the fixed PICO-8 screen height.
	LogicalHeight = 128
)

// --- Internal State for Drawing ---

// These hold the current context for the user's draw callback.
// They are set by the engine's Draw method each frame.
var (
	currentScreen  *ebiten.Image // Internal: Current screen image
	currentSprites []SpriteInfo  // Internal: Loaded sprites
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

// --- Internal Ebiten Game Implementation ---

// game is the internal struct that satisfies ebiten.Game interface.
type game struct {
	initialized bool
}

// Layout implements ebiten.Game.
func (g *game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return LogicalWidth, LogicalHeight
}

// Update implements ebiten.Game.
func (g *game) Update() error {
	if !g.initialized {
		log.Println("Cartridge Initializing...")
		loadedCartridge.Init()
		g.initialized = true
	}
	UpdateConnectedGamepads()
	loadedCartridge.Update()
	return nil
}

// Draw implements ebiten.Game.
func (g *game) Draw(screen *ebiten.Image) {
	currentScreen = screen
	loadedCartridge.Draw()
}

// --- Helper for User Code ---

// CurrentSprites returns the currently loaded sprite data slice.
// Useful if user code needs to check len(sprites) or access flags.
func CurrentSprites() []SpriteInfo {
	return currentSprites
}

// --- Play Functions ---

// PlayGameWith initializes Ebitengine, loads resources, and starts the main game loop.
// It uses the provided Settings and runs the Cartridge previously loaded via Insert.
func PlayGameWith(settings *Settings) {
	// Use default settings if nil is passed
	cfg := settings
	if cfg == nil {
		log.Println("Warning: pico8.PlayGameWith called with nil Settings, using defaults.")
		cfg = NewSettings()
	}

	// Configure Ebitengine window using Settings object
	ebiten.SetWindowTitle(cfg.WindowTitle)
	winWidth := LogicalWidth * cfg.ScaleFactor
	winHeight := LogicalHeight * cfg.ScaleFactor
	if winWidth <= 0 || winHeight <= 0 {
		log.Printf("Warning: Calculated window size (%dx%d based on ScaleFactor %d) is non-positive. Using default 128x128.", winWidth, winHeight, cfg.ScaleFactor)
		winWidth, winHeight = LogicalWidth, LogicalHeight
	}
	ebiten.SetWindowSize(winWidth, winHeight)
	ebiten.SetTPS(cfg.TargetFPS)

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
