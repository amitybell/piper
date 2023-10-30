package piper

import (
	"archive/tar"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
)

type tarZstReader struct {
	f io.ReadCloser
	z *zstd.Decoder
	t *tar.Reader
}

func (tz *tarZstReader) Next() (*tar.Header, error) {
	return tz.t.Next()
}

func (tz *tarZstReader) Read(p []byte) (int, error) {
	return tz.t.Read(p)
}

func (tz *tarZstReader) Close() error {
	tz.z.Close()
	if err := tz.f.Close(); err != nil {
		return fmt.Errorf("tarZstReader.Close: file: %w", err)
	}
	return nil
}

func openTarZst(f io.ReadCloser, o ...zstd.DOption) (*tarZstReader, error) {
	z, err := zstd.NewReader(f, o...)
	if err != nil {
		return nil, fmt.Errorf("openTarZst: %w", err)
	}
	t := tar.NewReader(z)
	tz := &tarZstReader{
		f: f,
		z: z,
		t: t,
	}
	return tz, nil
}
