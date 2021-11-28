package ya9p

import (
	"9fans.net/go/plan9"
	"io/fs"
	"time"
)

type info struct {
	d *plan9.Dir
}

func GoFileInfo(p *plan9.Dir) fs.FileInfo {
	return fs.FileInfo(&info{p})
}
func (i *info) Name() string       { return i.d.Name }
func (i *info) Size() int64        { return int64(i.d.Length) }
func (i *info) Mode() fs.FileMode  { return GoMode(i.d.Mode) }
func (i *info) ModTime() time.Time { return time.Unix(int64(i.d.Mtime), 0) }
func (i *info) IsDir() bool        { return GoMode(i.d.Mode).IsDir() }
func (i *info) Sys() interface{}   { return i.d }

func (i *info) Type() fs.FileMode { return GoMode(i.d.Mode) & fs.ModeType }
func (i *info) Info() (fs.FileInfo, error) {
	return fs.FileInfo(i), nil
}

func (i *info) AccTime() time.Time { return time.Unix(int64(i.d.Atime), 0) }
func (i *info) User() string       { return i.d.Uid }
func (i *info) Group() string      { return i.d.Gid }
func (i *info) ModUser() string    { return i.d.Muid }
