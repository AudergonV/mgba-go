package mgba_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/audergonv/mgba-go/pkg/mgba"
)


// minimalGBAROM returns a 192-byte slice that passes the GBA magic-byte check.
func minimalGBAROM() []byte {
	rom := make([]byte, 192)
	rom[0] = 0x00
	rom[1] = 0x00
	rom[2] = 0xFF
	rom[3] = 0xEA
	return rom
}

func writeTempROM(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.gba")
	if err != nil {
		t.Fatalf("failed to create temp ROM: %v", err)
	}
	if _, err := f.Write(minimalGBAROM()); err != nil {
		t.Fatalf("failed to write temp ROM: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp ROM: %v", err)
	}
	return f.Name()
}

func TestVFileFromBytes_Valid(t *testing.T) {
	data := minimalGBAROM()
	vf, err := mgba.VFileFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer vf.Close()

	if vf.Type != mgba.File {
		t.Errorf("expected Type == File, got %v", vf.Type)
	}
}

func TestVFileFromBytes_NilSlice(t *testing.T) {
	vf, err := mgba.VFileFromBytes(nil)
	if err == nil {
		vf.Close()
		t.Fatal("expected error for nil slice, got nil")
	}
	if vf != nil {
		t.Error("expected nil VFile on error")
	}
}

func TestVFileFromBytes_EmptySlice(t *testing.T) {
	vf, err := mgba.VFileFromBytes([]byte{})
	if err == nil {
		vf.Close()
		t.Fatal("expected error for empty slice, got nil")
	}
	if vf != nil {
		t.Error("expected nil VFile on error")
	}
}

func TestVFileFromBytes_Size(t *testing.T) {
	data := minimalGBAROM()
	vf, err := mgba.VFileFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer vf.Close()

	size, err := vf.Size()
	if err != nil {
		t.Fatalf("Size() returned error: %v", err)
	}
	if size != int64(len(data)) {
		t.Errorf("expected size %d, got %d", len(data), size)
	}
}

func TestVFileFromBytes_DataLifetime(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	vf, err := mgba.VFileFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range data {
		data[i] = 0xFF
	}

	if err := vf.Close(); err != nil {
		t.Errorf("Close() after data mutation returned error: %v", err)
	}
}

func TestVFileFromPath_Valid(t *testing.T) {
	path := writeTempROM(t)

	vf, err := mgba.VFileFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer vf.Close()

	if vf.Type != mgba.File {
		t.Errorf("expected Type == File, got %v", vf.Type)
	}
}

func TestVFileFromPath_NonExistent(t *testing.T) {
	vf, err := mgba.VFileFromPath("/nonexistent/path/to/file.gba")
	if err == nil {
		vf.Close()
		t.Fatal("expected error for non-existent path, got nil")
	}
	if vf != nil {
		t.Error("expected nil VFile on error")
	}
}

func TestVFileFromPath_Directory(t *testing.T) {
	dir := t.TempDir()

	vf, err := mgba.VFileFromPath(dir)
	if err == nil {
		vf.Close()
		t.Fatal("expected error for directory path, got nil")
	}
	if vf != nil {
		t.Error("expected nil VFile on error")
	}
}

func TestVFileFromPath_EmptyString(t *testing.T) {
	vf, err := mgba.VFileFromPath("")
	if err == nil {
		vf.Close()
		t.Fatal("expected error for empty path, got nil")
	}
	if vf != nil {
		t.Error("expected nil VFile on error")
	}
}

func TestVFileFromPath_Size(t *testing.T) {
	path := writeTempROM(t)
	data := minimalGBAROM()

	vf, err := mgba.VFileFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer vf.Close()

	size, err := vf.Size()
	if err != nil {
		t.Fatalf("Size() returned error: %v", err)
	}
	if size != int64(len(data)) {
		t.Errorf("expected size %d, got %d", len(data), size)
	}
}

func TestVFileClose_Idempotent(t *testing.T) {
	data := minimalGBAROM()
	vf, err := mgba.VFileFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := vf.Close(); err != nil {
		t.Fatalf("first Close() returned error: %v", err)
	}
	if err := vf.Close(); err != nil {
		t.Errorf("second Close() returned error: %v", err)
	}
}

func TestVFileClose_SizeAfterClose(t *testing.T) {
	data := minimalGBAROM()
	vf, err := mgba.VFileFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vf.Close()

	_, err = vf.Size()
	if err == nil {
		t.Error("expected error from Size() after Close(), got nil")
	}
}

func TestVFileType_FromBytes(t *testing.T) {
	vf, err := mgba.VFileFromBytes(minimalGBAROM())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer vf.Close()

	if vf.Type == mgba.Directory {
		t.Error("VFileFromBytes must never return Type == Directory")
	}
	if vf.Type == mgba.Unknown {
		t.Error("VFileFromBytes must never return Type == Unknown")
	}
}

func TestVFileType_FromPath(t *testing.T) {
	vf, err := mgba.VFileFromPath(writeTempROM(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer vf.Close()

	if vf.Type != mgba.File {
		t.Errorf("expected Type == File for file path, got %v", vf.Type)
	}
}

func BenchmarkVFileFromBytes(b *testing.B) {
	data := make([]byte, 32*1024*1024) // 32 MiB (max GBA ROM size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vf, err := mgba.VFileFromBytes(data)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		vf.Close()
	}
}

func BenchmarkVFileFromPath(b *testing.B) {
	f, err := os.CreateTemp(b.TempDir(), "bench-*.gba")
	if err != nil {
		b.Fatal(err)
	}
	f.Write(make([]byte, 32*1024*1024))
	f.Close()
	path := filepath.Clean(f.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vf, err := mgba.VFileFromPath(path)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		vf.Close()
	}
}