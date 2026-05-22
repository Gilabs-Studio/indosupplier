//go:build cgo

package utils

import (
	"bytes"
	"image"
	"mime/multipart"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

// ConvertToWebP converts an image to WebP format using libwebp (CGO).
func ConvertToWebP(file multipart.File) ([]byte, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, ErrInvalidImage
	}

	if _, err := file.Seek(0, 0); err != nil {
		return nil, ErrFileProcessing
	}

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 85)
	if err != nil {
		return nil, ErrFileProcessing
	}

	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, options); err != nil {
		return nil, ErrFileProcessing
	}

	return buf.Bytes(), nil
}

// WebPEnabled reports whether WebP conversion is available.
const WebPEnabled = true
