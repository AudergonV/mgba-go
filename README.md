# MGBA Go binding

[![Build](https://github.com/AudergonV/mgba-go/actions/workflows/build.yaml/badge.svg)](https://github.com/AudergonV/mgba-go/actions/workflows/build.yaml)
[![Test](https://github.com/AudergonV/mgba-go/actions/workflows/test.yaml/badge.svg)](https://github.com/AudergonV/mgba-go/actions/workflows/test.yaml)

A Go binding for [mGBA](https://mgba.io/), the Game Boy Advance emulator library. Built on top of **libmgba** via CGo, it exposes ROM loading, frame-by-frame emulation, and raw framebuffer access.

> **Status:** experimental — API may change.

---

## Requirements

- Go 1.21+
- `libmgba` (C library + headers)

**macOS (Homebrew)**
```sh
brew install mgba
```

**Ubuntu / Debian**
```sh
sudo apt-get install libmgba-dev
```

---

## Installation

```sh
go get github.com/audergonv/mgba-go
```

CGo must be able to find the libmgba headers and shared library. On macOS with Homebrew the paths are picked up automatically. On Linux the package manager install is sufficient.

---

## Quick start

```go
package main

import (
    "image"
    "image/png"
    "os"
    "time"

    "github.com/audergonv/mgba-go/pkg/mgba"
)

func main() {
    // 1. Open the ROM file.
    vf, err := mgba.VFileFromPath("game.gba")
    if err != nil {
        panic(err)
    }
    defer vf.Close()

    // 2. Detect the core type (GBA / GB / GBC) and create it.
    core, err := mgba.NewCore(vf)
    if err != nil {
        panic(err)
    }

    // 3. Initialise: allocates the internal framebuffer and wires up the
    //    software renderer.  Must be called before LoadROM or Reset.
    if err := core.Init(); err != nil {
        panic(err)
    }
    defer core.Deinit()

    // 4. Load the ROM. The core takes ownership of the VFile from this point.
    if err := core.LoadROM(vf); err != nil {
        panic(err)
    }

    // 5. Reset the hardware to its power-on state.
    core.Reset()

    // 6. Get a slice that always reflects the latest rendered frame.
    //    Layout: BGRA (little-endian 0x00RRGGBB), 4 bytes per pixel.
    width, height, _ := core.BaseVideoSize()
    buf, _ := core.VideoBuffer()

    // 7. Run frames at ~59.73 fps (GBA native rate).
    const frameDuration = time.Duration(16742706) // ns
    deadline := time.Now().Add(frameDuration)

    for frame := 0; frame < 300; frame++ {
        if err := core.RunFrame(); err != nil {
            panic(err)
        }

        // buf is updated in-place after every RunFrame call.
        _ = buf // use it here — render to terminal, encode as PNG, etc.

        if now := time.Now(); now.Before(deadline) {
            time.Sleep(deadline.Sub(now))
        }
        deadline = deadline.Add(frameDuration)
    }

    // Example: save the last frame as a PNG.
    img := image.NewRGBA(image.Rect(0, 0, width, height))
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            base := (y*width + x) * 4
            // mGBA byte order: [B, G, R, 0]
            img.Pix[(y*width+x)*4+0] = buf[base+2] // R
            img.Pix[(y*width+x)*4+1] = buf[base+1] // G
            img.Pix[(y*width+x)*4+2] = buf[base+0] // B
            img.Pix[(y*width+x)*4+3] = 0xFF         // A
        }
    }
    f, _ := os.Create("frame.png")
    defer f.Close()
    png.Encode(f, img)
}
```

---

## API overview

| Symbol | Description |
|---|---|
| `VFileFromPath(path)` | Open a ROM from the filesystem |
| `VFileFromBytes(data)` | Open a ROM from an in-memory byte slice |
| `NewCore(vf)` | Detect and create the appropriate emulator core |
| `(*GBACore).Init()` | Allocate framebuffer and configure the core |
| `(*GBACore).LoadROM(vf)` | Load the ROM; the core takes ownership of the VFile |
| `(*GBACore).Reset()` | Power-cycle the emulated hardware |
| `(*GBACore).RunFrame()` | Emulate exactly one frame |
| `(*GBACore).VideoBuffer()` | Raw `[]byte` slice of the current framebuffer |
| `(*GBACore).BaseVideoSize()` | Native resolution (240×160 for GBA) |
| `(*GBACore).Deinit()` | Release all C resources |

### Pixel format

`VideoBuffer()` returns a flat byte slice in **little-endian `0x00RRGGBB`** order:

```
offset + 0 → B
offset + 1 → G
offset + 2 → R
offset + 3 → 0x00 (padding)
```

4 bytes per pixel, row-major, stride = width.

---

## ASCII demo

`cmd/demo` contains a terminal renderer that plays a GBA ROM as ASCII art at native speed.

```sh
# Build
make build-demo

# Run (place your ROM as game.gba in the repo root)
./demo
```
---

## Project structure

```
pkg/mgba/
  vfs.go      — VFile abstraction (file-backed and memory-backed)
  core.go     — GBACore: init, ROM loading, frame execution, framebuffer
  cgo.go      — CGo compiler/linker flags
cmd/demo/
  main.go     — ASCII art terminal renderer
```

---

## License

MPL-2.0 — see [LICENSE](LICENSE).  
mGBA is copyright Jeffrey Pfau, also MPL-2.0.

---

## Notice on LLM-generated content

Claude Sonnet 4.6 was used to assist in writing and editing this README.