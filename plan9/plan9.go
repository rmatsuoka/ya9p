package plan9

import (
	"io/fs"
)

type Perm uint32

type Qid struct {
	Type uint8
	Vers uint32
	Path uint64
}

func FileInfoToQid(info fs.FileInfo) *Qid {
	panic("not implement")
}

func (q *Qid) Bytes() []byte {
	panic("not implement")
}

type Dir struct {
	Type   uint16
	Dev    uint32
	Qid    Qid
	Mode   Perm
	Atime  uint32
	Mtime  uint32
	Length uint64
	Name   string
	Uid    string
	Gid    string
	Muid   string
}

func FileInfoToDir(info fs.FileInfo) *Dir {
	panic("not implement")
}

func (d *Dir) Bytes() []byte {
	panic("not implement")
}
