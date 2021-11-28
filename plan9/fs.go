package plan9

import (
	"io/fs"

	"github.com/rmatsuoka/ya9p/extfs"
)

const (
	// Mode for Open or Create in Plan 9.
	OREAD   = 0
	OWRITE  = 1
	ORDWR   = 2
	OEXEC   = 3
	OTRUNC  = 16
	OCEXEC  = 32
	ORCLOSE = 64
)

func Open9(fsys fs.FS, name string, mode int) (fs.File, error) {
	var o int
	switch mode & 3 {
	case OREAD:
		o = extfs.O_RDONLY
	case ORDWR:
		o = extfs.O_RDWR
	case OWRITE:
		o = extfs.O_WRONLY
	case OEXEC:
		o = extfs.O_RDONLY
	}
	if (mode & OTRUNC) != 0 {
		o |= extfs.O_TRUNC
	}

	return extfs.OpenFile(fsys, name, o, 0)
}
/*
func Create9(fsys fs.FS, name string, mode int, perm Perm) (fs.File, error) {
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
	o := extfs.O_CREATE|extfs.O_EXCL
	switch mode & 3 {
	case OREAD:
		o = extfs.O_RDONLY
	case ORDWR:
		o = extfs.O_RDWR
	case OWRITE:
		o = extfs.O_WRONLY
	case OEXEC:
		o = extfs.O_RDONLY
	}
	if (omode & OTRUNC) != 0 {
		o |= extfs.O_TRUNC
	}
	return extfs.openFile(fsys, name, o, perm&0777)
}
*/