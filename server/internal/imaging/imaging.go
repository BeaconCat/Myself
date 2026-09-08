// Package imaging 提供纯 Go 的图片解码、裁切与压缩（替代 sharp）。
// 支持 png / jpeg / gif / webp；webp 经 libwebp WASM（无需 cgo）。
package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
)

// Info 是图片元数据。
type Info struct {
	Width    int
	Height   int
	HasAlpha bool
}

// Ext 返回小写扩展名（含点）。
func Ext(name string) string {
	return strings.ToLower(filepath.Ext(name))
}

// Meta 读取尺寸与透明通道信息。
func Meta(path string) (Info, error) {
	f, err := os.Open(path)
	if err != nil {
		return Info{}, err
	}
	defer f.Close()
	var cfg image.Config
	switch Ext(path) {
	case ".png":
		cfg, err = png.DecodeConfig(f)
	case ".jpg", ".jpeg":
		cfg, err = jpeg.DecodeConfig(f)
	case ".gif":
		cfg, err = gif.DecodeConfig(f)
	case ".webp":
		cfg, err = webp.DecodeConfig(f)
	default:
		return Info{}, fmt.Errorf("unsupported format %q", Ext(path))
	}
	if err != nil {
		return Info{}, err
	}
	return Info{Width: cfg.Width, Height: cfg.Height, HasAlpha: modelHasAlpha(cfg.ColorModel)}, nil
}

func modelHasAlpha(m color.Model) bool {
	switch m {
	case color.NRGBAModel, color.NRGBA64Model, color.AlphaModel, color.Alpha16Model:
		return true
	}
	if p, ok := m.(color.Palette); ok {
		for _, c := range p {
			if _, _, _, a := c.RGBA(); a < 0xffff {
				return true
			}
		}
	}
	return false
}

// Decode 按扩展名解码整张图。
func Decode(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	switch Ext(path) {
	case ".png":
		return png.Decode(f)
	case ".jpg", ".jpeg":
		return jpeg.Decode(f)
	case ".gif":
		return gif.Decode(f)
	case ".webp":
		return webp.Decode(f)
	}
	return nil, fmt.Errorf("unsupported format %q", Ext(path))
}

// Encode 按扩展名编码；quality 仅对有损格式生效（1-100）。
func Encode(w io.Writer, img image.Image, ext string, quality int) error {
	switch ext {
	case ".png":
		return (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(w, img)
	case ".jpg", ".jpeg":
		return jpeg.Encode(w, toOpaqueRGBA(img), &jpeg.Options{Quality: quality})
	case ".gif":
		return gif.Encode(w, img, &gif.Options{NumColors: 256})
	case ".webp":
		return webp.Encode(w, img, webp.Options{Quality: quality, Method: 4, Exact: true})
	}
	return fmt.Errorf("unsupported format %q", ext)
}

// EncodeBytes 编码为内存字节。
func EncodeBytes(img image.Image, ext string, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, img, ext, quality); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Rect 是裁切框。
type Rect struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Crop 从 img 提取 rect 区域（拷贝为独立图像，避免子图引用整幅内存）。
func Crop(img image.Image, r Rect) image.Image {
	b := img.Bounds()
	sub := image.Rect(b.Min.X+r.Left, b.Min.Y+r.Top, b.Min.X+r.Left+r.Width, b.Min.Y+r.Top+r.Height).Intersect(b)
	out := image.NewNRGBA(image.Rect(0, 0, sub.Dx(), sub.Dy()))
	draw.Draw(out, out.Bounds(), img, sub.Min, draw.Src)
	return out
}

// toOpaqueRGBA 将带透明的图像铺白底后输出，供 JPEG 编码。
func toOpaqueRGBA(img image.Image) image.Image {
	if _, ok := img.(*image.YCbCr); ok {
		return img
	}
	out := image.NewRGBA(img.Bounds())
	draw.Draw(out, out.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Over)
	return out
}

// Fit 等比缩放到最长边不超过 maxSide；已够小则原样返回。
func Fit(img image.Image, maxSide int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxSide && h <= maxSide {
		return img
	}
	if w >= h {
		h = h * maxSide / w
		w = maxSide
	} else {
		w = w * maxSide / h
		h = maxSide
	}
	out := image.NewNRGBA(image.Rect(0, 0, max(1, w), max(1, h)))
	draw.CatmullRom.Scale(out, out.Bounds(), img, b, draw.Over, nil)
	return out
}
