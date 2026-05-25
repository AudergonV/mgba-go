// Author: Vincent Audergon <github.com/audergonv>
// License: MPL-2.0

package mgba

/*
#include <mgba/core/core.h>
#include <mgba-util/vfs.h>
#include <stdbool.h>

typedef uint32_t mColor; // RGBA8888

static bool      coreInit(struct mCore* c)     { return c->init(c); }
static void      coreDeinit(struct mCore* c)   { c->deinit(c); }
static void      coreReset(struct mCore* c)    { c->reset(c); }
static void      coreRunFrame(struct mCore* c) { c->runFrame(c); }
static void      coreRunLoop(struct mCore* c)  { c->runLoop(c); }
static bool      coreLoadROM(struct mCore* c, struct VFile* vf) { return c->loadROM(c, vf); }
static void      coreSetVideoBuffer(struct mCore* c, mColor* buffer, size_t stride) { c->setVideoBuffer(c, buffer, stride); }

static void coreBaseVideoSize(struct mCore* c, unsigned* width, unsigned* height) {
    c->desiredVideoDimensions(c, width, height);
}
// Allocate a framebuffer (RGBA8888, 4 bytes per pixel)
uint32_t* mgbaAllocFramebuffer(unsigned width, unsigned height, unsigned bytesPerPixel) {
    return (uint32_t*)calloc(width * height, bytesPerPixel);
}
	
static void forceHLEBios(struct mCore* c) {
    mCoreConfigSetValue(&c->config, "bios", "");
    mCoreConfigSetIntValue(&c->config, "useBios", 0);
    mCoreConfigSetIntValue(&c->config, "skipBios", 1);
}

*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

const BytesPerPixel = 4 // RGBA8888

type GBACore struct {
	mCorePtr *C.struct_mCore
	framebuffer *C.uint32_t
	framebufferSize int
}

var (
	errCoreNotFound   		= errors.New("mgba: mCore not found for given VFile")
	errCoreNullPtr    		= errors.New("mgba: mCore pointer is null")
	errCoreInitFailed 		= errors.New("mgba: mCoreInit failed")
	errCoreVFileInvalid 	= errors.New("mgba: invalid VFile for coreLoadFile")
	errCoreCouldNotLoadFile = errors.New("mgba: mCoreLoadFile failed")
)

// NewCore creates a new GBACore from a VFile.
// Call Init() next to finish setup.
func NewCore(vf *VFile) (*GBACore, error) {
	if vf == nil || vf.closed {
		return nil, errors.New("mgba: VFile is nil or closed")
	}
	mCorePtr := C.mCoreFindVF(vf.ptr)
	if mCorePtr == nil {
		return nil, errCoreNotFound
	}
	return &GBACore{mCorePtr: mCorePtr}, nil
}

// Init initializes the core, allocates a framebuffer, and configures it for use.
func (c *GBACore) Init() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	if !C.coreInit(c.mCorePtr) {
		return errCoreInitFailed
	}
	C.mCoreInitConfig(c.mCorePtr, nil)
	C.mCoreLoadConfig(c.mCorePtr)
	C.forceHLEBios(c.mCorePtr)
	var width, height C.unsigned
	C.coreBaseVideoSize(c.mCorePtr, &width, &height)
	framebuffer := C.mgbaAllocFramebuffer(width, height, BytesPerPixel)
	if framebuffer == nil {
		return errors.New("mgba: failed to allocate framebuffer")
	}
	C.coreSetVideoBuffer(c.mCorePtr, framebuffer, C.size_t(width))
	c.framebuffer = framebuffer
	c.framebufferSize = int(width * height)
	return nil
}

// Deinit deinitializes the core and frees the framebuffer.
func (c *GBACore) Deinit() {
	if c.mCorePtr != nil {
		C.coreDeinit(c.mCorePtr)
		c.mCorePtr = nil
	}
	if c.framebuffer != nil {
		C.free(unsafe.Pointer(c.framebuffer))
		c.framebuffer = nil
	}
}

// LoadROM loads a ROM from the given VFile into the core. The VFile must be valid and not closed.
func (c *GBACore) LoadROM(vf *VFile) error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	if vf == nil || vf.closed {
		return errCoreVFileInvalid
	}
	if !C.coreLoadROM(c.mCorePtr, vf.ptr) {
		return errCoreCouldNotLoadFile
	}
	// Prevent the Go wrapper from closing the VFile, the core now owns it.
	runtime.SetFinalizer(vf, nil)
	vf.closed = true
	vf.ptr = nil
	return nil
}

// RunFrame runs one frame of emulation. Must be called repeatedly to run the core.
func (c *GBACore) RunFrame() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	C.coreRunFrame(c.mCorePtr)
	return nil
}

// RunLoop runs the core in a loop until the user quits. Blocks indefinitely.
func (c *GBACore) RunLoop() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	C.coreRunLoop(c.mCorePtr)
	return nil
}

// Reset resets the core to its initial state, as if it had just been powered on.
func (c *GBACore) Reset() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	C.coreReset(c.mCorePtr)
	return nil
}

// BaseVideoSize returns the base video dimensions of the core (e.g. 240x160 for GBA).
func (c *GBACore) BaseVideoSize() (width, height int, err error) {
	if c.mCorePtr == nil {
		return 0, 0, errCoreNullPtr
	}
	var w, h C.unsigned
	C.coreBaseVideoSize(c.mCorePtr, &w, &h)
	width = int(w)
	height = int(h)
	return width, height, nil
}

// VideoBuffer returns a Go byte slice that references the core's video buffer. 
// The caller can read from this slice after each RunFrame() to get the latest video output. 
// The slice is valid as long as the core is initialized and the framebuffer is allocated.
func (c *GBACore) VideoBuffer() ([]byte, error) {
	if c.framebuffer == nil {
		return nil, errors.New("mgba: framebuffer is not allocated")
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(c.framebuffer)), c.framebufferSize*BytesPerPixel), nil
}
