package ya9p

import (
	"os"
	"io"
	"io/fs"
	"time"
)

const (
	// Exactly one of O_RDONLY, O_WRONLY, or O_RDWR must be specified.
	O_RDONLY int = os.O_RDONLY // open the file read-only.
	O_WRONLY int = os.O_WRONLY // open the file write-only.
	O_RDWR   int = os.O_RDWR   // open the file read-write.
	// The remaining values may be or'ed in to control behavior.
	O_APPEND int = os.O_APPEND // append data to the file when writing.
	O_CREATE int = os.O_CREATE // create a new file if none exists.
	O_EXCL   int = os.O_EXCL   // used with O_CREATE, file must not exist.
	O_SYNC   int = os.O_SYNC   // open for synchronous I/O.
	O_TRUNC  int = os.O_TRUNC  // truncate regular writable file when opened.
)

type FS interface {
	fs.FS
	OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)
	Create(name string) (fs.File, error)
	Mkdir(name string, perm fs.FileMode) error
	Remove(name string) error
}

type File interface {
	fs.File
	io.Writer
	io.ReaderAt
	io.WriterAt
}

type FileInfo interface {
	fs.FileInfo
	AccTime() time.Time
	User() string
	Group() string
	ModUser() string
}
