package imaging

import (
	"bytes"
	"errors"
	"image"
	"io"
	"os"

	"github.com/dhowden/tag"
	"github.com/gen2brain/webp"
)

// Only metadata is read; seek over the audio/video payload without loading it.
// Bound bytes and operations even for malformed tags or deeply nested MP4 atoms.
type metadataReader struct {
	file                      *os.File
	position, size, remaining int64
	operations                int
}

var errArtwork = errors.New("media artwork unavailable")

func (r *metadataReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 || r.operations <= 0 {
		return 0, errArtwork
	}
	r.operations--
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.file.Read(p)
	r.remaining -= int64(n)
	r.position += int64(n)
	return n, err
}

func (r *metadataReader) Seek(offset int64, whence int) (int64, error) {
	if r.operations <= 0 {
		return 0, errArtwork
	}
	r.operations--
	base := int64(0)
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		base = r.position
	case io.SeekEnd:
		base = r.size
	default:
		return 0, errArtwork
	}
	if offset < -base || offset > r.size-base {
		return 0, errArtwork
	}
	position, err := r.file.Seek(base+offset, io.SeekStart)
	if err == nil {
		r.position = position
	}
	return position, err
}

// Artwork extracts embedded album/poster art from FLAC, ID3, MP4/M4A and Ogg tags.
func Artwork(path string) (img image.Image, err error) {
	defer func() {
		if recover() != nil {
			img, err = nil, errArtwork
		}
	}()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	metadata, err := tag.ReadFrom(&metadataReader{file: f, size: stat.Size(), remaining: 20 << 20, operations: 8192})
	if err != nil {
		return nil, err
	}
	picture := metadata.Picture()
	if picture == nil || len(picture.Data) == 0 || len(picture.Data) > 20<<20 {
		return nil, errArtwork
	}
	return decodeArtwork(picture.Data)
}

func decodeArtwork(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	isWebP := len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP"
	var cfg image.Config
	var err error
	if isWebP {
		cfg, err = webp.DecodeConfig(reader)
	} else {
		cfg, _, err = image.DecodeConfig(reader)
	}
	if err != nil {
		return nil, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > MaxPixels {
		return nil, ErrTooLarge
	}
	reader.Seek(0, io.SeekStart)
	if isWebP {
		return webp.Decode(reader)
	}
	img, _, err := image.Decode(reader)
	return img, err
}
