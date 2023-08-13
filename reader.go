package ya9p

import (
	"errors"
	"io"
	"sync"
)

type readerAt struct {
	reader io.Reader
	off    int64
	mu     sync.Mutex
}

func pretendReaderAt(r io.Reader) io.ReaderAt {
	if r, ok := r.(io.ReaderAt); ok {
		return r
	}
	return &readerAt{r}
}

func (r *readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.off != off {
		return errCannotSeek
	}

	n, err = ReadFull(r, p)
	r.off += int64(n)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		err = io.EOF
	}

	return n, err
}

type writerAt struct {
	writer io.Writer
	off    int64
	err    error
	mu     sync.Mutex
}

func pretendWriterAt(w io.Writer) io.WriterAt {
	if w, ok := w.(io.WriterAt); ok {
		return w
	}
	return &writerAt{w}
}

func (w *writerAt) WriteAt(p []byte, off int64) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.off != off {
		return errCannotSeek
	}

	n, err = Write(p)
	w.off += int64(n)
	return n, err
}

type dirReaderAt struct {
	lastRead *Dir
	f func() (*Dir, error)
	off int64
	err error
	mu sync.Mutex
}

func (r *dirReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.err != nil {
		return 0, r.err
	}

	if r.off != off {
		return 0, errCannotSeek
	}

	// initialize
	if r.lastRead != nil {
		r.lastRead, r.err = r.f()
	}

	n := 0
	for ; r.err == nil; r.lastRead, r.err = r.f() {
		b, _ := r.lastRead.Bytes() // (*Dir).Bytes() must not return non-nil error.
		if n+len(b) > len(p) {
			break
		}
		copy(p[n:], b)
		n += len(b)
	}

	r.off += n
	return n, r.err
}
