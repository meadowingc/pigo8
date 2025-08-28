// Package main basic super mario bros example using the new camera system
package main

import (
	"fmt"
	"math"

	"github.com/drpaneas/pigo8"
)

// =============================================================================
// GAME CONSTANTS - These control how the game behaves
// =============================================================================

// Physics constants - these control how Mario moves and feels
const (
	gravitySlow = 1.0 // Gravity when holding jump (slower fall)
	gravityFast = 5.0 // Normal gravity (faster fall)

	// Movement constants
	walkSpeed           = 1.5   // Normal walking speed
	runSpeed            = 2.5   // Running speed
	acceleration        = 10.0  // Small Mario acceleration
	skidAcceleration    = 32.0  // Acceleration when skidding
	normalAcceleration  = 14.0  // Normal acceleration
	physicsDivisor      = 256.0 // Divisor for physics calculations
	maxMovementPerFrame = 4.0   // Maximum pixels to move per frame
	coyoteTimeFrames    = 6.0   // Frames of grace period for jumping

	// Animation constants
	animationFrameThreshold = 12.0 // How many frames to wait before changing sprite
	spriteFrameWidth        = 16.0 // Width of each sprite frame in spritesheet
	skiddingThreshold       = 0.5  // Speed threshold for skidding animation
	walkingThreshold        = 0.1  // Speed threshold for walking animation
	animationSpeedDivisor   = 2.5  // Divisor for animation speed calculation
	animationRateMultiplier = 4.0  // Multiplier for animation rate

	// Starting position
	playerStartX = 32
	playerStartY = 59

	// Sprite sheet coordinates - where each animation frame is located
	playerIdleSX          = 8  // Mario standing still
	playerWalkAnimStartSX = 24 // Start of walking animation
	playerWalkAnimEndSX   = 56 // End of walking animation
	playerSkidSX          = 72 // Mario skidding to a stop
	playerJumpFallSX      = 88 // Mario jumping or falling
	playerSpriteSY        = 24 // Y position of Mario sprites in the sheet
)

// Jump force array - how high Mario jumps based on his speed
// Faster Mario = higher jump (exact values from SMB3 reference)
var jumpForce = []float64{-3.5, -3.625, -3.75, -4.0}

// =============================================================================
// GAME DATA STRUCTURES
// =============================================================================

// entity represents a game entity (like Mario) with animation capabilities
type entity struct {
	// Position and movement
	position pigo8.Vector2D // Current position
	velocity pigo8.Vector2D // Velocity (speed and direction)
	dir      float64        // Facing direction (1 = right, -1 = left)

	// Sprite and animation
	spriteIndex       float64 // Current sprite frame number
	spriteSourceW     float64 // Width of sprite in spritesheet
	spriteSourceH     float64 // Height of sprite in spritesheet
	spriteDestW       float64 // Width to draw on screen
	spriteDestH       float64 // Height to draw on screen
	spriteSourceY     float64 // Y position in spritesheet
	first, last       float64 // Animation frame range
	animationProgress float64 // How far through current animation
	currentAnimRate   float64 // Speed of animation

	// State flags
	onGround bool // Is Mario touching the ground?
	inAir    bool // Is Mario in the air?
	isBig    bool // Is Mario big (vs small)?
}

// Global game variables
var (
	player     entity  // The player character (Mario)
	coyoteTime float64 // Coyote time counter (frames Mario can still jump after leaving ground)
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// newEntity creates a new animated entity with default values
func newEntity(sprite, x, y, maxAnimRate, firstFrame, lastFrame, initialDir float64) entity {
	return entity{
		// Position and movement
		spriteIndex: sprite,
		position:    pigo8.NewVector2D(x, y),
		velocity:    pigo8.ZeroVector(),
		dir:         initialDir,

		// Sprite dimensions (16x16 pixels for small Mario)
		spriteSourceW: 16, spriteSourceH: 16,
		spriteDestW: 16, spriteDestH: 16,
		spriteSourceY: playerSpriteSY,

		// State
		onGround: false, inAir: true, isBig: false,

		// Animation
		animationProgress: 0, currentAnimRate: maxAnimRate,
		first: firstFrame, last: lastFrame,
	}
}

// animate updates the sprite index based on timing to create smooth animations
func (e *entity) animate() {
	// Don't animate if animation rate is 0 (for static sprites like idle)
	if e.currentAnimRate <= 0 {
		return
	}

	// Add to animation progress
	e.animationProgress += e.currentAnimRate

	// When we reach the threshold, move to next frame
	if e.animationProgress >= animationFrameThreshold {
		e.animationProgress -= animationFrameThreshold
		e.spriteIndex += spriteFrameWidth // Move to next frame in spritesheet

		// Loop back to first frame when we reach the end
		if e.spriteIndex > e.last {
			e.spriteIndex = e.first
		}
	}
}

// draw renders the entity to the screen
func (e *entity) draw() {
	// Flip sprite horizontally if facing left
	flip := e.dir < 0

	// Get position as integers for drawing
	x, y := e.position.ToInt()

	// Draw the sprite at current position
	pigo8.Sspr(int(e.spriteIndex), int(e.spriteSourceY), int(e.spriteSourceW), int(e.spriteSourceH),
		x, y, int(e.spriteDestW), int(e.spriteDestH), flip, false)
}

// =============================================================================
// INPUT HANDLING FUNCTIONS
// =============================================================================

// getPlayerInput reads all player input and returns a struct with the current state
type playerInput struct {
	isMovingLeft  bool
	isMovingRight bool
	jumpPressed   bool
	isRunning     bool
}

func getPlayerInput() playerInput {
	return playerInput{
		isMovingLeft:  pigo8.Btn(pigo8.LEFT),  // Left arrow or A
		isMovingRight: pigo8.Btn(pigo8.RIGHT), // Right arrow or D
		jumpPressed:   pigo8.Btnp(pigo8.X),    // X key (pressed this frame)
		isRunning:     pigo8.Btn(pigo8.O),     // O key (hold to run)
	}
}

// =============================================================================
// MOVEMENT PHYSICS FUNCTIONS
// =============================================================================

// handleHorizontalMovement applies physics to Mario's left/right movement
func handleHorizontalMovement(input playerInput) {
	// Set movement speed based on whether player is running
	topSpeed := walkSpeed // Normal walking speed
	if input.isRunning {
		topSpeed = runSpeed // Running speed
	}

	// Set acceleration/friction for small Mario
	accelFric := acceleration // Small Mario acceleration

	// Determine which direction player wants to move
	hitDir := 0.0 // 0 = no input, 1 = right, -1 = left
	if input.isMovingRight {
		hitDir = 1.0
	} else if input.isMovingLeft {
		hitDir = -1.0
	}

	// Apply movement physics
	if hitDir == 0 {
		// No input - apply friction to slow down
		if !player.inAir {
			if player.velocity.X < 0 {
				player.velocity.X += accelFric / physicsDivisor // Slow down when moving left
				if player.velocity.X > 0 {
					player.velocity.X = 0 // Stop if we overshoot
				}
			} else if player.velocity.X > 0 {
				player.velocity.X -= accelFric / physicsDivisor // Slow down when moving right
				if player.velocity.X < 0 {
					player.velocity.X = 0 // Stop if we overshoot
				}
			}
		}
	} else {
		// Input detected - apply acceleration or skidding
		absVx := math.Abs(player.velocity.X)

		// Use switch statement to satisfy linter while maintaining functionality
		switch {
		case (player.velocity.X > 0 && hitDir < 0) || (player.velocity.X < 0 && hitDir > 0):
			// Skidding - player changed direction while moving
			player.velocity.X += hitDir * skidAcceleration / physicsDivisor
		case absVx < topSpeed:
			// Accelerate toward top speed
			player.velocity.X += hitDir * normalAcceleration / physicsDivisor
		case absVx > topSpeed && !player.inAir:
			// Slow down if over top speed (only when on ground)
			player.velocity.X -= hitDir * accelFric / physicsDivisor
		}
	}
}

// handleVerticalMovement manages jumping and gravity
func handleVerticalMovement(input playerInput) {
	// Jump when button is pressed and Mario has coyote time or is on ground
	if input.jumpPressed && (player.onGround || coyoteTime > 0) {
		// Jump height depends on Mario's speed (faster = higher jump)
		dx := math.Abs(player.velocity.X)
		jumpIndex := int(math.Min(float64(pigo8.Flr(dx)), 3)) // Clamp to array bounds
		player.velocity.Y = jumpForce[jumpIndex]
		player.inAir = true
		coyoteTime = 0 // Consume coyote time when jumping
	}

	// Apply gravity when in air
	if player.inAir {
		if player.velocity.Y < -2 && pigo8.Btn(pigo8.X) {
			// Slow gravity when holding jump (variable jump height)
			player.velocity.Y += gravitySlow / 16
		} else {
			// Normal gravity
			player.velocity.Y += gravityFast / 16
		}
	}
}

// updateFacingDirection makes Mario face the way he's moving
func updateFacingDirection(input playerInput) {
	switch {
	case input.isMovingRight:
		player.dir = 1
	case input.isMovingLeft:
		player.dir = -1
	}
}

// =============================================================================
// COLLISION AND MOVEMENT FUNCTIONS
// =============================================================================

// applyMovementAndCollision handles moving Mario and detecting collisions
func applyMovementAndCollision() {
	// Calculate how far to move this frame (clamp to prevent moving too fast)
	movement := player.velocity
	movement.X = math.Max(-maxMovementPerFrame, math.Min(maxMovementPerFrame, movement.X))
	if !player.inAir {
		movement.Y = 0 // Only apply vertical movement when in air
	} else {
		movement.Y = math.Min(maxMovementPerFrame, movement.Y)
	}

	// Try to move horizontally
	if movement.X != 0 {
		newPosition := player.position.Add(pigo8.NewVector2D(movement.X, 0))
		if !pigo8.MapCollision(newPosition.X, newPosition.Y, 0, 16) {
			// No collision - move to new position
			player.position.X = newPosition.X
		} else {
			// Collision detected - stop horizontal movement
			player.velocity.X = 0
		}
	}

	// Try to move vertically
	if movement.Y != 0 {
		newPosition := player.position.Add(pigo8.NewVector2D(0, movement.Y))
		if !pigo8.MapCollision(newPosition.X, newPosition.Y, 0, 16) {
			// No collision - move to new position
			player.position.Y = newPosition.Y
			if movement.Y > 0 {
				player.inAir = true // Still falling
			}
		} else {
			// Collision detected
			if movement.Y > 0 {
				player.inAir = false // Landed on ground
			}
			player.velocity.Y = 0 // Stop vertical movement
		}
	}
}

// checkLedgeDetection determines if Mario walked off a platform
func checkLedgeDetection() {
	if !player.inAir && player.velocity.Y == 0 {
		// Check if both feet are over empty space
		leftFoot := pigo8.MapCollision(player.position.X+2, player.position.Y+16, 0, 1)
		rightFoot := pigo8.MapCollision(player.position.X+13, player.position.Y+16, 0, 1)
		if !leftFoot && !rightFoot {
			player.inAir = true // Mario is falling
		}
	}
}

// updateGroundState keeps track of whether Mario is touching the ground
func updateGroundState() {
	player.onGround = !player.inAir
}

// updateCoyoteTime manages the grace period for jumping after leaving ground
func updateCoyoteTime() {
	if player.onGround {
		// Reset coyote time when on ground
		coyoteTime = coyoteTimeFrames // 6 frames of grace period (0.1 seconds at 60fps)
	} else if coyoteTime > 0 {
		// Decrease coyote time when in air
		coyoteTime -= 1.0
	}
}

// =============================================================================
// CAMERA FUNCTIONS - Using the new camera system!
// =============================================================================

// =============================================================================
// ANIMATION FUNCTIONS
// =============================================================================

// updateAnimation chooses which animation to play based on Mario's state
func updateAnimation(input playerInput) {
	var newFirst, newLast float64
	var newAnimRate float64

	// Check if Mario is skidding (moving one way but pressing the other direction)
	skidding := player.onGround && math.Abs(player.velocity.X) > skiddingThreshold &&
		((player.velocity.X > 0 && input.isMovingLeft) || (player.velocity.X < 0 && input.isMovingRight))

	// Choose animation based on current state
	switch {
	case skidding:
		// Skidding animation (static frame)
		newFirst, newLast, newAnimRate = playerSkidSX, playerSkidSX, 0
	case math.Abs(player.velocity.X) > walkingThreshold && player.onGround:
		// Walking animation (speed affects animation rate)
		newFirst, newLast = playerWalkAnimStartSX, playerWalkAnimEndSX
		newAnimRate = (math.Abs(player.velocity.X) / animationSpeedDivisor) * animationRateMultiplier
	case !player.onGround:
		// Jumping/falling animation (static frame)
		newFirst, newLast, newAnimRate = playerJumpFallSX, playerJumpFallSX, 0
	default:
		// Idle animation (static frame)
		newFirst, newLast, newAnimRate = playerIdleSX, playerIdleSX, 0
	}

	// Apply new animation state if it changed
	if player.first != newFirst || player.last != newLast {
		player.first, player.last = newFirst, newLast
		player.spriteIndex = player.first
		player.animationProgress = 0
	}
	player.currentAnimRate = newAnimRate
	player.animate()
}

// =============================================================================
// MAIN GAME STRUCTURE
// =============================================================================

// Game is the main game structure that implements the ebiten.Game interface
type Game struct{}

// Init is called once when the game starts
func (m *Game) Init() {
	// Create Mario at starting position with walking animation
	player = newEntity(playerIdleSX, playerStartX, playerStartY, 4.0, playerWalkAnimStartSX, playerWalkAnimEndSX, 1)

	// Configure camera with correct map dimensions for Mario demo
	// The actual map content extends to about 64x30 tiles = 512x240 pixels
	pigo8.SetCameraOptions(pigo8.CameraFollowOptions{
		Lerp:                   0.08,  // how fast the camera follows the player
		DeadZoneW:              16.0,  // how far the player can move on X before the camera starts moving
		DeadZoneH:              16.0,  // how far the player can move on Y before the camera starts moving
		LookAheadX:             24.0,  // how far ahead of the player the camera can look
		LookAheadY:             0.0,   // how far ahead of the player the camera can look
		HorizontalMovementOnly: true,  // whether the camera can move vertically
		ClampToMap:             true,  // whether the camera can move outside the map
		MapWidth:               512.0, // 64 tiles * 8 pixels (actual map content)
		MapHeight:              240.0, // 30 tiles * 8 pixels (actual map content)
	})

	// Initialize camera target to player's starting position
	// This is crucial - without this, the camera starts at (0,0) and shows nothing
	pigo8.SetCameraTarget(player.position.X, player.position.Y)
}

// Update is called every frame to handle game logic
func (m *Game) Update() {
	// Get all player input at the start of the frame
	input := getPlayerInput()

	// Handle all movement physics
	handleHorizontalMovement(input)
	handleVerticalMovement(input)
	updateFacingDirection(input)

	// Handle collision detection and movement
	applyMovementAndCollision()
	checkLedgeDetection()
	updateGroundState()
	updateCoyoteTime()

	// NES-style: Always update camera target every frame for consistency
	// This ensures the camera system always has the latest target position
	pigo8.SetCameraTarget(player.position.X, player.position.Y)

	// Update Mario's animation
	updateAnimation(input)
}

// Draw is called every frame to render the game
func (m *Game) Draw() {
	pigo8.Cls(7) // Clear screen with light gray

	// The new camera system automatically handles all drawing offsets!
	// No need to manually calculate camera positions anymore
	pigo8.Map()   // Draw the level (automatically offset by camera)
	player.draw() // Draw Mario (automatically offset by camera)

	// Debug: print camera and player positions
	camX, camY := pigo8.GetCameraPosition()
	pigo8.Print(fmt.Sprintf("Cam:%.0f,%.0f", camX, camY), 0, 0)
	pigo8.Print(fmt.Sprintf("Player:%.0f,%.0f", player.position.X, player.position.Y), 0, 10)
	pigo8.Print(fmt.Sprintf("Screen:256x240", 0, 20))
}

// =============================================================================
// GAME ENTRY POINT
// =============================================================================

func main() {
	// Configure global settings
	settings := pigo8.NewSettings()
	settings.WindowTitle = "Super Mario Bros - New Camera System Demo"
	settings.ScreenWidth = 256
	settings.ScreenHeight = 240
	settings.TargetFPS = 60

	// Camera system settings - these work with the new camera system
	settings.CameraLerp = 0.08     // Smooth camera movement
	settings.DeadZoneWidth = 16.0  // Dead zone for camera
	settings.DeadZoneHeight = 16.0 // Smaller vertical dead zone
	settings.LookAheadX = 24.0     // Look ahead in movement direction
	settings.LookAheadY = 0.0      // No vertical look-ahead
	settings.HorizontalMovementOnly = true

	// Map streaming settings for large levels
	settings.TileSize = 8       // Standard PICO-8 tile size
	settings.MapChunkSize = 32  // 32x32 tile chunks for streaming
	settings.MapChunkMargin = 2 // Keep 2 chunk margin around camera
	settings.ClampToMap = true

	// Start the game
	pigo8.InsertGame(&Game{})
	pigo8.PlayGameWith(settings)
}
