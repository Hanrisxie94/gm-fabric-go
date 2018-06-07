package httpmetrics

import "io"

// CountReader implements the io.ReadCloser interface in ordr to count
// bytes read from an http.Request.Body
type CountReader struct {
	BytesRead int64
	Next      io.ReadCloser
}

// Read hands off the read and counts the number of bytes read
func (cr *CountReader) Read(p []byte) (int, error) {
	n, err := cr.Next.Read(p)
	cr.BytesRead += int64(n)

	return n, err
}

// Close hands off the close
func (cr *CountReader) Close() error {
	return cr.Next.Close()
}
