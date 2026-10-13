package storage

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
)

type FileReader interface {
	io.Reader
	io.Closer
}

type File interface {
	Open() (FileReader, error)
}

type fileHeader struct {
	h *multipart.FileHeader
}

func (h *fileHeader) Open() (FileReader, error) {
	f, err := h.h.Open()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func FileFromMultipartHeader(h *multipart.FileHeader) File {
	return &fileHeader{
		h: h,
	}
}

type fileReader struct {
	reader FileReader
}

func (f *fileReader) Open() (FileReader, error) {
	return f.reader, nil
}

func FileFromReader(reader FileReader) File {
	return &fileReader{
		reader: reader,
	}
}

type fileBytes struct {
	b []byte
}

type cloableReader struct {
	*bytes.Reader
}

func (r *cloableReader) Close() error {
	return nil
}

func (f *fileBytes) Open() (FileReader, error) {
	reader := &cloableReader{
		Reader: bytes.NewReader(f.b),
	}
	return reader, nil
}

func FileFromByteSlice(b []byte) File {
	return &fileBytes{
		b: b,
	}
}

func FileFromBuffer(b *bytes.Buffer) File {
	return &fileBytes{
		b: b.Bytes(),
	}
}

type FileObject struct {
	Name string
	Size int64
	File File
	ContentType string // Optional, can be used to store MIME type
}

type Storage interface {
	UploadFiles(ctx context.Context, e []*FileObject) error
	DeleteFiles(ctx context.Context, e []*FileObject) error
	DeleteFileByKey(ctx context.Context, key string) error
	GetFile(ctx context.Context, key string) ([]byte, error)
	FileExists(ctx context.Context, key string) (bool, error)
}
