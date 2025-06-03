// Package main basic super mario bros example
package main

import (
	"math"

	. "github.com/drpaneas/pigo8"
)

// Game specific constants
const (
	// ScreenWidth and ScreenHeight are defined in the pigo8 package
	CameraY = 0 // Always 0 for horizontal platformers like Mario

	// Static region sizes (customize as needed)
	StaticRegionWidth  = 80.0  // Mario can move this far left/right before camera follows
	StaticRegionHeight = 120.0 // (Not used much in Mario, mostly horizontal)

	// Offset so Mario sees more *ahead*
	StaticRegionForwardOffset = 32.0 // How much more the box leans forward
)

// Game specific constants
const (
	// Player animation sprite X coordinates (sourceX on spritesheet)
	playerIdleSX          = 8                                                   // Idle animation frame
	playerWalkAnimStartSX = 24                                                  // Start of 3-frame walk animation (24, 40, 56)
	playerWalkAnimFrames  = 3                                                   // Number of frames in walk animation
	playerWalkAnimEndSX   = playerWalkAnimStartSX + (playerWalkAnimFrames-1)*16 // End of walk: 24 + (2*16) = 56
	playerSkidSX          = 72
	playerJumpFallSX      = 88

	// Player animation sprite Y coordinate (sourceY on spritesheet)
	playerSpriteSY = 24

	// Tile flags
	tileSolidFlag = 0 // Assuming flag 0 means solid ground
)

// entity represents a game entity with animation capabilities
type entity struct {
	sprite16, x, y, first, last, dir float64
	width, height                    float64 // Collision dimensions
	spriteSourceW, spriteSourceH     float64 // Source sprite dimensions on spritesheet
	spriteDestW, spriteDestH         float64 // Destination sprite dimensions on screen
	spriteSourceY                    float64 // Source sprite Y offset on spritesheet
	vx                               float64 // Horizontal velocity (sign indicates direction)
	maxSpeed                         float64 // Current maximum horizontal speed (dynamic based on walking/running)
	walkSpeed                        float64 // Max speed when walking
	runSpeed                         float64 // Max speed when running
	acceleration                     float64 // Rate of horizontal acceleration
	deceleration                     float64 // Rate of horizontal deceleration
	skidAcceleration                 float64 // Acceleration rate when skidding
	isSkidding                       bool    // True if player is in skid animation

	vy                        float64 // Vertical velocity
	gravity                   float64
	jumpStrength              float64
	onGround                  bool
	maxFallSpeed              float64 // Maximum downward speed
	jumpHoldTimer             int     // Frames jump button has been held during ascent
	jumpHoldDurationFrames    int     // Max frames jump hold affects height
	jumpHoldGravityMultiplier float64 // Factor to reduce gravity during jump hold

	// Animation rate fields
	baseWalkAnimRate           float64 // Target animation rate when at walkSpeed
	baseRunAnimRate            float64 // Target animation rate when at runSpeed
	currentBaseAnimRate        float64 // Current target animation rate (either walk or run)
	animationProgress          float64 // Progress towards next animation frame change
	currentActualAnimationRate float64 // The actual rate used by animate() based on currentSpeed
}

const (
	playerWidth        = 14.0 // Player collision width (slightly less than sprite for better feel)
	playerHeight       = 16.0 // Player collision box height
	playerJumpStrength = 3.5  // Initial jump velocity
)

var (
	player  entity
	cameraX float64 = 0 // horizontal camera position
)

// newEntity creates a new animated entity with the given parameters
func newEntity(sprite, x, y, maxAnimRate, maxSpeedVal, firstFrame, lastFrame, initialDir float64) entity {
	return entity{
		sprite16:                   sprite,
		x:                          x,
		y:                          y,
		width:                      playerWidth,
		height:                     playerHeight,
		spriteSourceW:              16,             // Default small Mario sprite width
		spriteSourceH:              16,             // Default small Mario sprite height
		spriteDestW:                16,             // Default small Mario render width
		spriteDestH:                16,             // Default small Mario render height
		spriteSourceY:              playerSpriteSY, // Default Y offset on spritesheet
		vx:                         0.0,
		walkSpeed:                  maxSpeedVal,         // Initialize walkSpeed with the base speed (e.g., 1.5)
		runSpeed:                   maxSpeedVal * 1.667, // Run speed (e.g., 1.5 * 1.667 = ~2.5)
		maxSpeed:                   maxSpeedVal,         // Initially, maxSpeed is walkSpeed
		acceleration:               0.055,               // Default acceleration rate (SMB3: 14/256 ≈ 0.055)
		deceleration:               0.055,               // Default deceleration rate (SMB3 friction: ~0.054, skid: 0.125)
		skidAcceleration:           0.125,               // Skid acceleration rate (SMB3: ~0.125)
		isSkidding:                 false,
		vy:                         0.0,
		gravity:                    0.3125,             // SMB3 GRAVITY_FAST / 16
		jumpStrength:               playerJumpStrength, // SMB3 JUMP_FORCE base: 3.5 (negative in JS)
		onGround:                   false,
		maxFallSpeed:               4.0, // SMB3 matches
		jumpHoldTimer:              0,
		jumpHoldDurationFrames:     24,                // Keep our frame-based hold
		jumpHoldGravityMultiplier:  0.2,               // To achieve SMB3 GRAVITY_SLOW / 16 effect (0.0625) with our new gravity
		baseWalkAnimRate:           maxAnimRate,       // Base animation rate for walking
		baseRunAnimRate:            maxAnimRate * 1.5, // Base animation rate for running (e.g., 50% faster)
		currentBaseAnimRate:        maxAnimRate,       // Initially, use walk animation rate
		animationProgress:          0.0,               // Start progress at 0
		currentActualAnimationRate: 0.0,               // Initial actual rate
		first:                      firstFrame,
		last:                       lastFrame,
		dir:                        initialDir,
	}
}

// animate updates the sprite index based on timing to create animation
const animationFrameThreshold = 16.0 // The target value for progress

func (e *entity) animate() {
	if e.currentActualAnimationRate <= 0 { // Don't animate if rate is zero or negative
		return
	}
	e.animationProgress += e.currentActualAnimationRate // Accumulate progress based on the *actual* rate

	if e.animationProgress >= animationFrameThreshold { // Check if enough progress has been made
		e.animationProgress -= animationFrameThreshold // Subtract threshold, keeping any remainder for smooth fractional rates

		// Advance to the next frame of animation
		e.sprite16 += 16
		if e.sprite16 > e.last { // If current frame exceeds the last frame, loop back to the start of the current sequence
			e.sprite16 = e.first
		}
	}
}

// Draw renders the entity to the screen
func (e *entity) draw() {
	flip := false
	if e.dir < 0 {
		flip = true
	}
	Sspr(int(e.sprite16), int(e.spriteSourceY), int(e.spriteSourceW), int(e.spriteSourceH), int(math.Round(e.x)), int(math.Round(e.y)), int(e.spriteDestW), int(e.spriteDestH), flip, false)
}

type Game struct{}

func (m *Game) Init() {
	// player = newEntity(initialSprite, x, y, maxAnimRate, walkSpeed, walk_firstFrame, walk_lastFrame, initialDirection)
	player = newEntity(playerIdleSX, 32, 59, 2.5, 1.5, playerWalkAnimStartSX, playerWalkAnimEndSX, 1) // Changed initial X from -8 to 32

}

func (m *Game) Update() {
	isMovingLeft := Btn(LEFT)
	isMovingRight := Btn(RIGHT)
	jumpPressed := Btnp(X) // Using Btnp for jump to avoid continuous jumping if held
	isRunning := Btn(O)    // Check if O button is held for running

	// Update player's max speed based on whether they are running
	if isRunning {
		player.maxSpeed = player.runSpeed
		player.currentBaseAnimRate = player.baseRunAnimRate
	} else {
		player.maxSpeed = player.walkSpeed
		player.currentBaseAnimRate = player.baseWalkAnimRate
	}

	// --- Horizontal Movement (SMB3 Style) ---
	inputDir := 0.0
	if isMovingRight {
		inputDir = 1.0
	} else if isMovingLeft {
		inputDir = -1.0
	}

	player.isSkidding = false // Reset skid state at the beginning of horizontal logic for the frame

	if inputDir == 0 { // No horizontal input
		if player.onGround { // Apply friction only on ground
			if player.vx > 0 {
				player.vx -= player.deceleration
				if player.vx < 0 {
					player.vx = 0
				}
			} else if player.vx < 0 {
				player.vx += player.deceleration
				if player.vx > 0 {
					player.vx = 0
				}
			}
		}
	} else { // Player is pressing left or right
		player.dir = inputDir // Update facing direction
		absVx := math.Abs(player.vx)

		// Condition 1: Skidding (on ground, current velocity exists, input is opposite to current velocity)
		if player.onGround && player.vx != 0 && (math.Copysign(1.0, player.vx) == -inputDir) {
			player.vx += inputDir * player.skidAcceleration
			player.isSkidding = true // Set skid state
		} else {
			// Not skidding: either accelerating or over speed limit.
			// player.isSkidding is already false from the top of this logic block.

			if absVx < player.maxSpeed {
				// Condition 2: Normal acceleration (speed is less than maxSpeed, or starting, or moving in same direction as input)
				player.vx += inputDir * player.acceleration
				// If this acceleration pushes vx over maxSpeed in the direction of input, clamp it.
				if math.Copysign(1.0, player.vx) == inputDir && math.Abs(player.vx) > player.maxSpeed {
					player.vx = inputDir * player.maxSpeed
				}
			} else if player.onGround && absVx > player.maxSpeed && (math.Copysign(1.0, player.vx) == inputDir) {
				// Condition 3: Speed > maxSpeed, on ground, and input is in the same direction of motion
				// Apply friction to slow down towards maxSpeed.
				if player.vx > 0 { // Moving right (inputDir is 1)
					player.vx -= player.deceleration // Apply friction against movement
					if player.vx < player.maxSpeed { // Don't undershoot if friction is too strong
						player.vx = player.maxSpeed
					}
				} else { // Moving left (inputDir is -1)
					player.vx += player.deceleration  // Apply friction against movement
					if player.vx > -player.maxSpeed { // Don't undershoot
						player.vx = -player.maxSpeed
					}
				}
			}
			// If absVx == player.maxSpeed and input is in the same direction, no acceleration or friction is applied by these conditions,
			// so speed is maintained. This matches the JS reference behavior.
		}
	}
	// --- End Horizontal Movement ---

	dx := player.vx

	// Horizontal Collision Check
	// Check one pixel ahead for collision, or at player.width/2 for center checks
	// For simplicity, we check the leading edge based on direction.
	nextX := player.x + dx
	collisionRectX := nextX
	if dx < 0 { // Moving left
		collisionRectX = nextX
	} else if dx > 0 { // Moving right
		collisionRectX = nextX // + player.width -1 ; check leading edge
	}

	if dx != 0 && MapCollision(int(collisionRectX), int(player.y), tileSolidFlag, int(player.width), int(player.height)) {
		// If collision, snap to tile edge and stop horizontal movement
		if dx > 0 { // Moving right, hit wall on right
			player.x = float64(int(player.x/8+1)*8) - player.width - 0.01 // Snap to left of tile
		} else { // Moving left, hit wall on left
			player.x = float64(int(player.x/8)*8) + 0.01 // Snap to right of tile
		}
		dx = 0 // Stop horizontal movement
		player.vx = 0
	} else {
		player.x = nextX
	}

	// --- Camera Logic Below ---

	// Find Mario's static region boundaries:
	// (static region is always relative to the camera, not the world!)
	// Camera always shows cameraX .. cameraX+ScreenWidth
	// We want Mario to be allowed to walk within [regionMin, regionMax] before the camera moves

	regionMin := cameraX + (float64(ScreenWidth)-StaticRegionWidth)/2 - StaticRegionForwardOffset/2
	regionMax := cameraX + (float64(ScreenWidth)+StaticRegionWidth)/2 + StaticRegionForwardOffset/2

	// If Mario moves beyond right edge
	if player.x+player.width > regionMax {
		cameraX += (player.x + player.width - regionMax)
	}
	// If Mario moves beyond left edge
	if player.x < regionMin {
		cameraX -= (regionMin - player.x)
	}

	// Clamp cameraX to 0 (don't scroll beyond the left edge of the level)
	if cameraX < 0 {
		cameraX = 0
	}

	// Optionally, clamp cameraX to right edge of the level if you have a fixed level size
	// Example: if levelWidth > ScreenWidth { if cameraX > levelWidth - ScreenWidth { cameraX = levelWidth - ScreenWidth } }

	// --- Vertical Movement ---
	// Jump action
	if jumpPressed && player.onGround {
		// Variable jump strength based on horizontal speed (matching SMB3)
		dx := math.Floor(math.Abs(player.vx))
		jumpForce := player.jumpStrength

		// Apply speed-based jump boost (similar to JUMP_FORCE array in SMB3 JS)
		if dx >= 1 {
			// Add between 0.125 and 0.5 based on speed (0-3)
			speedBoost := math.Min(dx, 3) * 0.125
			jumpForce += speedBoost
		}

		player.vy = -jumpForce
		player.onGround = false  // Player is now airborne
		player.jumpHoldTimer = 0 // Reset jump hold timer
	}

	// Apply gravity and handle jump hold
	if !player.onGround {
		effectiveGravity := player.gravity
		// If jump button is held, player is moving up, and hold duration not exceeded
		if Btn(X) && player.vy < 0 && player.jumpHoldTimer < player.jumpHoldDurationFrames {
			effectiveGravity *= player.jumpHoldGravityMultiplier
			player.jumpHoldTimer++
		}
		player.vy += effectiveGravity

		// Cap falling speed
		if player.vy > player.maxFallSpeed {
			player.vy = player.maxFallSpeed
		}
	}

	dy := player.vy
	nextY := player.y + dy

	// // --- DEBUG LOGGING START ---
	// fmt.Printf("[DEBUG] y=%.2f vy=%.2f onGround=%v\n", player.y, player.vy, player.onGround)
	// // --- DEBUG LOGGING END ---

	// Vertical Collision Check
	// For simplicity, check bottom center point for ground, top for ceiling.
	// A more robust check would use the full player.width.
	checkX := player.x + player.width/2 // Check middle of player for vertical collision

	var groundCollision1, groundCollision2 bool
	if dy > 0 { // Moving Down
		groundCollision1 = MapCollision(int(checkX-player.width/4), int(nextY+player.height-1), tileSolidFlag, int(player.width/2), 1)
		groundCollision2 = MapCollision(int(player.x), int(nextY+player.height-1), tileSolidFlag, int(player.width), 1)
		// fmt.Printf("[DEBUG] Downward MapCollision1=%v MapCollision2=%v\n", groundCollision1, groundCollision2)
		if groundCollision1 || groundCollision2 { // Check feet
			// fmt.Printf("[DEBUG] Player SNAPPED to ground at y=%.2f\n", float64(int((nextY+player.height)/8)*8)-player.height)
			player.y = float64(int((nextY+player.height)/8)*8) - player.height // Align with ground
			player.vy = 0
			player.onGround = true
			dy = 0
		} else {
			// fmt.Printf("[DEBUG] Player is AIRBORNE after collision check\n")
			player.onGround = false
		}
	} else if dy < 0 { // Moving Up
		if MapCollision(int(checkX-player.width/4), int(nextY), tileSolidFlag, int(player.width/2), 1) ||
			MapCollision(int(player.x), int(nextY), tileSolidFlag, int(player.width), 1) { // Check head
			player.y = float64(int(nextY/8+1) * 8) // Align with ceiling
			player.vy = 0
			dy = 0
		}
	}
	player.y += dy // Apply vertical movement if no collision or after adjustment

	// --- FINAL GROUND CHECK: Ensure onGround is correct even if not moving vertically ---
	feetX := player.x + player.width/2
	feetY := player.y + player.height
	tileX := int(feetX) / 8
	tileY := int(feetY) / 8
	spriteID := Mget(tileX, tileY)
	_, isSolid := Fget(spriteID, tileSolidFlag)
	if isSolid {
		player.onGround = true
	} else {
		player.onGround = false
	}

	// --- Animation Control ---
	var newFirst, newLast float64
	var newAnimRate float64
	isNewAnimationState := false // Flag to check if we are transitioning to a new animation sequence

	// Determine current animation state and its parameters
	if player.isSkidding && player.onGround { // SKIDDING
		newFirst = playerSkidSX
		newLast = playerSkidSX
		newAnimRate = 0
		// Check if we are transitioning into skidding from a different animation state
		if player.first != newFirst || player.last != newLast {
			isNewAnimationState = true
		}
	} else if !player.onGround { // JUMPING/FALLING
		newFirst = playerJumpFallSX
		newLast = playerJumpFallSX
		newAnimRate = 0
		// Check if we are transitioning into jumping/falling
		if player.first != newFirst || player.last != newLast {
			isNewAnimationState = true
		}
	} else if math.Abs(player.vx) > 0 && player.onGround { // WALKING
		newFirst = playerWalkAnimStartSX
		newLast = playerWalkAnimEndSX
		normalizedSpeed := math.Abs(player.vx) / player.maxSpeed
		newAnimRate = normalizedSpeed * player.currentBaseAnimRate
		// Check if we are transitioning into walking (e.g., from idle, jump, or skid)
		// or if walking animation was previously stopped (rate was 0), indicating a restart of walk.
		if player.first != newFirst || player.last != newLast || player.currentActualAnimationRate == 0 {
			isNewAnimationState = true
		}
	} else { // IDLE (vx is 0 and onGround)
		newFirst = playerIdleSX
		newLast = playerIdleSX
		newAnimRate = 0
		// Check if we are transitioning into idle
		if player.first != newFirst || player.last != newLast {
			isNewAnimationState = true
		}
	}

	// Apply new animation state parameters
	player.first = newFirst
	player.last = newLast
	player.currentActualAnimationRate = newAnimRate

	// If the animation state has changed (e.g., jump to walk, walk to idle),
	// reset the sprite to the first frame of the new sequence and reset animation progress.
	// This ensures animations start correctly from their beginning.
	if isNewAnimationState {
		player.sprite16 = player.first
		player.animationProgress = 0
	}

	player.animate() // Call animate unconditionally to handle frame progression or static display
}

func (m *Game) Draw() {
	Cls(7)
	// Pass float64 cameraX to Camera; rounding is handled inside the engine for pixel-perfect rendering
	Camera(cameraX, CameraY)
	Map()
	player.draw()
	// Debug: print camera and player positions when camera moves
	// if math.Abs(player.vx) > 0.01 || math.Abs(player.x-cameraX) > 0.01 {
	// 	fmt.Printf("cameraX: %.3f (rounded: %.0f), player.x: %.3f (rounded: %.0f)\n", cameraX, math.Round(cameraX), player.x, math.Round(player.x))
	// }
}

func main() {
	// Configure the game settings
	settings := NewSettings()
	settings.WindowTitle = "Super Mario Bros"
	settings.ScreenWidth = 256
	settings.ScreenHeight = 240
	settings.TargetFPS = 60

	// Start the game
	InsertGame(&Game{})
	PlayGameWith(settings)
}
