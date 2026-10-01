package httpapi

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"myself/server/internal/imaging"
)

func flacArtwork(picture []byte) []byte {
	var block bytes.Buffer
	for _, value := range []uint32{3, 9} {
		binary.Write(&block, binary.BigEndian, value)
	}
	block.WriteString("image/png")
	for _, value := range []uint32{0, 24, 24, 24, 0, uint32(len(picture))} {
		binary.Write(&block, binary.BigEndian, value)
	}
	block.Write(picture)
	n := block.Len()
	return append(append([]byte("fLaC"), 0x86, byte(n>>16), byte(n>>8), byte(n)), block.Bytes()...)
}

func id3Artwork(picture []byte) []byte {
	data := append([]byte{0}, []byte("image/png\x00\x03\x00")...)
	data = append(data, picture...)
	var frame bytes.Buffer
	frame.WriteString("APIC")
	binary.Write(&frame, binary.BigEndian, uint32(len(data)))
	frame.Write([]byte{0, 0})
	frame.Write(data)
	n := frame.Len()
	return append([]byte{'I', 'D', '3', 3, 0, 0, byte(n>>21) & 127, byte(n>>14) & 127, byte(n>>7) & 127, byte(n) & 127}, frame.Bytes()...)
}

func artworkAtom(name string, data []byte) []byte {
	var atom bytes.Buffer
	binary.Write(&atom, binary.BigEndian, uint32(8+len(data)))
	atom.WriteString(name)
	atom.Write(data)
	return atom.Bytes()
}

func mp4Artwork(picture []byte) []byte {
	data := artworkAtom("data", append([]byte{0, 0, 0, 14, 0, 0, 0, 0}, picture...))
	metadata := artworkAtom("moov", artworkAtom("udta", artworkAtom("meta", append([]byte{0, 0, 0, 0}, artworkAtom("ilst", artworkAtom("covr", data))...))))
	return append(artworkAtom("ftyp", []byte("M4A \x00\x00\x00\x00")), metadata...)
}

func TestEmbeddedMediaArtworkThumbnails(t *testing.T) {
	e := newEnv(t)
	picture := pngOf(color.RGBA{230, 40, 60, 255})
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"album.flac", flacArtwork(picture)}, {"album.mp3", id3Artwork(picture)},
		{"album.m4a", mp4Artwork(picture)}, {"poster.mp4", mp4Artwork(picture)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := e.uploadOne(tc.name, tc.data)
			if item.Thumb != thumbURL(item.Name) {
				t.Fatal("Missing artwork URL")
			}
			res := e.do(http.MethodGet, item.Thumb, nil, nil)
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode != 200 || !bytes.HasPrefix(body, []byte("RIFF")) {
				t.Fatalf("Thumbnail failed: %d", res.StatusCode)
			}
			thumb := filepath.Join(e.root, "uploads", "thumbs", item.Name+".webp")
			meta, err := imaging.Meta(thumb)
			if err != nil || meta.Width != 24 || meta.Height != 24 {
				t.Fatalf("Wrong cover dimensions: %+v %v", meta, err)
			}
			if original, err := os.ReadFile(filepath.Join(e.root, "uploads", item.Name)); err != nil || !bytes.Equal(original, tc.data) {
				t.Fatal("Cover extraction changed media content")
			}
			e.call(http.MethodDelete, "/api/v1/admin/media/"+item.Name, nil, nil, 200)
			if _, err := os.Stat(thumb); !os.IsNotExist(err) {
				t.Fatal("Deleting media retained its cover")
			}
		})
	}
}

func TestArtworkFailureAndChangedSource(t *testing.T) {
	e := newEnv(t)
	item := e.uploadOne("no-cover.flac", []byte("fLaC invalid metadata"))
	for range 2 {
		res := e.do(http.MethodGet, item.Thumb, nil, nil)
		res.Body.Close()
		if res.StatusCode != 404 {
			t.Fatal("Invalid metadata should use the placeholder")
		}
	}
	filename := filepath.Join(e.root, "uploads", item.Name)
	if err := os.WriteFile(filename, flacArtwork(pngOf(color.RGBA{20, 190, 100, 255})), 0600); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Second)
	os.Chtimes(filename, future, future)
	res := e.do(http.MethodGet, item.Thumb, nil, nil)
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal("Changed media did not invalidate failed-cover cache")
	}
}
