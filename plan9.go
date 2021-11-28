package ya9p

import (
	"io/fs"
)

type Qid struct {
	Type uint8
	Vers uint32
	Path uint64
}

func infoToQid(info fs.FileInfo) *Qid {
	panic("not implement")
}

func (q *Qid) Bytes() []byte {
	panic("not implement")
}

type Dir struct {
	
}

func infoToDir(info fs.FileInfo) *Dir {
	panic("not implement")
}

func (d *Dir) Bytes() []byte {
	panic("not implement")
}
