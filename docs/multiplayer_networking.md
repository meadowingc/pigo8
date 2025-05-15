# PIGO8 Multiplayer Networking Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Networking Architecture](#networking-architecture)
3. [Setting Up a Multiplayer Game](#setting-up-a-multiplayer-game)
4. [Network Callbacks](#network-callbacks)
5. [Message Types and Data Structures](#message-types-and-data-structures)
6. [Client-Side Prediction](#client-side-prediction)
7. [Case Study: Multiplayer Gameboy](#case-study-multiplayer-gameboy)
8. [Advanced Topics](#advanced-topics)
9. [Troubleshooting](#troubleshooting)

## Introduction

PIGO8 provides a built-in networking system that allows developers to add multiplayer functionality to their games. This guide will walk you through the process of converting a single-player game to multiplayer, with a focus on the Gameboy example.

The networking system in PIGO8 uses UDP for low-latency communication, making it suitable for real-time games. It follows a client-server architecture where one instance of the game acts as the server and other instances connect as clients.

## Networking Architecture

### Client-Server Model

PIGO8 uses a client-server architecture for multiplayer games:

- **Server**: Authoritative source of game state, processes game logic
- **Clients**: Send inputs to the server, receive and display game state

### UDP Protocol

PIGO8 uses UDP (User Datagram Protocol) for networking:

- **Advantages**: Lower latency, better for real-time games
- **Challenges**: No guaranteed delivery, packets may arrive out of order

### Network Roles

Each instance of a PIGO8 game can have one of two roles:

- **Server**: Hosts the game, processes game logic, broadcasts game state
- **Client**: Connects to server, sends inputs, receives game state

## Setting Up a Multiplayer Game

### Step 1: Initialize Network Roles

First, modify your game structure to include network-related fields:

```go
type Game struct {
    // Existing game fields
    
    // Network-related fields
    isServer        bool
    isClient        bool
    remotePlayerID  string
    lastStateUpdate time.Time
}

func (g *Game) Init() {
    // Existing initialization code
    
    // Set up network roles
    g.isServer = p8.GetNetworkRole() == p8.RoleServer
    g.isClient = p8.GetNetworkRole() == p8.RoleClient
    g.lastStateUpdate = time.Now()
}
```

### Step 2: Register Network Callbacks

In your `main()` function, register the network callbacks before starting the game:

```go
func main() {
    // Create and initialize the game
    game := &Game{}
    p8.InsertGame(game)
    game.Init()
    
    // Register network callbacks
    p8.SetOnGameStateCallback(handleGameState)
    p8.SetOnPlayerInputCallback(handlePlayerInput)
    p8.SetOnConnectCallback(handlePlayerConnect)
    p8.SetOnDisconnectCallback(handlePlayerDisconnect)
    
    // Start the game
    settings := p8.NewSettings()
    settings.WindowTitle = "PIGO8 Multiplayer Game"
    p8.PlayGameWith(settings)
}
```

### Step 3: Modify Game Loop for Network Play

Update your game's `Update()` function to handle network roles:

```go
func (g *Game) Update() {
    // Handle waiting for players state
    if p8.IsWaitingForPlayers() {
        return
    }
    
    // Server-specific logic
    if g.isServer {
        // Process game physics, AI, etc.
        
        // Send game state updates
        if time.Since(g.lastStateUpdate) > 16*time.Millisecond {
            g.sendGameState()
            g.lastStateUpdate = time.Now()
        }
    }
    
    // Client-specific logic
    if g.isClient {
        // Send player input
        g.sendPlayerInput()
        
        // Client-side prediction (optional)
    }
    
    // Common logic for both client and server
}
```

### Step 4: Add Network Status Display

Update your game's `Draw()` function to display network status:

```go
func (g *Game) Draw() {
    p8.Cls(0)
    
    // Display network status
    if p8.IsWaitingForPlayers() || p8.GetNetworkError() != "" {
        p8.DrawNetworkStatus(10, 10, 7)
        return
    }
    
    // Draw game elements
    // ...
}
```

## Network Callbacks

PIGO8 requires four callback functions for multiplayer functionality:

### Game State Callback

Receives game state updates from the server:

```go
func handleGameState(playerID string, data []byte) {
    game, ok := p8.CurrentCartridge().(*Game)
    if !ok {
        return
    }
    
    var state GameState
    if err := json.Unmarshal(data, &state); err != nil {
        log.Printf("Error unmarshaling game state: %v", err)
        return
    }
    
    // Update game state with received data
    // Example: Update player positions, game objects, etc.
}
```

### Player Input Callback

Processes player input on the server:

```go
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
    
    // Update game state based on player input
    // Example: Move remote player based on input
}
```

### Connect Callback

Handles new player connections:

```go
func handlePlayerConnect(playerID string) {
    game, ok := p8.CurrentCartridge().(*Game)
    if !ok || !game.isServer {
        return
    }
    
    // Handle new player connection
    // Example: Initialize player state, assign player ID
    log.Printf("Player connected: %s", playerID)
}
```

### Disconnect Callback

Handles player disconnections:

```go
func handlePlayerDisconnect(playerID string) {
    game, ok := p8.CurrentCartridge().(*Game)
    if !ok || !game.isServer {
        return
    }
    
    // Handle player disconnection
    // Example: Remove player from game
    log.Printf("Player disconnected: %s", playerID)
}
```

## Message Types and Data Structures

### Game State Structure

Define a structure for your game state:

```go
type GameState struct {
    // Game state variables that need to be synchronized
    PlayerPositions map[string]Position
    GameObjects     []GameObject
    GameTime        float64
}

type Position struct {
    X float64
    Y float64
}

type GameObject struct {
    ID       int
    Position Position
    Type     int
    State    int
}
```

### Player Input Structure

Define a structure for player input:

```go
type PlayerInput struct {
    // Player input variables
    Up     bool
    Down   bool
    Left   bool
    Right  bool
    A      bool
    B      bool
    Start  bool
    Select bool
}
```

### Sending Game State

Implement a function to send game state from server to clients:

```go
func (g *Game) sendGameState() {
    if !g.isServer {
        return
    }
    
    // Create game state object
    state := GameState{
        PlayerPositions: make(map[string]Position),
        GameObjects:     make([]GameObject, 0),
        GameTime:        g.gameTime,
    }
    
    // Fill with current game state
    for id, player := range g.players {
        state.PlayerPositions[id] = Position{X: player.x, Y: player.y}
    }
    
    for i, obj := range g.gameObjects {
        state.GameObjects = append(state.GameObjects, GameObject{
            ID:       i,
            Position: Position{X: obj.x, Y: obj.y},
            Type:     obj.type,
            State:    obj.state,
        })
    }
    
    // Serialize and send
    data, err := json.Marshal(state)
    if err != nil {
        log.Printf("Error marshaling game state: %v", err)
        return
    }
    
    p8.SendMessage(p8.MsgGameState, "all", data)
}
```

### Sending Player Input

Implement a function to send player input from client to server:

```go
func (g *Game) sendPlayerInput() {
    if !g.isClient {
        return
    }
    
    // Create input object based on current button states
    input := PlayerInput{
        Up:     p8.Btn(p8.UP),
        Down:   p8.Btn(p8.DOWN),
        Left:   p8.Btn(p8.LEFT),
        Right:  p8.Btn(p8.RIGHT),
        A:      p8.Btn(p8.A),
        B:      p8.Btn(p8.B),
        Start:  p8.Btn(p8.START),
        Select: p8.Btn(p8.SELECT),
    }
    
    // Serialize and send
    data, err := json.Marshal(input)
    if err != nil {
        log.Printf("Error marshaling player input: %v", err)
        return
    }
    
    p8.SendMessage(p8.MsgPlayerInput, "", data)
}
```

## Client-Side Prediction

Client-side prediction improves the feel of multiplayer games by immediately showing the results of player input locally, then reconciling with the server's authoritative state.

### Implementing Prediction

```go
// In client's Update() function
if g.isClient {
    // Send input to server
    g.sendPlayerInput()
    
    // Store current position
    originalX := g.localPlayer.x
    originalY := g.localPlayer.y
    
    // Apply input locally for immediate feedback
    if p8.Btn(p8.LEFT) {
        g.localPlayer.x -= g.playerSpeed
    }
    if p8.Btn(p8.RIGHT) {
        g.localPlayer.x += g.playerSpeed
    }
    if p8.Btn(p8.UP) {
        g.localPlayer.y -= g.playerSpeed
    }
    if p8.Btn(p8.DOWN) {
        g.localPlayer.y += g.playerSpeed
    }
    
    // Store prediction for reconciliation
    g.predictions = append(g.predictions, Prediction{
        Time:      g.gameTime,
        OriginalX: originalX,
        OriginalY: originalY,
        PredictedX: g.localPlayer.x,
        PredictedY: g.localPlayer.y,
    })
}
```

### Reconciliation with Server State

```go
// In handleGameState function
if game.isClient {
    // Get server's position for local player
    serverPos, ok := state.PlayerPositions[p8.GetPlayerID()]
    if !ok {
        return
    }
    
    // Calculate difference between prediction and server state
    diffX := math.Abs(game.localPlayer.x - serverPos.X)
    diffY := math.Abs(game.localPlayer.y - serverPos.Y)
    
    // If significant difference, smoothly correct
    if diffX > 5 || diffY > 5 {
        // Smooth interpolation
        game.localPlayer.x = game.localPlayer.x + (serverPos.X - game.localPlayer.x) * 0.3
        game.localPlayer.y = game.localPlayer.y + (serverPos.Y - game.localPlayer.y) * 0.3
    }
    
    // Update other players' positions directly
    for id, pos := range state.PlayerPositions {
        if id != p8.GetPlayerID() {
            if player, ok := game.players[id]; ok {
                player.x = pos.X
                player.y = pos.Y
                game.players[id] = player
            } else {
                // Create new player if not exists
                game.players[id] = Player{x: pos.X, y: pos.Y}
            }
        }
    }
}
```

## Case Study: Multiplayer Gameboy

Let's convert the Gameboy example to a multiplayer game where two players can move around the screen.

### Step 1: Define Game Structures

```go
type Player struct {
    x      float64
    y      float64
    sprite int
    speed  float64
}

type Game struct {
    // Game world
    map        [][]int
    tileSize   float64
    
    // Players
    players    map[string]Player
    localPlayer Player
    
    // Network
    isServer        bool
    isClient        bool
    lastStateUpdate time.Time
}
```

### Step 2: Define Network Messages

```go
type GameState struct {
    Players map[string]PlayerState
}

type PlayerState struct {
    X      float64
    Y      float64
    Sprite int
}

type PlayerInput struct {
    Up     bool
    Down   bool
    Left   bool
    Right  bool
    A      bool
    B      bool
}
```

### Step 3: Initialize Game with Network Roles

```go
func (g *Game) Init() {
    // Initialize map and game world
    g.map = loadMap()
    g.tileSize = 8
    
    // Initialize players
    g.players = make(map[string]Player)
    
    // Set up network roles
    g.isServer = p8.GetNetworkRole() == p8.RoleServer
    g.isClient = p8.GetNetworkRole() == p8.RoleClient
    g.lastStateUpdate = time.Now()
    
    // Initialize local player
    g.localPlayer = Player{
        x:      64,
        y:      64,
        sprite: 1,
        speed:  1.0,
    }
    
    // Add local player to players map
    playerID := "server"
    if g.isClient {
        playerID = p8.GetPlayerID()
    }
    g.players[playerID] = g.localPlayer
}
```

### Step 4: Implement Network Callbacks

```go
func handleGameState(playerID string, data []byte) {
    game, ok := p8.CurrentCartridge().(*Game)
    if !ok {
        return
    }
    
    var state GameState
    if err := json.Unmarshal(data, &state); err != nil {
        log.Printf("Error unmarshaling game state: %v", err)
        return
    }
    
    // Update all players except local player
    for id, playerState := range state.Players {
        if game.isClient && id == p8.GetPlayerID() {
            // For local player, apply reconciliation
            diffX := math.Abs(game.localPlayer.x - playerState.X)
            diffY := math.Abs(game.localPlayer.y - playerState.Y)
            
            if diffX > 5 || diffY > 5 {
                // Smooth interpolation
                game.localPlayer.x = game.localPlayer.x + (playerState.X - game.localPlayer.x) * 0.3
                game.localPlayer.y = game.localPlayer.y + (playerState.Y - game.localPlayer.y) * 0.3
                
                // Update players map
                game.players[id] = game.localPlayer
            }
        } else {
            // For remote players, update directly
            game.players[id] = Player{
                x:      playerState.X,
                y:      playerState.Y,
                sprite: playerState.Sprite,
                speed:  1.0,
            }
        }
    }
}

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
    
    // Get player or create if not exists
    player, ok := game.players[playerID]
    if !ok {
        player = Player{
            x:      64,
            y:      64,
            sprite: 2, // Different sprite for client
            speed:  1.0,
        }
    }
    
    // Update player based on input
    if input.Left {
        player.x -= player.speed
    }
    if input.Right {
        player.x += player.speed
    }
    if input.Up {
        player.y -= player.speed
    }
    if input.Down {
        player.y += player.speed
    }
    
    // Apply collision detection
    player = game.applyCollision(player)
    
    // Update player in map
    game.players[playerID] = player
}

func handlePlayerConnect(playerID string) {
    game, ok := p8.CurrentCartridge().(*Game)
    if !ok || !game.isServer {
        return
    }
    
    // Initialize new player
    game.players[playerID] = Player{
        x:      64,
        y:      64,
        sprite: 2, // Different sprite for client
        speed:  1.0,
    }
    
    log.Printf("Player connected: %s", playerID)
}

func handlePlayerDisconnect(playerID string) {
    game, ok := p8.CurrentCartridge().(*Game)
    if !ok || !game.isServer {
        return
    }
    
    // Remove player
    delete(game.players, playerID)
    
    log.Printf("Player disconnected: %s", playerID)
}
```

### Step 5: Implement Game Update Logic

```go
func (g *Game) Update() {
    // Handle waiting for players state
    if p8.IsWaitingForPlayers() {
        return
    }
    
    // Update local player based on input
    if g.isServer {
        // Server controls its local player
        originalX := g.localPlayer.x
        originalY := g.localPlayer.y
        
        if p8.Btn(p8.LEFT) {
            g.localPlayer.x -= g.localPlayer.speed
        }
        if p8.Btn(p8.RIGHT) {
            g.localPlayer.x += g.localPlayer.speed
        }
        if p8.Btn(p8.UP) {
            g.localPlayer.y -= g.localPlayer.speed
        }
        if p8.Btn(p8.DOWN) {
            g.localPlayer.y += g.localPlayer.speed
        }
        
        // Apply collision detection
        g.localPlayer = g.applyCollision(g.localPlayer)
        
        // Update player in map
        g.players["server"] = g.localPlayer
        
        // Send game state to clients
        if time.Since(g.lastStateUpdate) > 16*time.Millisecond {
            g.sendGameState()
            g.lastStateUpdate = time.Now()
        }
    } else if g.isClient {
        // Client controls its local player with prediction
        originalX := g.localPlayer.x
        originalY := g.localPlayer.y
        
        if p8.Btn(p8.LEFT) {
            g.localPlayer.x -= g.localPlayer.speed
        }
        if p8.Btn(p8.RIGHT) {
            g.localPlayer.x += g.localPlayer.speed
        }
        if p8.Btn(p8.UP) {
            g.localPlayer.y -= g.localPlayer.speed
        }
        if p8.Btn(p8.DOWN) {
            g.localPlayer.y += g.localPlayer.speed
        }
        
        // Apply collision detection
        g.localPlayer = g.applyCollision(g.localPlayer)
        
        // Update player in map
        g.players[p8.GetPlayerID()] = g.localPlayer
        
        // Send input to server
        g.sendPlayerInput()
    }
}
```

### Step 6: Implement Send Functions

```go
func (g *Game) sendGameState() {
    if !g.isServer {
        return
    }
    
    // Create game state object
    state := GameState{
        Players: make(map[string]PlayerState),
    }
    
    // Fill with current player states
    for id, player := range g.players {
        state.Players[id] = PlayerState{
            X:      player.x,
            Y:      player.y,
            Sprite: player.sprite,
        }
    }
    
    // Serialize and send
    data, err := json.Marshal(state)
    if err != nil {
        log.Printf("Error marshaling game state: %v", err)
        return
    }
    
    p8.SendMessage(p8.MsgGameState, "all", data)
}

func (g *Game) sendPlayerInput() {
    if !g.isClient {
        return
    }
    
    // Create input object based on current button states
    input := PlayerInput{
        Left:   p8.Btn(p8.LEFT),
        Right:  p8.Btn(p8.RIGHT),
        Up:     p8.Btn(p8.UP),
        Down:   p8.Btn(p8.DOWN),
        A:      p8.Btn(p8.A),
        B:      p8.Btn(p8.B),
    }
    
    // Serialize and send
    data, err := json.Marshal(input)
    if err != nil {
        log.Printf("Error marshaling player input: %v", err)
        return
    }
    
    p8.SendMessage(p8.MsgPlayerInput, "", data)
}
```

### Step 7: Implement Draw Function

```go
func (g *Game) Draw() {
    p8.Cls(0)
    
    // Display network status
    if p8.IsWaitingForPlayers() || p8.GetNetworkError() != "" {
        p8.DrawNetworkStatus(10, 10, 7)
        return
    }
    
    // Draw map
    for y := 0; y < len(g.map); y++ {
        for x := 0; x < len(g.map[y]); x++ {
            tileType := g.map[y][x]
            p8.Spr(tileType, float64(x)*g.tileSize, float64(y)*g.tileSize)
        }
    }
    
    // Draw all players
    for id, player := range g.players {
        p8.Spr(player.sprite, player.x, player.y)
        
        // Draw player ID above sprite
        p8.Print(id, player.x, player.y-8, 7)
    }
    
    // Draw network role
    if g.isServer {
        p8.Print("SERVER", 2, 2, 8)
    } else if g.isClient {
        p8.Print("CLIENT", 2, 2, 12)
    }
}
```

### Step 8: Implement Collision Detection

```go
func (g *Game) applyCollision(player Player) Player {
    // Get tile coordinates
    tileX := int(player.x / g.tileSize)
    tileY := int(player.y / g.tileSize)
    
    // Check surrounding tiles
    for y := tileY - 1; y <= tileY + 1; y++ {
        for x := tileX - 1; x <= tileX + 1; x++ {
            // Check map bounds
            if y >= 0 && y < len(g.map) && x >= 0 && x < len(g.map[y]) {
                tileType := g.map[y][x]
                
                // Check if tile is solid (e.g., tile type 1 is solid)
                if tileType == 1 {
                    // Simple collision detection
                    tileLeft := float64(x) * g.tileSize
                    tileRight := tileLeft + g.tileSize
                    tileTop := float64(y) * g.tileSize
                    tileBottom := tileTop + g.tileSize
                    
                    playerLeft := player.x
                    playerRight := player.x + g.tileSize
                    playerTop := player.y
                    playerBottom := player.y + g.tileSize
                    
                    // Check for collision
                    if playerRight > tileLeft && playerLeft < tileRight &&
                       playerBottom > tileTop && playerTop < tileBottom {
                        // Resolve collision
                        overlapX := math.Min(playerRight - tileLeft, tileRight - playerLeft)
                        overlapY := math.Min(playerBottom - tileTop, tileBottom - playerTop)
                        
                        if overlapX < overlapY {
                            if playerLeft < tileLeft {
                                player.x -= overlapX
                            } else {
                                player.x += overlapX
                            }
                        } else {
                            if playerTop < tileTop {
                                player.y -= overlapY
                            } else {
                                player.y += overlapY
                            }
                        }
                    }
                }
            }
        }
    }
    
    return player
}
```

### Step 9: Set Up Main Function

```go
func main() {
    // Create and initialize the game
    game := &Game{}
    p8.InsertGame(game)
    game.Init()
    
    // Register network callbacks
    p8.SetOnGameStateCallback(handleGameState)
    p8.SetOnPlayerInputCallback(handlePlayerInput)
    p8.SetOnConnectCallback(handlePlayerConnect)
    p8.SetOnDisconnectCallback(handlePlayerDisconnect)
    
    // Start the game
    settings := p8.NewSettings()
    settings.WindowTitle = "PIGO8 Multiplayer Gameboy"
    p8.PlayGameWith(settings)
}
```

## Advanced Topics

### Bandwidth Optimization

To reduce bandwidth usage:

1. **Send Only What Changed**: Only include changed values in game state
2. **Compression**: Compress data before sending
3. **Update Frequency**: Adjust update frequency based on game needs
4. **Delta Encoding**: Send only differences from previous state

### Handling Latency

Strategies for handling network latency:

1. **Client-Side Prediction**: Predict movement locally
2. **Server Reconciliation**: Correct client predictions
3. **Entity Interpolation**: Smooth movement of remote entities
4. **Input Buffering**: Buffer inputs to handle jitter

### Synchronization Strategies

Different approaches to game synchronization:

1. **Lockstep**: All clients wait for all inputs before advancing
2. **Snapshot Interpolation**: Interpolate between received snapshots
3. **State Synchronization**: Server sends authoritative state
4. **Event-Based**: Synchronize via events rather than full state

## Troubleshooting

### Common Issues

1. **Jittery Movement**: Implement client-side prediction and interpolation
2. **Desynchronization**: Ensure server is authoritative for game logic
3. **High Latency**: Optimize message size and frequency
4. **Connection Issues**: Check network configuration and firewalls

### Debugging Tools

1. **Logging**: Add detailed logging for network events
2. **Visualization**: Visualize network state and predictions
3. **Artificial Latency**: Test with artificial latency
4. **Packet Inspection**: Analyze packet contents and timing

### Best Practices

1. **Keep It Simple**: Start with minimal networking and add complexity as needed
2. **Test Early and Often**: Test multiplayer functionality throughout development
3. **Graceful Degradation**: Handle network issues gracefully
4. **Security**: Validate all inputs on the server

---

By following this guide, you should be able to convert any PIGO8 game to multiplayer, including the Gameboy example. The key is to identify what needs to be synchronized, implement proper client-server communication, and add client-side prediction for a smooth player experience.
