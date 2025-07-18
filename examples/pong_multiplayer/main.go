// Package main is a multiplayer Pong game using the PIGO8 networking functionality
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	p8 "github.com/drpaneas/pigo8"
	p8net "github.com/drpaneas/pigo8/network"
)

// Court boundaries
const (
	courtLeft   = 0
	courtRight  = 127
	courtTop    = 10
	courtBottom = 127
	centerX     = (courtRight + courtLeft) / 2
	lineLen     = 4
)

// RightSide side player/paddle
const RightSide = "Right"

// Paddle represents a player or remote paddle
type Paddle struct {
	x, y, width, height, speed float64
	color                      int
}

// Ball holds position, velocity, and rendering info
type Ball struct {
	x, y, size           float64
	dx, dy, speed, boost float64
	color                int
}

// GameState represents the serializable game state for network transmission
type GameState struct {
	BallX            float64 `json:"bx,omitempty"`
	BallY            float64 `json:"by,omitempty"`
	BallDX           float64 `json:"bdx,omitempty"`
	BallDY           float64 `json:"bdy,omitempty"`
	LeftPaddleY      float64 `json:"lpy,omitempty"`
	RightPaddleY     float64 `json:"rpy,omitempty"`
	LeftScore        int     `json:"ls,omitempty"`
	RightScore       int     `json:"rs,omitempty"`
	LastScored       string  `json:"scored,omitempty"`
	ResetBall        bool    `json:"reset,omitempty"`
	GameStarted      bool    `json:"started,omitempty"`
	WaitingForPlayer bool    `json:"waiting,omitempty"`
}

// PlayerInput represents the serializable player input for network transmission
type PlayerInput struct {
	Up   bool `json:"up,omitempty"`
	Down bool `json:"down,omitempty"`
}

// Game encapsulates all game state
type Game struct {
	leftPaddle       Paddle
	rightPaddle      Paddle
	ball             Ball
	leftScore        int
	rightScore       int
	lastScored       string
	isServer         bool
	isClient         bool
	gameStarted      bool
	waitingForPlayer bool
	lastStateUpdate  time.Time
	lastInputSent    time.Time
	remotePlayerID   string
}

// Init initializes the game state with default paddle and ball positions
func (g *Game) Init() {
	// Set up paddles with identical speed values
	g.leftPaddle = Paddle{8, 63, 2, 10, 1.5, 12}
	g.rightPaddle = Paddle{117, 63, 2, 10, 1.5, 8}

	// Set up ball with random direction
	ballDy := float64(p8.Flr(p8.Rnd(2))) - 0.5
	g.ball = Ball{
		x:     63,
		y:     63,
		size:  2,
		color: 7,
		dx:    0.6,
		dy:    ballDy,
		speed: 1,
		boost: 0.05,
	}

	// Play sound based on who scored
	switch g.lastScored {
	case "Left":
		p8.Music(3)
	case RightSide:
		p8.Music(4)
	default:
		p8.Music(5)
	}

	// Reset scores if this is a new game
	if g.lastScored == "" {
		g.leftScore = 0
		g.rightScore = 0
	}

	// Set network status
	g.isServer = p8net.IsServer()
	g.isClient = p8net.IsClient()
	g.gameStarted = false
	g.waitingForPlayer = true
	g.lastStateUpdate = time.Now()
	g.lastInputSent = time.Now()
}

// Update handles game logic each frame including input, AI, collisions and scoring
func (g *Game) Update() {
	// Check for network connection issues
	if p8net.IsConnectionLost() {
		// Network error is now handled by PIGO8 library
		return
	}

	// If we're waiting for a player, check if someone connected
	if p8net.IsWaitingForPlayers() {
		// Waiting state is now handled by PIGO8 library
		g.waitingForPlayer = true
		return
	} else if g.waitingForPlayer {
		// Player just connected
		g.waitingForPlayer = false
		g.gameStarted = true

		// Get the first connected player as remote player
		players := p8net.GetConnectedPlayers()
		if len(players) > 0 {
			g.remotePlayerID = players[0]
		}

		// Send initial game state if we're the server
		if g.isServer {
			g.sendGameState()
		}
	}

	// Handle local player input (left paddle for server, right paddle for client)
	if g.isServer {
		// Server controls left paddle
		if p8.Btn(p8.UP) && g.leftPaddle.y > courtTop+1 {
			g.leftPaddle.y -= g.leftPaddle.speed
		}
		if p8.Btn(p8.DOWN) && g.leftPaddle.y+g.leftPaddle.height < courtBottom-1 {
			g.leftPaddle.y += g.leftPaddle.speed
		}
	} else if g.isClient {
		// Client controls right paddle
		// Always send input state for more responsive control with UDP
		g.sendPlayerInput()

		// Implement client-side prediction for smoother feel
		// We'll still use the server's authoritative updates, but predict movement locally
		// for immediate visual feedback
		predictedY := g.rightPaddle.y

		// Calculate predicted position based on input
		if p8.Btn(p8.UP) && predictedY > courtTop+1 {
			predictedY -= g.rightPaddle.speed
		}
		if p8.Btn(p8.DOWN) && predictedY+g.rightPaddle.height < courtBottom-1 {
			predictedY += g.rightPaddle.speed
		}

		// Apply prediction, but allow server corrections if there's a significant difference
		// This provides responsive local movement while still allowing server to be authoritative
		g.rightPaddle.y = predictedY
	}

	// Server handles game logic
	if g.isServer && g.gameStarted {
		// Ball collisions with paddles
		if collide(g.ball, g.rightPaddle) {
			g.ball.dx = -(g.ball.dx + g.ball.boost)
			p8.Music(0)
		}
		if collide(g.ball, g.leftPaddle) {
			g.ball.dx = -(g.ball.dx - g.ball.boost)
			p8.Music(1)
		}

		// Ball collisions with top/bottom
		if g.ball.y <= courtTop+1 || g.ball.y+g.ball.size >= courtBottom-1 {
			g.ball.dy = -g.ball.dy
			p8.Music(2)
		}

		// Scoring
		if g.ball.x > courtRight {
			g.leftScore++
			g.lastScored = "Left"
			g.resetBall()
		}
		if g.ball.x < courtLeft {
			g.rightScore++
			g.lastScored = RightSide
			g.resetBall()
		}

		// Move ball
		g.ball.x += g.ball.dx
		g.ball.y += g.ball.dy

		// Send game state to client more frequently for better responsiveness
		// Use a very high update frequency for UDP networking to ensure smooth paddle movement
		if time.Since(g.lastStateUpdate) > 4*time.Millisecond {
			g.sendGameState()
			g.lastStateUpdate = time.Now()
		}
	}
}

// Draw renders the game elements to the screen each frame
func (g *Game) Draw() {
	p8.Cls(0)

	// Display network status using the standardized PIGO8 function
	if p8net.IsWaitingForPlayers() || p8net.GetNetworkError() != "" {
		p8net.DrawNetworkStatus()
		return
	}

	// Court outline
	p8.Rect(courtLeft, courtTop, courtRight, courtBottom, 5)

	// Center dashed line
	for y := courtTop; y < courtBottom; y += lineLen * 2 {
		p8.Line(centerX, float64(y), centerX, float64(y+lineLen), 5)
	}

	// Ball and paddles
	p8.Rectfill(g.ball.x, g.ball.y, g.ball.x+g.ball.size, g.ball.y+g.ball.size, g.ball.color)
	p8.Rectfill(g.leftPaddle.x, g.leftPaddle.y, g.leftPaddle.x+g.leftPaddle.width, g.leftPaddle.y+g.leftPaddle.height, g.leftPaddle.color)
	p8.Rectfill(g.rightPaddle.x, g.rightPaddle.y, g.rightPaddle.x+g.rightPaddle.width, g.rightPaddle.y+g.rightPaddle.height, g.rightPaddle.color)

	// Scores
	p8.Print(fmt.Sprint(g.leftScore), 30, 2, 7)
	p8.Print(fmt.Sprint(g.rightScore), 95, 2, 7)

	// Show role
	if g.isServer {
		p8.Print("Server (Left Paddle)", 10, courtBottom+3, 12)
	} else {
		p8.Print("Client (Right Paddle)", 10, courtBottom+3, 8)
	}
}

// resetBall resets the ball to the center with a new direction
func (g *Game) resetBall() {
	ballDy := float64(p8.Flr(p8.Rnd(2))) - 0.5
	g.ball.x = 63
	g.ball.y = 63
	g.ball.dx = 0.6
	if g.lastScored == RightSide {
		g.ball.dx = -g.ball.dx
	}
	g.ball.dy = ballDy
}

// sendGameState sends the current game state to the client
func (g *Game) sendGameState() {
	if !g.isServer {
		return
	}

	state := GameState{
		BallX:            g.ball.x,
		BallY:            g.ball.y,
		BallDX:           g.ball.dx,
		BallDY:           g.ball.dy,
		LeftPaddleY:      g.leftPaddle.y,
		RightPaddleY:     g.rightPaddle.y,
		LeftScore:        g.leftScore,
		RightScore:       g.rightScore,
		LastScored:       g.lastScored,
		GameStarted:      g.gameStarted,
		WaitingForPlayer: g.waitingForPlayer,
	}

	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling game state: %v", err)
		return
	}

	p8net.SendGameState(data, "all")
}

// sendPlayerInput sends the player's input to the server
func (g *Game) sendPlayerInput() {
	if !g.isClient || time.Since(g.lastInputSent) < 8*time.Millisecond {
		return
	}

	// Create input message with current button states
	input := PlayerInput{
		Up:   p8.Btn(p8.UP),
		Down: p8.Btn(p8.DOWN),
	}

	// Log the input being sent for debugging
	log.Printf("Client sending input: Up=%v, Down=%v", input.Up, input.Down)

	data, err := json.Marshal(input)
	if err != nil {
		log.Printf("Error marshaling player input: %v", err)
		return
	}

	// Send input to server
	p8net.SendPlayerInput(data)
	g.lastInputSent = time.Now()
}

// handleGameState processes game state received from the server
func handleGameState(_ string, data []byte) {
	game, ok := p8.CurrentCartridge().(*Game)
	if !ok {
		log.Printf("Error: current cartridge is not a Game")
		return
	}

	var state GameState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("Error unmarshaling game state: %v", err)
		return
	}

	// Log received game state for debugging
	log.Printf("Client received game state: ball=(%v,%v), left=%v, right=%v",
		state.BallX, state.BallY, state.LeftPaddleY, state.RightPaddleY)

	// Update game state
	game.ball.x = state.BallX
	game.ball.y = state.BallY
	game.ball.dx = state.BallDX
	game.ball.dy = state.BallDY

	// Update left paddle (server's paddle) from server state
	game.leftPaddle.y = state.LeftPaddleY

	// For the right paddle (client's paddle), we need to handle client-side prediction reconciliation
	if game.isClient {
		// Calculate the difference between our predicted position and server's position
		diff := math.Abs(game.rightPaddle.y - state.RightPaddleY)

		// If the difference is significant (more than a small threshold), use server position
		// This prevents the paddle from getting too far out of sync
		if diff > 3.0 {
			// Smoothly interpolate to the server position rather than snapping
			game.rightPaddle.y += (state.RightPaddleY - game.rightPaddle.y) * 0.5
		}
		// Otherwise, keep using our predicted position for smoother local movement
	} else {
		// For the server or spectators, always use the authoritative position
		game.rightPaddle.y = state.RightPaddleY
	}

	game.leftScore = state.LeftScore
	game.rightScore = state.RightScore
	game.lastScored = state.LastScored
	game.gameStarted = state.GameStarted
	game.waitingForPlayer = state.WaitingForPlayer
}

// handlePlayerInput processes player input received from the client
func handlePlayerInput(playerID string, data []byte) {
	game, ok := p8.CurrentCartridge().(*Game)
	if !ok {
		log.Printf("Error: current cartridge is not a Game")
		return
	}

	if !game.isServer {
		return
	}

	var input PlayerInput
	if err := json.Unmarshal(data, &input); err != nil {
		log.Printf("Error unmarshaling player input: %v", err)
		return
	}

	// Log received input for debugging
	log.Printf("Server received input from client %s: Up=%v, Down=%v", playerID, input.Up, input.Down)

	// Update right paddle based on client input
	// Use the same speed as the left paddle for consistency
	if input.Up && game.rightPaddle.y > courtTop+1 {
		game.rightPaddle.y -= game.rightPaddle.speed
	}
	if input.Down && game.rightPaddle.y+game.rightPaddle.height < courtBottom-1 {
		game.rightPaddle.y += game.rightPaddle.speed
	}

	// Immediately send updated game state after processing input
	game.sendGameState()
}

// handlePlayerConnect is called when a player connects
func handlePlayerConnect(playerID string) {
	game, ok := p8.CurrentCartridge().(*Game)
	if !ok {
		log.Printf("Error: current cartridge is not a Game")
		return
	}

	if !game.isServer {
		return
	}

	game.remotePlayerID = playerID
	game.waitingForPlayer = false
	game.gameStarted = true
}

// handlePlayerDisconnect is called when a player disconnects
func handlePlayerDisconnect(playerID string) {
	game, ok := p8.CurrentCartridge().(*Game)
	if !ok {
		log.Printf("Error: current cartridge is not a Game")
		return
	}

	if !game.isServer {
		return
	}

	if playerID == game.remotePlayerID {
		game.waitingForPlayer = true
		game.gameStarted = false
		game.remotePlayerID = ""
	}
}

// collide checks axis-aligned collision between ball and paddle
func collide(b Ball, p Paddle) bool {
	return b.x+b.size >= p.x && b.x <= p.x+p.width &&
		b.y+b.size >= p.y && b.y <= p.y+p.height
}

// No longer needed as we use p8.DrawNetworkStatus

func main() {
	// Create the game
	game := &Game{}
	p8.InsertGame(game)
	game.Init()

	// IMPORTANT: Register network callbacks BEFORE initializing the network
	log.Printf("Registering network callbacks...")
	p8net.SetOnGameStateCallback(handleGameState)
	p8net.SetOnPlayerInputCallback(handlePlayerInput)
	p8net.SetOnConnectCallback(handlePlayerConnect)
	p8net.SetOnDisconnectCallback(handlePlayerDisconnect)

	// Configure the game settings
	settings := p8.NewSettings()
	settings.TargetFPS = 60
	settings.WindowTitle = "PIGO8 Multiplayer Pong"
	// Enable multiplayer
	settings.Multiplayer = true

	// Initialize network manually first to ensure callbacks are registered
	config := p8net.ParseNetworkArgs()
	config.GameName = "PIGO8 Multiplayer Pong"
	if err := p8net.InitNetwork(config); err != nil {
		log.Printf("Error initializing network: %v", err)
	}

	// Force register callbacks directly on the network manager as a fallback
	log.Printf("Force registering callbacks to ensure they're set...")
	p8net.ForceRegisterCallbacks(
		handleGameState,
		handlePlayerInput,
		handlePlayerConnect,
		handlePlayerDisconnect,
	)

	// Verify callbacks are registered
	if !p8net.AreCallbacksRegistered() {
		log.Printf("WARNING: Callbacks are still not registered after force registration!")
	} else {
		log.Printf("SUCCESS: All callbacks are now registered")
	}

	// We need to modify PlayGameWith to avoid reinitializing the network if it's already initialized
	// For now, we'll just call PlayGameWith directly
	p8.PlayGameWith(settings)
}
