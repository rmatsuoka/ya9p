package ya9p

import (
	"io"
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

	// Mode for Open or Create in Plan 9.
	OREAD   = 0
	OWRITE  = 1
	ORDWR   = 2
	OEXEC   = 3
	OTRUNC  = 16
	OCEXEC  = 32
	ORCLOSE = 64
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

// openFile is a helper function which do OpenFile with fs.FS.
// If fsys is CreateFS (and not OpenFileFS), then perm will be ignored.
func openFile(fsys fs.FS, name string, flag int, perm fs.FileMode) (fs.File, error) {
	if fsys, ok := fsys.(OpenFileFS); ok {
		return fsys.OpenFile(name, flag, perm)
	} else if fsys, ok := fsys.(CreateFS); ok && flag == O_RDWR|O_CREATE|O_TRUNC {
		return fsys.Create(name)
	} else if flag == O_RDONLY {
		return fsys.Open(name)
	}
	return nil, errors.New("not implemented OpenFile")
}

type readDirFile struct {
	fs.File
	dirEntries []fs.DirEntry
	offset     int
	readDirErr error
}

func toReadDirFile(f fs.File, fsys fs.FS, name string) fs.ReadDirFile {
	if f, ok := f.(fs.ReadDirFile); ok {
		return f
	}
	rf := &readDirFile{ File: f }
	rf.dirEntries, rf.readDirErr = fs.ReadDir(fsys, name)
	return rf
}

func (f *readDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.readDirErr != nil {
		return []fs.DirEntry{}, f.readDirErr
	}
	d := len(f.dirEntries) - f.offset
	if d == 0 {
		if n <= 0 {
			return []fs.DirEntry{}, nil
		} else {
			return []fs.DirEntry{}, io.EOF
		}
	}
	if n > 0 && d > n {
		d = n
	}
	ret := make([]fs.DirEntry, d)
	copy(ret, f.dirEntries[f.offset:f.offset + d])
	f.offset += d
	return ret, nil
}

func open9(fsys fs.FS, name string, mode int) (fs.File, error) {
	var o int
	switch mode & 3 {
	case OREAD:
		o = O_RDONLY
	case ORDWR:
		o = O_RDWR
	case OWRITE:
		o = O_WRONLY
	case OEXEC:
		o = O_RDONLY
	}
	if (mode & OTRUNC) != 0 {
		o |= O_TRUNC
	}

	return openFile(fsys, name, o, 0)
}
/*
func create9(fsys fs.FS, name string, mode int, perm uint32) (fs.File, error) {
	if perm & DMDIR {
		mfsys, ok := fsys.(MkdirFS)
		if !ok {
			return nil, errors.New("not implement mkdir")
		}
		if (mode & ^ORCLOSE) != OREAD {
			return nil, errPerm
		}
		if err := mfsys.Mkdir(name, (perm|0400)&0777); err != nil {
			return nil, err
		}
		f, err := mfsys.Open(name)
		f = toReadDirFile(f, fsys, name)
		return f, err
	}
	o := O_CREATE|O_EXCL
	switch mode & 3 {
	case OREAD:
		o = O_RDONLY
	case ORDWR:
		o = O_RDWR
	case OWRITE:
		o = O_WRONLY
	case OEXEC:
		o = O_RDONLY
	}
	if (omode & OTRUNC) != 0 {
		o |= O_TRUNC
	}
	return openFile(fsys, name, o, perm&0777)
}
*/