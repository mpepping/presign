package s3client

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestProgressReader(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 1000)
	var lastRead, lastTotal int64
	callCount := 0

	pr := newProgressReader(bytes.NewReader(data), 1000, func(read, total int64) {
		lastRead = read
		lastTotal = total
		callCount++
	})

	buf := make([]byte, 100)
	for {
		_, err := pr.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if lastRead != 1000 {
		t.Errorf("lastRead = %d, want 1000", lastRead)
	}
	if lastTotal != 1000 {
		t.Errorf("lastTotal = %d, want 1000", lastTotal)
	}
	if callCount != 10 {
		t.Errorf("callCount = %d, want 10", callCount)
	}
}

func TestProgressReaderNilCallback(t *testing.T) {
	data := []byte("hello")
	pr := newProgressReader(bytes.NewReader(data), int64(len(data)), nil)

	out, err := io.ReadAll(pr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(out, data) {
		t.Errorf("got %q, want %q", out, data)
	}
}

func TestProgressReaderEmptyData(t *testing.T) {
	callCount := 0
	pr := newProgressReader(bytes.NewReader(nil), 0, func(read, total int64) {
		callCount++
	})

	out, err := io.ReadAll(pr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("got %d bytes, want 0", len(out))
	}
	if callCount != 0 {
		t.Errorf("callCount = %d, want 0 for empty reader", callCount)
	}
}

func TestProgressReaderSingleByteReads(t *testing.T) {
	data := []byte("abcde")
	var readings []int64

	pr := newProgressReader(bytes.NewReader(data), int64(len(data)), func(read, total int64) {
		readings = append(readings, read)
	})

	buf := make([]byte, 1)
	for {
		_, err := pr.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	want := []int64{1, 2, 3, 4, 5}
	if len(readings) != len(want) {
		t.Fatalf("got %d readings, want %d", len(readings), len(want))
	}
	for i, r := range readings {
		if r != want[i] {
			t.Errorf("readings[%d] = %d, want %d", i, r, want[i])
		}
	}
}

type errReader struct {
	data []byte
	pos  int
}

func (e *errReader) Read(p []byte) (int, error) {
	if e.pos >= len(e.data) {
		return 0, errors.New("disk error")
	}
	n := copy(p, e.data[e.pos:])
	e.pos += n
	return n, nil
}

func TestProgressReaderPropagatesError(t *testing.T) {
	r := &errReader{data: []byte("abc")}
	pr := newProgressReader(r, 10, func(read, total int64) {})

	// First read succeeds.
	buf := make([]byte, 10)
	n, err := pr.Read(buf)
	if err != nil {
		t.Fatalf("first Read() error: %v", err)
	}
	if n != 3 {
		t.Errorf("first Read() n = %d, want 3", n)
	}

	// Second read returns the underlying error.
	_, err = pr.Read(buf)
	if err == nil {
		t.Fatal("expected error from underlying reader")
	}
	if err.Error() != "disk error" {
		t.Errorf("error = %q, want %q", err.Error(), "disk error")
	}
}

func TestProgressReaderPreservesData(t *testing.T) {
	data := bytes.Repeat([]byte("0123456789"), 100) // 1000 bytes
	pr := newProgressReader(bytes.NewReader(data), int64(len(data)), func(_, _ int64) {})

	out, err := io.ReadAll(pr)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if !bytes.Equal(out, data) {
		t.Error("progressReader altered the data")
	}
}

func TestProgressReaderTotalPassedThrough(t *testing.T) {
	var totals []int64
	pr := newProgressReader(bytes.NewReader([]byte("ab")), 999, func(_, total int64) {
		totals = append(totals, total)
	})

	io.ReadAll(pr)

	for i, total := range totals {
		if total != 999 {
			t.Errorf("totals[%d] = %d, want 999", i, total)
		}
	}
}
