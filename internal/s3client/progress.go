package s3client

import "io"

// ProgressFunc is called during upload with the number of bytes read so far and the total size.
type ProgressFunc func(bytesRead, totalSize int64)

type progressReader struct {
	reader     io.Reader
	total      int64
	read       int64
	onProgress ProgressFunc
}

func newProgressReader(r io.Reader, total int64, fn ProgressFunc) *progressReader {
	return &progressReader{reader: r, total: total, onProgress: fn}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.read += int64(n)
		if pr.onProgress != nil {
			pr.onProgress(pr.read, pr.total)
		}
	}
	return n, err
}
