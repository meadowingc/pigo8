#  Building for Other Platforms (Cross-Compilation) and WebAssembly

Once you’ve confirmed your game runs locally, Go’s built-in cross-compilation makes it trivial to produce binaries for Windows, macOS, Linux—even for embedded ARM devices—and even WebAssembly for the browser. You don’t need any extra toolchain setup beyond Go itself.

## Cross-compiling for Native Architectures

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
$ go tool dist list
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

By the way, since PIGO8 uses Ebiten, you can also build against [Nintendo Switch](https://ebitengine.org/en/blog/nintendo_switch.html) if you like.

## Building for WebAssembly

Go can compile to WebAssembly, letting you embed your pigo8 game in a webpage. Here’s how:

First, please set target to `JavaScript/WASM`:

```bash
$ GOOS=js   GOARCH=wasm go build -o main.wasm
```

Then copy the Go runtime support file. The Go distribution includes a small JavaScript shim (`wasm_exec.js`) that initializes the WebAssembly module. You can find it in your Go root:

```bash
$ cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
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