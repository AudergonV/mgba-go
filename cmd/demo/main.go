// Author: Vincent Audergon <github.com/audergonv>
// License: MPL-2.0

package main

import (
	"fmt"
	"time"
	"os"

	"github.com/audergonv/mgba-go/pkg/mgba"
)

// RGBA8888ToASCII converts an RGBA8888 pixel to an ASCII character.
func RGBA8888ToASCII(r, g, b uint8) string {
	const ramp = `@#$%B&8WM*o+=~-:,. `
	avg := (int(r) + int(g) + int(b)) / 3
	idx := avg * (len(ramp) - 1) / 255
	return string(ramp[idx])
}

// PrintVideoBufferASCII prints the video buffer as ASCII art to the console. 
// It samples every 2x4 block of pixels and averages their colors to produce one character.
func PrintVideoBufferASCII(buffer []byte, width, height int) {
	for y := 0; y+3 < height; y += 4 {
		for x := 0; x+1 < width; x += 2 {
			i00 := (y*width+x) * 4
			i01 := (y*width+x+1) * 4
			i10 := ((y+1)*width+x) * 4
			i11 := ((y+1)*width+x+1) * 4
			b := uint8((int(buffer[i00]) + int(buffer[i01]) + int(buffer[i10]) + int(buffer[i11])) / 4)
			g := uint8((int(buffer[i00+1]) + int(buffer[i01+1]) + int(buffer[i10+1]) + int(buffer[i11+1])) / 4)
			r := uint8((int(buffer[i00+2]) + int(buffer[i01+2]) + int(buffer[i10+2]) + int(buffer[i11+2])) / 4)
			fmt.Print(RGBA8888ToASCII(r, g, b))
		}
		fmt.Println()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: demo <path-to-gba-rom>")
		return
	}
	vf, err := mgba.VFileFromPath(os.Args[1])
	if err != nil {
		fmt.Println("Error creating VFile:", err)
		return
	}
	defer vf.Close()
	fmt.Printf("VFile created: Type=%v\n", vf.Type)
	size, err := vf.Size()
	if err != nil {
		fmt.Println("Error getting VFile size:", err)
		return
	}
	fmt.Printf("VFile size: %d bytes\n", size)
	core, err := mgba.NewCore(vf)
	if err != nil {
		fmt.Println("Error creating GBACore:", err)
		return
	}
	err = core.Init()
	defer core.Deinit()
	if err != nil {
		fmt.Println("Error initializing GBACore:", err)
		return
	}
	width, height, err := core.BaseVideoSize()
	if err != nil {
		fmt.Println("Error getting base video size:", err)
		return
	}
	fmt.Printf("Base video size: %dx%d\n", width, height)
	err = core.LoadROM(vf)
	if err != nil {
		fmt.Println("Error loading file into GBACore:", err)
		return
	}
	core.Reset()
	fmt.Println("Core reset successfully")
	err = core.RunFrame()
	if err != nil {
		fmt.Println("Error running frame:", err)
		return
	}
	videobuffer, err := core.VideoBuffer()
	if err != nil {
		fmt.Println("Error getting video buffer:", err)
			return
	}
	const frameDuration = time.Duration(16742706) // nanoseconds ≈ 1s/59.7275
	nextFrame := time.Now().Add(frameDuration)
	for i := 0; i < 100000; i++ {
		err = core.RunFrame()
		if err != nil {
			fmt.Printf("Error running frame %d: %v\n", i+2, err)
			return
		}
		fmt.Printf("\033[%dA", height/4)
		PrintVideoBufferASCII(videobuffer, width, height)
		now := time.Now()
		if now.Before(nextFrame) {
			time.Sleep(nextFrame.Sub(now))
		}
		nextFrame = nextFrame.Add(frameDuration)
	}
}
