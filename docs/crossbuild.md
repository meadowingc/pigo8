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
    <meta charset="utf-8" />
    <title>pigo8 Game</title>
    <meta
      name="viewport"
      content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no"
    />
    <style>
      html,
      body {
        margin: 0;
        padding: 0;
        background: #000;
        color: #ccc;
        touch-action: none;
        font-family: monospace;
        -webkit-user-select: none;
        user-select: none;
      }
      canvas {
        display: block;
        margin: 0 auto; /* horizontal center only */
        image-rendering: pixelated;
        position: absolute;
        top: 0;
      }
      #touch_layer {
        position: fixed;
        inset: 0;
        pointer-events: none;
        z-index: 20;
      }
      .panel {
        position: absolute;
        opacity: 0.55;
        image-rendering: pixelated;
        pointer-events: none;
      }
      @media (hover: hover) and (pointer: fine) {
        #touch_layer {
          display: none;
        }
      }
    </style>
    <script src="wasm_exec.js"></script>
    <script>
      /* ======== CONFIG ======== */
      const DEADZONE = 0.6; // directional gating slope ratio (like original 0.6 heuristic)
      const FUDGE = 0; // extra pixels expanding controller vertical trigger band (0..r)
      const DIR_DEAD_FRAC = 1 / 3; // deadzone radius fraction relative to r (same logic as original r/3 cap)
      /* ======================== */

      /* Bitfield consumed by Go engine */
      window.pigo8BtnBits = 0;
      function setBits(mask) {
        window.pigo8BtnBits = mask;
      }

      /* Base64 panel images (taken from PICO-8 html) */
      const P8_GFX = {
        left: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAASwAAAEsCAYAAAB5fY51AAAEI0lEQVR42u3dMU7DQBCG0Tjam9DTcP8jpEmfswS5iHBhAsLxev/hvQY6pGXyZRTQ+nQCAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAHqbHEEtl+vt7hS+fLy/mXHBQqxEi/6aI/AiFW9SnB2BWDkDBAtAsADBAhAsAMECBAtAsAAECxAsAMECECxAsAAEC0CwONJ8tYvrXRAsImK19j0IFsPGSrQQLCJiNV+et7xAT7QQLIaN1dr3ooVgMWysRAvBIipWooVgERUr0UKwiIqVaCFYRMVKtBAsomIlWggWUbESLQSLqFiJFoJFVKxEC8EiKlaihWARFSvRQrDYJSSVfhaCBSBYAIIFCBbAHpoj4Bl/scOGBWDD4lX8iwE2LADBAgQLQLAABAsQLADBAhAsQLAABAtAsADBAhAsAMECBAtAsAAECxAsAMECECxAsAAECxAsAMECECxAsMh1ud7uTsHZVDcZyFo8Yt5sVJ6NyUAaSNEyIymaXwZepIKd4mwoQbAFC0CwAMECECwAwQIEC0CwAAQLECwAwQIQLECwAAQLQLAAwQI4UHME2/10QZq7usyBObBhRQwpmBUb1nADuPbuaUD/p2ezMH+1admwhosVfBcxb2SCJVaIlmAhVoiWYIkVoiVagiVWiJZgiZVYIVqCJVaIlmgJllghWoIlViBagiVWiJZoCZZYIVqCJVYgWoIlViBaggUIlnc0sPELlmghVmIlWKKFWAmWaIFYCZZoIVYIlmghVoIlWiBWgiVaiJVgIVqIlWCJFoiVYIkWYiVYiBZiJViihViJ1XbNEWyL1mMQRYvfvIGJlQ1rmE0LzIoNyyBiDrBhAYIFIFiAYAEIFoBgAYIFIFgAggUIFoBgAQgWIFgAggUgWIBgDc+Nn1D/tdH8YupwgZy5qG4ykKIlVmZDsDjshSlazqQqH7p793Q2CBaAYAGCBSBYAIIFCBaAYAEIFiBYAIIFIFiAYAEIFoBgAYIFIFgAggUIFoBgAQgWIFgAggUgWIBgAQgWwENzBKxZPub9CJ7WjA0LsGFRV+9N5+jNDhsWgGABggUgWACCxW56fgjuA3cEiz9Z/nWwR0iWP8P/YCFYDBstsUKwiIiWWCFYRERLrBAsIqIlVggWEdESKwSLiGiJFYJFRLTECsEiIlpihWARES2xQrCIiJZYIVhEREusECwioiVWCBYx0RIrBIuoaIkVr+YhFHTZtMCGBQgWgGABCBYgWACCBSBYgGABCBaAYAGCBSBYAIIFCBbj2uOR8s6AEbhexgsWYri3SKhKczcXAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMA2n+e0UMDzh3yTAAAAAElFTkSuQmCC",
        right:
          "data:image/png;base64,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",
      };

      let leftPanel, rightPanel;

      function initPanels() {
        const layer = document.getElementById("touch_layer");
        leftPanel = document.createElement("img");
        rightPanel = document.createElement("img");
        leftPanel.className = rightPanel.className = "panel";
        leftPanel.src = P8_GFX.left;
        rightPanel.src = P8_GFX.right;
        layer.appendChild(leftPanel);
        layer.appendChild(rightPanel);

        layoutPanels();
        window.addEventListener("resize", layoutPanels);
      }

      function layoutPanels() {
        const w = window.innerWidth,
          h = window.innerHeight;
        let r = Math.min(w, h) / 12;
        if (r > 40) r = 40;
        // Left 6r x 6r at bottom-left
        leftPanel.style.left = "0px";
        leftPanel.style.top = h - r * 6 + "px";
        leftPanel.style.width = r * 6 + "px";
        leftPanel.style.height = r * 6 + "px";
        // Right 6r x 7r at bottom-right
        rightPanel.style.left = w - r * 6 + "px";
        rightPanel.style.top = h - r * 7 + "px";
        rightPanel.style.width = r * 6 + "px";
        rightPanel.style.height = r * 7 + "px";
      }

      /* Core touch -> bit conversion (mirrors original Pico-8 geometry) */
      function recomputeBits(touches) {
        const w = window.innerWidth,
          h = window.innerHeight;
        let r = Math.min(w, h) / 12;
        if (r > 40) r = 40;
        const r6 = r * 6;
        let mask = 0;
        for (let i = 0; i < touches.length; i++) {
          const t = touches[i];
          const x = t.clientX,
            y = t.clientY;
          if (y >= h - r * 8 - FUDGE) {
            // controller vertical band
            // MENU row (half-height above buttons)
            if (y < h - r * 6 && y > h - r * 8) {
              if (x > w - r * 3) mask |= 0x40;
            } else if (x < w / 2 && x < r6) {
              // D-Pad region
              const cx = r * 3;
              const cy = h - r * 3;
              const dx = x - cx;
              const dy = y - cy;
              const dead = r * DIR_DEAD_FRAC;
              const adx = Math.abs(dx),
                ady = Math.abs(dy);
              if (adx > ady * DEADZONE) {
                if (dx < -dead) mask |= 0x1;
                if (dx > dead) mask |= 0x2;
              }
              if (ady > adx * DEADZONE) {
                if (dy < -dead) mask |= 0x4;
                if (dy > dead) mask |= 0x8;
              }
            } else if (x > w - r6) {
              // Buttons region (diagonal split)
              if (h - y > (w - x) * 0.8) mask |= 0x10; // O
              if (w - x > (h - y) * 0.8) mask |= 0x20; // X
            }
          }
        }
        // Map Pico-8 mask bits to engine bitfield 0..6
        let out =
          (mask & 0x1 ? 1 : 0) |
          (mask & 0x2 ? 1 << 1 : 0) |
          (mask & 0x4 ? 1 << 2 : 0) |
          (mask & 0x8 ? 1 << 3 : 0) |
          (mask & 0x10 ? 1 << 4 : 0) |
          (mask & 0x20 ? 1 << 5 : 0) |
          (mask & 0x40 ? 1 << 6 : 0);
        setBits(out);
      }

      /* Unified touch handlers */
      function handleTouch(ev) {
        ev.preventDefault();
        recomputeBits(ev.touches);
      }
      /* debug helpers removed */

      window.addEventListener("load", () => {
        initPanels();

        /* --- Robust multi-touch state tracking to avoid "stuck" directional buttons --- */
        const activeTouches = new Map(); // id -> Touch
        function syncFromActive() {
          recomputeBits(Array.from(activeTouches.values()));
        }
        function onStart(ev) {
          ev.preventDefault();
          for (const t of ev.changedTouches) activeTouches.set(t.identifier, t);
          syncFromActive();
        }
        function onMove(ev) {
          ev.preventDefault();
          for (const t of ev.changedTouches) activeTouches.set(t.identifier, t);
          syncFromActive();
        }
        function onEnd(ev) {
          ev.preventDefault();
          for (const t of ev.changedTouches) activeTouches.delete(t.identifier);
          if (activeTouches.size === 0) {
            setBits(0);
          } else {
            syncFromActive();
          }
        }
        document.addEventListener("touchstart", onStart, {
          passive: false,
          capture: true,
        });
        document.addEventListener("touchmove", onMove, {
          passive: false,
          capture: true,
        });
        document.addEventListener("touchend", onEnd, {
          passive: false,
          capture: true,
        });
        document.addEventListener("touchcancel", onEnd, {
          passive: false,
          capture: true,
        });

        /* Safety: also clear on visibility change (e.g., user switches tabs mid-touch) */
        document.addEventListener("visibilitychange", () => {
          if (document.hidden) {
            activeTouches.clear();
            setBits(0);
          }
        });

        /* ======== WASM INIT ======== */
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
          .then((r) => go.run(r.instance))
          .catch((err) => console.error("WASM init error:", err));
        /* =========================== */
      });
    </script>
  </head>
  <body>
    <div id="touch_layer"></div>
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