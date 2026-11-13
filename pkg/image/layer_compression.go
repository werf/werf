package image

import (
	"fmt"
	"strconv"
)

const (
	LayerCompressionGzip = "gzip"
	LayerCompressionZstd = "zstd"
	LayerCompressionNone = "none"

	LayerCompressionLevelMinGzip = 0
	LayerCompressionLevelMaxGzip = 9
	LayerCompressionLevelMinZstd = 1
	LayerCompressionLevelMaxZstd = 22
)

// LayerCompression is how the layers of an image are compressed when it is published. The zero
// value is the default: gzip at the default level of the compressor.
type LayerCompression struct {
	Format string
	Level  *int
}

func (c LayerCompression) Validate() error {
	switch c.Format {
	case "", LayerCompressionGzip, LayerCompressionZstd:
	case LayerCompressionNone:
		if c.Level != nil {
			return fmt.Errorf("unexpected compression level %d: the %s compression format has no level", *c.Level, LayerCompressionNone)
		}
		return nil
	default:
		return fmt.Errorf("unexpected compression format %q: expected %s, %s or %s", c.Format, LayerCompressionGzip, LayerCompressionZstd, LayerCompressionNone)
	}

	if c.Level == nil {
		return nil
	}

	format, minLevel, maxLevel := LayerCompressionGzip, LayerCompressionLevelMinGzip, LayerCompressionLevelMaxGzip
	if c.Format == LayerCompressionZstd {
		format, minLevel, maxLevel = LayerCompressionZstd, LayerCompressionLevelMinZstd, LayerCompressionLevelMaxZstd
	}
	if *c.Level < minLevel || *c.Level > maxLevel {
		return fmt.Errorf("unexpected compression level %d: expected %d-%d for the %s compression format", *c.Level, minLevel, maxLevel, format)
	}

	return nil
}

// PublishesOCIImage reports whether images with this compression are published in the OCI image
// format: zstd and uncompressed layers have no Docker image format media types.
func (c LayerCompression) PublishesOCIImage() bool {
	return c.Format == LayerCompressionZstd || c.Format == LayerCompressionNone
}

func (c LayerCompression) IsSet() bool {
	return c.Format != "" || c.Level != nil
}

// IsDefault reports whether the layers are published exactly as without any compression settings.
func (c LayerCompression) IsDefault() bool {
	return (c.Format == "" || c.Format == LayerCompressionGzip) && c.Level == nil
}

// DigestInput is the stage digest input for the compression: empty for the default, so that
// stages built without compression settings keep their digests.
func (c LayerCompression) DigestInput() string {
	if c.IsDefault() {
		return ""
	}

	format := c.Format
	if format == "" {
		format = LayerCompressionGzip
	}
	if c.Level == nil {
		return format
	}

	return format + ":" + strconv.Itoa(*c.Level)
}
