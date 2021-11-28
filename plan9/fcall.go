package plan9

import (
	"errors"
	"io"
)

const (
	Tversion = 100 + iota
	Rversion
	Tauth
	Rauth
	Tattach
	Rattach
	Terror
	Rerror
	Tflush
	Rflush
	Twalk
	Rwalk
	Topen
	Ropen
	Tcreate
	Rcreate
	Tread
	Rread
	Twrite
	Rwrite
	Tclunk
	Rclunk
	Tremove
	Rremove
	Tstat
	Rstat
	Twstat
	Rwstat
)

var (
	errBadFcall = errors.New("Bad Fcall")
)

type Fcall struct {
	Type uint8
	Tag  uint16
	Args []byte
}

func UnmarshalFcall(m []byte) (*Fcall, error) {
	if len(m) < 7 {
		return nil, errBadFcall
	}
	f := new(Fcall)
	Unpack(m[4:], &f.Type, &f.Tag)
	f.Args = m[7:]
	return f, nil
}

func (f *Fcall) Bytes() []byte {
	return MustPack(uint32(4 + 1 + 2 + len(f.Args)), f.Type, f.Tag, f.Args)
}

func (f *Fcall) SetErr(e error) {
	f.Type = Rerror
	s := e.Error()
	l := len(s)
	if l >= (1 << 16) {
		l = (1 << 16) - 1
		s = s[:1<<16]
	}
	f.Args = MustPack(s)
}

func ReadFcall(r io.Reader) (*Fcall, error) {
	panic("not implement")
}

func WriteFcall(w io.Writer, f *Fcall) error {
	panic("not implement")
}
