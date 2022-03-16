package extfs

import (
	"errors"
	"io/fs"
	"os"
	"time"
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
	} else if fsys, ok := fsys.(CreateFS); ok && flag == os.O_RDWR|os.O_CREATE|os.O_TRUNC {
		return fsys.Create(name)
	} else if flag == os.O_RDONLY {
		return fsys.Open(name)
	}
	return nil, errors.New("not implemented OpenFile")
}
