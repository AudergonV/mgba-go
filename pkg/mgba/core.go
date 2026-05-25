package mgba

/*
#include <mgba/core/core.h>
#include <mgba-util/vfs.h>
#include <stdbool.h>

static bool      coreInit(struct mCore* c)     { return c->init(c); }
static void      coreDeinit(struct mCore* c)   { c->deinit(c); }
static void      coreReset(struct mCore* c)    { c->reset(c); }
static void      coreRunFrame(struct mCore* c) { c->runFrame(c); }
static void      coreRunLoop(struct mCore* c)  { c->runLoop(c); }
*/
import "C"

import "errors"

type GBACore struct {
	mCorePtr *C.struct_mCore
}

var (
	errCoreNotFound   = errors.New("mgba: mCore not found for given VFile")
	errCoreNullPtr    = errors.New("mgba: mCore pointer is null")
	errCoreInitFailed = errors.New("mgba: mCoreInit failed")
)

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

func (c *GBACore) Init() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	if !C.coreInit(c.mCorePtr) {
		return errCoreInitFailed
	}
	return nil
}

func (c *GBACore) Deinit() {
	if c.mCorePtr != nil {
		C.coreDeinit(c.mCorePtr)
		c.mCorePtr = nil
	}
}

func (c *GBACore) RunFrame() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	C.coreRunFrame(c.mCorePtr)
	return nil
}

func (c *GBACore) RunLoop() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	C.coreRunLoop(c.mCorePtr)
	return nil
}

func (c *GBACore) Reset() error {
	if c.mCorePtr == nil {
		return errCoreNullPtr
	}
	C.coreReset(c.mCorePtr)
	return nil
}

