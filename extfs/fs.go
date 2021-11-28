package extfs

import (
	"io/fs"
	"os"
	"time"
	"errors"
)

const (
	// Mode for OpenFile in Go
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

type OpenFileFS interface {
	fs.FS
	OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)
}

type CreateFS interface {
	fs.FS
	Create(name string) (fs.File, error)
}

type MkdirFS interface {
	fs.FS
	Mkdir(name string, perm fs.FileMode) error
}

type RenameFS interface {
	fs.FS
	Rename(oldpath, newpath string) error
}

type RemoveFS interface {
	fs.FS
	Remove(name string) error
}

type ChmodFS interface {
	fs.FS
	Chmod(name string, mode fs.FileMode) error
}

type FileInfo interface {
	fs.FileInfo
	AccTime() time.Time
	User() string
	Group() string
	ModUser() string
}

// OpenFile is a helper function which do OpenFile with fs.FS.
// If fsys is CreateFS (and not OpenFileFS), then do Create and ignore perm.
func OpenFile(fsys fs.FS, name string, flag int, perm fs.FileMode) (fs.File, error) {
	if fsys, ok := fsys.(OpenFileFS); ok {
		return fsys.OpenFile(name, flag, perm)
	} else if fsys, ok := fsys.(CreateFS); ok && flag == O_RDWR|O_CREATE|O_TRUNC {
		return fsys.Create(name)
	} else if flag == O_RDONLY {
		return fsys.Open(name)
	}
	return nil, errors.New("not implemented OpenFile")
}
