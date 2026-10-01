package imaging

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestArtworkBoundsAndMalformedAtoms(t *testing.T) {
	var buf bytes.Buffer
	png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	data := buf.Bytes()
	binary.BigEndian.PutUint32(data[16:20], 50_000)
	binary.BigEndian.PutUint32(data[20:24], 50_000)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	if _, err := decodeArtwork(data); err != ErrTooLarge {
		t.Fatalf("Oversized embedded cover: %v", err)
	}
	for _, raw := range [][]byte{
		[]byte("ID3\x03\x00\x00\x7f\x7f\x7f\x7fAPIC\xff\xff\xff\xff\x00\x00"),
		append([]byte{0, 0, 0, 16, 'f', 't', 'y', 'p', 'M', '4', 'A', ' ', 0, 0, 0, 0}, bytes.Repeat([]byte{0, 0, 0, 8, 'm', 'o', 'o', 'v'}, 9000)...),
	} {
		filename := filepath.Join(t.TempDir(), "bad.mp4")
		os.WriteFile(filename, raw, 0600)
		if _, err := Artwork(filename); err == nil {
			t.Fatal("Malformed metadata accepted")
		}
	}
}
