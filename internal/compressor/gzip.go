package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

type CompressWriter struct {
	responseWriter http.ResponseWriter
	gzipWriter     *gzip.Writer
}

func NewGzipCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		responseWriter: w,
		gzipWriter:     gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.responseWriter.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.gzipWriter.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.responseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.responseWriter.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.gzipWriter.Close()
}

type CompressReader struct {
	reader     io.ReadCloser
	gzipReader *gzip.Reader
}

func NewGzipCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		reader:     r,
		gzipReader: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.gzipReader.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.gzipReader.Close()
}
