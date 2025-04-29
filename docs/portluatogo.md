# Port Lua to Go code

Ok so let's start the usual Go procedure by initializing Go Modules.

```bash
$ go mod init github.com/yourname/myGame
$ go mod tidy
```

Now you’re set up: we have Go installed, a project folder, and the PICO-8 cart and sprite image in place. Next, let’s review the Lua code we want to port.

The [NerdyTeachers “Animate Multiple Sprites”](https://nerdyteachers.com/PICO-8/Game_Mechanics/4) tutorial uses Lua tables and loops to animate a player, some enemies, and items. Let’s highlight the key parts:

## Lua code analysis

Let us study the original Lua code written for PICO-8, before try to port it to Go and PIGO8.
We need to understand it. 

###  Variables and tables

In PICO-8 Lua, global tables hold object data. For example, in `_init()` they create:

```lua
player = { sprite=1, x=-8, y=59, timing=0.25 }
enemies = {}
enemy1 = { sprite=5, x=-20, y=5, timing=0.1, speed=1.25, first=5, last=9 }
add(enemies, enemy1)
-- (and similarly enemy2, enemy3, items, etc)
```

Here each table has fields like `sprite`, `x`, `y`, `timing`, and (for enemies/items) `first`, `last`, `speed`.
The `player` _table_ holds its **current frame number** and **position​**.

### Animation timing

The key trick in the tutorial is that each object’s `sprite` field is **a number** (not necessarily integer). Each update, they do:

```lua
object.sprite += object.timing
if object.sprite >= object.last then
    object.sprite = object.first
end
```

This floats `sprite` by a small increment so that frames advance more slowly than every tick. Because PICO-8 rounds down when drawing, a sprite index of 1.25 still draws sprite 1 until it reaches 2. This lets them animate at a fraction of the frame rate.

### Movement

Each enemy moves horizontally by `enemy.speed` (or `player.x += 1` for the player), and _when_ `x > 127`, it _resets_ to `-8` to wrap around​. In the code:

```lua
x += 1
if x > 127 then x = -8 end
```

The tutorials explains that the screen is 128 pixels wide (`0`–`127`), so setting `x = -8` places the sprite just **off-screen** on the left, giving a *smooth* wrap.

### A simplified game loop

Putting it together, the full Lua update code looks like this (single-object version for simplicity):

```lua
function _update()
    -- animate
    sprite += timing
    if sprite >= 5 then sprite = 1 end

    -- move
    x += 1
    if x > 127 then x = -8 end
end
```

This updates the sprite index and position each tick. For multiple objects, they repeat similar blocks inside loops.

The `_draw()` function simply loops through all objects and calls `spr()` on each.

We’ll mirror each of these concepts in Go.

## Translate concepts to Go

Now we port these ideas into Go. In Go we’ll define a struct to represent an animated object, write methods for animation and movement, and set up update/draw loops. Unlike Lua’s flexible tables, Go has static typing: every field has a declared type. We’ll use `float64` for everything so we don't bother type-casting. Here’s a basic struct:

```go
// Entity represents an animated object (player, enemy, or item).
type Entity struct {
    Sprite float64   // current sprite index (can be fractional for timing)
    X, Y   float64   // position on screen
    Timing float64   // how much to advance per frame
    Speed  float64   // horizontal movement speed (0 for static items)
    First  float64       // first sprite index in animation loop
    Last   float64       // one past the last sprite index in animation loop
}
```

Notice the fields correspond to the Lua table keys. 

For example, `player = {sprite=1, x=-8, y=59, timing=0.25}` becomes something like `Entity{Sprite:1, X:-8, Y:59, Timing:0.25, First:1, Last:5}`. 

We include `First` and `Last` so each entity knows its animation range (for the player in the tutorial, `first=1` and `last=5` since sprites `1`–`4` are used). We’ll write a [Factory constructor](https://refactoring.guru/design-patterns/factory-method/go/example) function to create these easily:

```go
// NewEntity creates a new AnimatedEntity.
func NewEntity(sprite, x, y, timing, speed, first, last float64) Entity {
	return Entity{
		// Animation properties
		sprite: sprite,
		timing: timing,
		first:  first,
		last:   last,

		// Movement properties
		x:     x,
		y:     y,
		speed: speed,
	}
}
```

This mirrors the `Lua enemy1 = { sprite=5, x=-20, y=5, timing=0.1, speed=1.25, first=5, last=9 }`.

We have to pass numeric arguments in the correct order; Go’s strictness means we can’t omit fields like you can in Lua. Using a constructor helps avoid mistakes.

Next, we’ll give `Entity` two methods:

1. `Animate()`
2. `Move()`.

These will update the _sprite index_ and _position_, similar to the Lua `_update` logic:

```go
// Animate updates the sprite based on the timing and resets it within its cycle.
// Requires first and last values for each entity.
func (ae *Entity) Animate() {
	ae.sprite += ae.timing
	if ae.sprite >= ae.last {
		ae.sprite = ae.first
	}
}

// Move updates the entity's x-coordinate using the provided offset.
// It wraps the position around if it exceeds the right boundary (128).
func (ae *Entity) Move(offset float64) {
	ae.x += offset
	if ae.x > 128 {
		ae.x = -8
	}
}
```

With our `Entity` defined, let’s build the game. We can create slices (dynamic arrays) to hold enemies and items:

```go
var player Entity
var enemies = []Entity{}
var items = []Entity{}
```

In the tutorial’s `_init()`, they set up each enemy and then use `add(enemies, enemy)`.
In Go we’ll do something like:

```go
func (m *myGame) Init() {
	player = NewEntity(1, -8, 59, 0.25, 1, 1, 5)

	enemy1 := NewEntity(5, -20, 5, 0.1, 1.25, 5, 9)
	enemy2 := NewEntity(9, -14, 30, 0.2, 0.4, 9, 13)
	enemy3 := NewEntity(13, -11, 90, 0.4, 0.75, 13, 17)
	enemies = append(enemies, enemy1, enemy2, enemy3)

	item1 := NewEntity(48, 30, 110, 0.3, 48, 50, 56)
	item2 := NewEntity(56, 60, 110, 0.25, 54, 56, 60)
	item3 := NewEntity(60, 90, 110, 0.15, 4, 60, 64)
	items = append(items, item1, item2, item3)
}
```
Here we’re mimicking the Lua tables from the tutorial​, just using Go syntax.

Note how we pack each enemy and item into Go slices; this replaces `Lua’s add(enemies, enemy1)` and the `for ... in all(enemies) logic​.` In Go, to loop over a slice we will later write `for _, enemy := range g.Enemies { ... }`.

### Building the Update and Draw Loop

```go
func (m *myGame) Update() {
	// Update player: animate and move (player moves by 1 unit per frame)
	player.Animate()
	player.Move(player.speed)

	// Update enemies: animate and move based on each entity's speed
	for i := range enemies {
		enemies[i].Animate()
		enemies[i].Move(enemies[i].speed)
	}

	// Update items: animate only, don't move
	for i := range items {
		items[i].Animate()
	}
}

func (m *myGame) Draw() {
	p8.Cls(0) // clear screen
    
	player.Draw() // Draw the player

    // Draw all enemies
	for _, enemy := range enemies {
		enemy.Draw()
	}

    // // Draw all items
	for _, item := range items {
		item.Draw()
	}
}
```

In these snippets, we call a hypothetical `pigo8.Spr(index, x, y)` function (mirroring PICO-8’s `spr()`) and `pigo8.Cls()` to clear the screen. The logic is the same as the Lua `_draw()`: draw each object’s current frame at its position​.

Notice how we converted the Lua loops into Go for loops. For instance, the Lua code:

```lua
for enemy in all(enemies) do
    spr(enemy.sprite, enemy.x, enemy.y)
end
```
becomes Go:

```go
for _, enemy := range g.Enemies {
    pigo8.Spr(enemy.Sprite, enemy.X, enemy.Y)
}
```

We use `range` to iterate over the `slice`.

## Full Go Program

```go
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type Entity struct {
	sprite, x, y, timing, speed, first, last float64
}

func NewEntity(sprite, x, y, timing, speed, first, last float64) Entity {
	return Entity{
		sprite: sprite,
		timing: timing,
		first:  first,
		last:   last,
		x:      x,
		y:      y,
		speed:  speed,
	}
}

func (ae *Entity) Animate() {
	ae.sprite += ae.timing
	if ae.sprite >= ae.last {
		ae.sprite = ae.first
	}
}

func (ae *Entity) Move(offset float64) {
	ae.x += offset
	if ae.x > 128 {
		ae.x = -8
	}
}

func (ae *Entity) Draw() {
	p8.Spr(ae.sprite, ae.x, ae.y)
}

var player Entity
var enemies = []Entity{}
var items = []Entity{}

type myGame struct{}

func (m *myGame) Init() {
	player = NewEntity(1, -8, 59, 0.25, 1, 1, 5)
	enemy1 := NewEntity(5, -20, 5, 0.1, 1.25, 5, 9)
	enemy2 := NewEntity(9, -14, 30, 0.2, 0.4, 9, 13)
	enemy3 := NewEntity(13, -11, 90, 0.4, 0.75, 13, 17)
	enemies = append(enemies, enemy1, enemy2, enemy3)
	item1 := NewEntity(48, 30, 110, 0.3, 48, 50, 56)
	item2 := NewEntity(56, 60, 110, 0.25, 54, 56, 60)
	item3 := NewEntity(60, 90, 110, 0.15, 4, 60, 64)
	items = append(items, item1, item2, item3)
}

func (m *myGame) Update() {
	player.Animate()
	player.Move(player.speed)

	for i := range enemies {
		enemies[i].Animate()
		enemies[i].Move(enemies[i].speed)
	}

	for i := range items {
		items[i].Animate()
	}
}

func (m *myGame) Draw() {
	p8.Cls(0)
	player.Draw()

	for _, enemy := range enemies {
		enemy.Draw()
	}

	for _, item := range items {
		item.Draw()
	}
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
```

To try the game, use the Go tools. In your project directory, run:

```bash
go run .
```

This compiles and runs the `main.g`o program (the `.` means run the current module).
You should see a window or output with your animated sprites moving, just like in the PICO-8 demo.

To build a standalone executable, use:

```bash
go build -o mygame
```

This produces a binary named `mygame` (or `mygame.exe` on Windows).
You can then run `./mygame` anytime to play your game.

##  Building for Other Platforms (Cross-Compilation) and WebAssembly

Once you’ve confirmed your game runs locally, Go’s built-in cross-compilation makes it trivial to produce binaries for Windows, macOS, Linux—even for embedded ARM devices—and even WebAssembly for the browser. You don’t need any extra toolchain setup beyond Go itself.

### Cross-compiling for Native Architectures

Go uses two environment variables to target a specific OS and CPU architecture:

* `GOOS`: target operating system (`linux`, `windows`, `darwin` for macOS, etc.)

* `GOARCH`: target CPU architecture (`amd64`, `386`, `arm64`, `arm`, etc.)

From your project folder, simply run:

```bash
# Linux on AMD64
GOOS=linux   GOARCH=amd64 go build -o mygame-linux-amd64

# Windows on 386 (32-bit)
GOOS=windows GOARCH=386  go build -o mygame-windows-386.exe

# macOS on ARM64 (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o mygame-darwin-arm64

# Linux on ARM (e.g. Raspberry Pi)
GOOS=linux   GOARCH=arm   go build -o mygame-linux-arm
```

You can mix and match any supported `GOOS/GOARCH`. The output binary name (`-o`) is up to you.
No extra downloads: Go’s standard distribution already contains everything needed.

To see all valid pairs, run:

```bash
go tool dist list
```

So you PIGO8 can run in the following computers:

```
aix/ppc64
android/386
android/amd64
android/arm
android/arm64
darwin/amd64
darwin/arm64
dragonfly/amd64
freebsd/386
freebsd/amd64
freebsd/arm
freebsd/arm64
freebsd/riscv64
illumos/amd64
ios/amd64
ios/arm64
js/wasm
linux/386
linux/amd64
linux/arm
linux/arm64
linux/loong64
linux/mips
linux/mips64
linux/mips64le
linux/mipsle
linux/ppc64
linux/ppc64le
linux/riscv64
linux/s390x
netbsd/386
netbsd/amd64
netbsd/arm
netbsd/arm64
openbsd/386
openbsd/amd64
openbsd/arm
openbsd/arm64
openbsd/ppc64
openbsd/riscv64
plan9/386
plan9/amd64
plan9/arm
solaris/amd64
wasip1/wasm
windows/386
windows/amd64
windows/arm64
```

### Building for WebAssembly

Go can compile to WebAssembly, letting you embed your pigo8 game in a webpage. Here’s how:

First, please set target to `JavaScript/WASM`:

```bash
GOOS=js   GOARCH=wasm go build -o main.wasm
```

Then copy the Go runtime support file. The Go distribution includes a small JavaScript shim (`wasm_exec.js`) that initializes the WebAssembly module. You can find it in your Go root:

```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

Ok, now let's make it loading into a web browser. So we need to create a simple HTML.
Save this as `index.html` alongside `main.wasm` and `wasm_exec.js` (all three of them in the same folder):

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>My pigo8 Game</title>
  <script src="wasm_exec.js"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
      .then((res) => go.run(res.instance))
      .catch(console.error);
  </script>
</head>
<body>
  <canvas id="ebiten_canvas"></canvas>
</body>
</html>
```

Now you can upload these to a web-server, or GitHub's actions for your repository and have people play your game.
You can test it locally as well, You can use any static file server, for example Python’s:

```bash
$ python3 -m http.server 8080
```

Then open `http://localhost:8080` in your browser.

With just these environment variables and a tiny HTML wrapper, you can ship your pigo8-powered game to nearly any platform—including the web—without touching C toolchains or external build systems.

Enjoy spreading your PICO-8 magic far and wide!