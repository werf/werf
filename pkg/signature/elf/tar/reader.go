package tar

import (
	"archive/tar"
	"bytes"
	"debug/elf"
	"errors"
	"io"
)

// Reader is a wrapper around tar.Reader that works almost like tar.Reader with one difference:
// 1. It returns header.IsELF boolean field when no error.
type Reader struct {
	tr         *tar.Reader
	bodyReader *bytes.Reader
}

func NewReader(tr *tar.Reader) *Reader {
	return &Reader{
		tr:         tr,
		bodyReader: bytes.NewReader([]byte(nil)),
	}
}

// Next works almost like tar.Reader.Next() with one difference:
// 1. It returns header.IsELF boolean field when no error.
func (etr *Reader) Next() (*Header, error) {
	etr.bodyReader.Reset([]byte(nil))

	hdr, err := etr.tr.Next()
	if err != nil {
		return nil, err
	}

	if hdr.Typeflag != tar.TypeReg {
		return newHeader(hdr, false), nil
	}

	data := make([]byte, hdr.Size)

	if _, err = io.ReadFull(etr.tr, data); err != nil {
		return nil, err
	}

	etr.bodyReader.Reset(data)

	var isELF bool
	if isELF, err = isELFFileStream(etr.bodyReader); err != nil {
		return nil, err
	}

	etr.bodyReader.Reset(data)

	return newHeader(hdr, isELF), nil
}

func (etr *Reader) Read(p []byte) (n int, err error) {
	if etr.bodyReader.Size() > 0 {
		return etr.bodyReader.Read(p)
	}

	return etr.tr.Read(p)
}

// Header is a wrapper around tar.Header with an IsELF boolean field
type Header struct {
	*tar.Header

	IsELF bool
}

func newHeader(hdr *tar.Header, isELF bool) *Header {
	return &Header{hdr, isELF}
}

func isELFFileStream(readerAt io.ReaderAt) (bool, error) {
	ef, err := elf.NewFile(readerAt)
	if err != nil {
		var fe *elf.FormatError

		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &fe) {
			return false, nil
		}

		return false, err
	}
	defer ef.Close()

	switch ef.Machine {
	case elf.EM_386, elf.EM_X86_64:
		// Good ELF
		return true, nil
	default:
		// Bad ELF
		return false, nil
	}
}
