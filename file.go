package ya9p

import (
	"io/fs"

	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
	"github.com/rmatsuoka/ya9p/extfs"
)

type CFS struct {
	fsys *client.Fsys
}

func newCFS() (*CFS, error) {
	return nil, nil
}

func (fsys *CFS) Open(name string) (fs.File, error) {
	return fsys.OpenFile(name, extfs.O_RDONLY, 0)
}

func (fsys *CFS) Create(name string) (fs.File, error) {
	return fsys.OpenFile(name, extfs.O_RDWR|extfs.O_CREATE|extfs.O_TRUNC, 0666)
}

func (fsys *CFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	return nil, nil
}

func (fsys *CFS) Open9(name string, mode uint8) (fs.File, error) {
	return nil, nil
}

func (fsys *CFS) Create9(name string, mode uint8, perm plan9.Perm) (fs.File, error) {
	return nil, nil
}

func (fsys *CFS) Remove(name string) error {
	return nil
}

func (fsys *CFS) Mkdir(name string) error {
	return nil
}
