package ya9p

import (
	"io"
	"io/fs"
)

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
