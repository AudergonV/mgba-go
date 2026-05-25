package mgba

/*
#include <mgba-util/vfs.h>
#include <fcntl.h>
#include <stdlib.h>
#include <stdbool.h>

static ssize_t vfileSize(struct VFile* vf) {
    return vf->size(vf);
}

static bool vfileClose(struct VFile* vf) {
    return vf->close(vf);
}
*/
import "C"
import (
    "errors"
    "runtime"
    "unsafe"
)

var (
    errNilOrEmptyData = errors.New("mgba: data must not be nil or empty")
    errEmptyPath      = errors.New("mgba: path must not be empty")
    errIsDirectory    = errors.New("mgba: path is a directory, not a file")
    errOpenFailed     = errors.New("mgba: VFileOpen failed")
    errMemoryFailed   = errors.New("mgba: VFileFromConstMemory failed")
    errAlreadyClosed  = errors.New("mgba: VFile is already closed")
)

type VFileType int

const (
    Unknown   VFileType = iota
    File
    Directory
)

type VFile struct {
    Type   VFileType
    ptr    *C.struct_VFile
    data   []byte
    closed bool
}

// VFileFromBytes creates a VFile from a byte slice (read-only).
// The slice must not be modified while the VFile is open.
func VFileFromBytes(data []byte) (*VFile, error) {
    if len(data) == 0 {
        return nil, errNilOrEmptyData
    }

    ptr := C.VFileFromConstMemory(unsafe.Pointer(&data[0]), C.size_t(len(data)))
    if ptr == nil {
        return nil, errMemoryFailed
    }

    vf := &VFile{
        Type: File,
        ptr:  ptr,
        data: data,
    }
    runtime.SetFinalizer(vf, (*VFile).finalize)
    return vf, nil
}

// VFileFromPath creates a VFile from a filesystem path.
// Returns an error if the path is empty, doesn't exist, or is a directory.
func VFileFromPath(path string) (*VFile, error) {
    if path == "" {
        return nil, errEmptyPath
    }

    cpath := C.CString(path)
    defer C.free(unsafe.Pointer(cpath))
    ptr := C.VFileOpen(cpath, C.O_RDONLY)
    if ptr == nil {
        return nil, errOpenFailed
    }
    size := C.vfileSize(ptr)
    if size < 0 {
        C.vfileClose(ptr)
        return nil, errIsDirectory
    }
    vf := &VFile{
        Type: File,
        ptr:  ptr,
    }
    runtime.SetFinalizer(vf, (*VFile).finalize)
    return vf, nil
}

// Close releases the underlying C resources.
// Safe to call multiple times.
func (vf *VFile) Close() error {
    if vf.closed {
        return nil
    }
    vf.closed = true
    runtime.SetFinalizer(vf, nil)
    if vf.ptr != nil {
        C.vfileClose(vf.ptr)
        vf.ptr = nil
    }
    vf.data = nil
    return nil
}

// Size returns the size in bytes of the VFile.
// Returns an error if the VFile has been closed.
func (vf *VFile) Size() (int64, error) {
    if vf.closed || vf.ptr == nil {
        return 0, errAlreadyClosed
    }
    size := C.vfileSize(vf.ptr)
    if size < 0 {
        return 0, errors.New("mgba: VFile.size() returned negative value")
    }
    return int64(size), nil
}

// finalizer to ensure resources are freed if Close() is not called (safety net for GC)
func (vf *VFile) finalize() {
    vf.Close() //nolint:errcheck
}