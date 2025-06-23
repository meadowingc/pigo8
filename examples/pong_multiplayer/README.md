# PIGO8 Multiplayer Pong

This example demonstrates how to convert a single-player game to multiplayer using PIGO8's networking capabilities. The original Pong example has been transformed to support two players over a network connection.

## How Single-Player Pong Was Converted to Multiplayer

### Game Structure Changes

The single-player Pong was converted to multiplayer by making several key architectural changes:

1. **Paddle Ownership**:
   - In single-player mode, one paddle was controlled by the player and the other by AI
   - In multiplayer mode, the server player controls the left paddle, and the client player controls the right paddle

2. **Naming Convention**:
   - `player` and `computer` paddles became `leftPaddle` and `rightPaddle`
   - `playerScore` and `computerScore` became `leftScore` and `rightScore`

   **Single-player structure:**

   ```go
   // Game encapsulates all game state
   type Game struct {
       player        Paddle
       computer      Paddle
       ball          Ball
       playerScore   int
       computerScore int
       Scored        string
   }
   ```

   **Multiplayer structure:**

   ```go
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
       networkError     string
   }
   ```

### Network Communication Architecture

The multiplayer implementation follows a client-server model with serializable data structures:

```go
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
```

#### Server-Side Logic

The server controls the left paddle and manages game physics:

```go
// Update handles game logic each frame including input, physics, and networking
func (g *Game) Update() {
    // Check for network connection issues
    if p8.IsConnectionLost() {
        return
    }

    // Handle waiting for players
    if p8.IsWaitingForPlayers() {
        g.waitingForPlayer = true
        return
    }

    // Server-specific logic
    if g.isServer && g.gameStarted {
        // Handle left paddle input
        if p8.Btn(p8.ButtonUp) && g.leftPaddle.y > courtTop+1 {
            g.leftPaddle.y -= g.leftPaddle.speed
        }
        if p8.Btn(p8.ButtonDown) && g.leftPaddle.y+g.leftPaddle.height < courtBottom-1 {
            g.leftPaddle.y += g.leftPaddle.speed
        }

        // Ball physics and collision detection
        // ...

        // Send game state to client regularly
        if time.Since(g.lastStateUpdate) > 16*time.Millisecond {
            g.sendGameState()
            g.lastStateUpdate = time.Now()
        }
    }
}
```

#### Client-Side Logic

The client sends input and applies received game state:

```go
// Client controls right paddle and sends input to server
if g.isClient {
    // Send input when buttons are pressed
    if p8.Btn(p8.ButtonUp) || p8.Btn(p8.ButtonDown) {
        g.sendPlayerInput()
    }
}

// sendPlayerInput sends the player's input to the server
func (g *Game) sendPlayerInput() {
    if !g.isClient || time.Since(g.lastInputSent) < 16*time.Millisecond {
        return
    }

    input := PlayerInput{
        Up:   p8.Btn(p8.ButtonUp),
        Down: p8.Btn(p8.ButtonDown),
    }

    data, err := json.Marshal(input)
    if err != nil {
        log.Printf("Error marshaling player input: %v", err)
        return
    }

    p8.SendPlayerInput(data)
    g.lastInputSent = time.Now()
}
```

### State Synchronization

The server sends complete game state to clients:

```go
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

    p8.SendGameState(data, "all")
}
```

The client applies received state:

```go
// handleGameState processes game state received from the server
func handleGameState(playerID string, data []byte) {
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

    // Update game state
    game.ball.x = state.BallX
    game.ball.y = state.BallY
    game.ball.dx = state.BallDX
    game.ball.dy = state.BallDY
    game.leftPaddle.y = state.LeftPaddleY
    game.rightPaddle.y = state.RightPaddleY
    game.leftScore = state.LeftScore
    game.rightScore = state.RightScore
    game.lastScored = state.LastScored
    game.gameStarted = state.GameStarted
    game.waitingForPlayer = state.WaitingForPlayer
}
```

### Latency Compensation

To compensate for network latency, the server applies more movement per input:

```go
// handlePlayerInput processes player input received from the client
func handlePlayerInput(playerID string, data []byte) {
    game, ok := p8.CurrentCartridge().(*Game)
    if !ok || !game.isServer {
        return
    }

    var input PlayerInput
    if err := json.Unmarshal(data, &input); err != nil {
        log.Printf("Error unmarshaling player input: %v", err)
        return
    }

    // Apply more movement per input to compensate for network latency
    if input.Up && game.rightPaddle.y > courtTop+1 {
        game.rightPaddle.y -= game.rightPaddle.speed * 2
    }
    if input.Down && game.rightPaddle.y+game.rightPaddle.height < courtBottom-1 {
        game.rightPaddle.y += game.rightPaddle.speed * 2
    }
}
```

### Network Status Display

The game displays network status to provide feedback to players:

```go
// Draw renders the game elements to the screen each frame
func (g *Game) Draw() {
    p8.Cls(0)

    // Display network status using the standardized PIGO8 function
    if p8.IsWaitingForPlayers() || p8.GetNetworkError() != "" {
        p8.DrawNetworkStatus(networkStatusX, networkStatusY, networkTextColor)
        return
    }

    // Draw game elements
    // ...

    // Show role
    if g.isServer {
        p8.Print("Server (Left Paddle)", 10, courtBottom+3, 12)
    } else {
        p8.Print("Client (Right Paddle)", 10, courtBottom+3, 8)
    }
}
```

### Network Initialization

The game initializes the network and sets up callbacks:

```go
func main() {
    // Parse command line flags using the standardized PIGO8 function
    config := p8.ParseNetworkArgs()
    config.GameName = "PIGO8 Multiplayer Pong"

    // Initialize network
    if err := p8.InitNetwork(config); err != nil {
        fmt.Printf("Error initializing network: %v\n", err)
        os.Exit(1)
    }
    defer p8.ShutdownNetwork()

    // Set up network callbacks
    p8.SetOnGameStateCallback(handleGameState)
    p8.SetOnPlayerInputCallback(handlePlayerInput)
    p8.SetOnConnectCallback(handlePlayerConnect)
    p8.SetOnDisconnectCallback(handlePlayerDisconnect)

    // Create and start the game
    game := &Game{}
    p8.InsertGame(game)
    game.Init()

    settings := p8.NewSettings()
    settings.TargetFPS = 60
    settings.WindowTitle = "PIGO-8 Multiplayer Pong"
    p8.PlayGameWith(settings)
}
```

## Running the Multiplayer Game

To run the multiplayer game:

1. **Server Mode**:

   ```bash
   ./pong_multiplayer -server
   ```

2. **Client Mode**:

   ```bash
   ./pong_multiplayer -connect <server_ip>
   ```

The game uses PIGO8's built-in networking infrastructure to handle connections, state synchronization, and player input.

This example serves as a template for converting other single-player PIGO8 games to multiplayer by following the same patterns and utilizing the PIGO8 networking API.

## Migrating from Old Multiplayer API to New Network API

The PIGO8 library has undergone a refactoring of its multiplayer functionality. If you have existing code using the older multiplayer API, here's how to migrate to the newer, more robust network API:

### 1. Configuration

**Old API:**

```go
// Using Settings struct
settings := p8.NewSettings()
settings.WithMultiplayer(true)
settings.IsMultiplayerHost()
settings.ConnectTo("192.168.1.100")
settings.WithPort(8080)

// Or using MultiplayerSettings
multiSettings := p8.ParseMultiplayerArgs()
p8.InitMultiplayer(multiSettings)
```

**New API:**

```go
// Using NetworkConfig
config := p8.ParseNetworkArgs() // Parses command line arguments
// Or create manually
config := &p8.NetworkConfig{
    Role:     p8.RoleServer, // or p8.RoleClient
    Address:  "192.168.1.100", // For client, empty for server
    Port:     8080,
    GameName: "My Game",
}
p8.InitNetwork(config)
```

### 2. State Synchronization

**Old API:**

```go
// Register entire game state for automatic sync
p8.RegisterGameState(&myGame)
// State sync happens automatically
```

**New API:**

```go
// Define serializable state structures
type GameState struct {
    // Game state fields with json tags
    Position float64 `json:"pos"`
}

// Explicitly send state (server)
state := GameState{Position: player.x}
data, _ := json.Marshal(state)
p8.SendGameState(data, "all")

// Receive state (client)
p8.SetOnGameStateCallback(func(playerID string, data []byte) {
    var state GameState
    json.Unmarshal(data, &state)
    // Apply state
})
```

### 3. Network Status

**Old API:**
Limited network status handling.

**New API:**

```go
// Check connection status
if p8.IsConnectionLost() {
    // Handle disconnection
}

// Check if waiting for players
if p8.IsWaitingForPlayers() {
    // Show waiting message
}

// Display network status
p8.DrawNetworkStatus(x, y, color)

// Get connected players
players := p8.GetConnectedPlayers()
```

### 4. Bridge Function

If you need to gradually migrate, you can use the bridge function:

```go
// Convert old settings to new network config
multiSettings := p8.DefaultMultiplayerSettings()
// Configure multiSettings...
p8.InitNetworkFromMultiplayerSettings(multiSettings)
```

The new network API provides more explicit control over networking, better error handling, and improved performance.
