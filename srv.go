package ya9p

import (
	"errors"
	"fmt"
	"io"
	"os"

	"9fans.net/go/plan9"
)

type Fcall = plan9.Fcall
type Qid = plan9.Qid
type Perm = plan9.Perm
type Dir = plan9.Dir

var (
	Version = "9P2000"
)

var (
	NoAuthRequired = errors.New("no authentication required")
	ErrAuth        = errors.New("authentication failed")
	ErrBadOffset   = errors.New("bad offset")
	ErrBadUseFid   = errors.New("bad use of fid")
	ErrNoCreate    = errors.New("create prohibited")
	ErrNoRemove    = errors.New("remove prohibited")
	ErrNoStat      = errors.New("stat prohibited")
	ErrNoWrite     = errors.New("write prohibited")
	ErrNoWstat     = errors.New("wstat prohibited")
	ErrPerm        = errors.New("permission denied")
	ErrWalkNoDir   = errors.New("walk in non-directory")
)

var (
	errUnknownFid  = errors.New("unknown fid")
	errDupFid      = errors.New("duplicate fid")
	errNotAttached = errors.New("not attached")
)

type Fid interface {
	Walk(name []string) (Fid, []Qid, error)
	Open(mode uint8) (Qid, uint32, error)
	Create(name string, mode uint8, perm Perm) (Qid, uint32, error)
	io.ReaderAt
	io.WriterAt
	io.Closer
	Remove() error
	Stat() (*Dir, error)
	WStat(*Dir) error
}

type Srv interface {
	Auth(user, aname string) (Fid, Qid, error)
	Attach(afid Fid, user, aname string) (Fid, Qid, error)
	End(afid Fid) error
}

type conn struct {
	s      Srv
	afid   Fid
	fids   map[uint32]Fid
	rw     io.ReadWriter
	logger io.Writer
}

func Serve(rw io.ReadWriter, s Srv) {
	c := &conn{s: s, fids: make(map[uint32]Fid), rw: rw, logger: os.Stderr}
	c.serve()
}

func (c *conn) serve() {
	for {
		rx, err := plan9.ReadFcall(c.rw)
		if err != nil {
			fmt.Fprintln(c.logger, err)
			return
		}

		var tx *Fcall
		switch rx.Type {
		case plan9.Tversion:
			tx = c.version(rx)
		case plan9.Tauth:
			tx = c.auth(rx)
		case plan9.Tattach:
			tx = c.attach(rx)
		case plan9.Twalk:
			tx = c.walk(rx)
		case plan9.Tclunk:
			tx = c.clunk(rx)
		case plan9.Topen:
			tx = c.open(rx)
		case plan9.Tcreate:
			tx = c.create(rx)
		case plan9.Tread:
			tx = c.read(rx)
		case plan9.Twrite:
			tx = c.write(rx)
		case plan9.Tremove:
			tx = c.remove(rx)
		case plan9.Tstat:
			tx = c.stat(rx)
		case plan9.Twstat:
			tx = c.wstat(rx)
		default:
			tx = errFcall(errors.New("unknown message"))
		}

		tx.Tag = rx.Tag
		err = plan9.WriteFcall(c.rw, tx)
		if err != nil {
			fmt.Fprintln(c.logger, err)
			return
		}
	}
	c.s.End(c.afid)
}

func errFcall(e error) *Fcall {
	return &Fcall{Type: plan9.Rerror, Ename: e.Error()}
}

func (c *conn) version(rx *Fcall) *Fcall {
	return &Fcall{Type: plan9.Rversion, Version: Version, Msize: rx.Msize}
}

func (c *conn) auth(rx *Fcall) *Fcall {
	if _, ok := c.fids[rx.Afid]; ok {
		return errFcall(errDupFid)
	}
	afid, qid, err := c.s.Auth(rx.Uname, rx.Aname)
	if err != nil {
		return errFcall(err)
	}
	c.afid = afid
	c.fids[rx.Afid] = afid
	return &Fcall{Type: plan9.Rauth, Aqid: qid}
}

func (c *conn) attach(rx *Fcall) *Fcall {
	var afid Fid
	if rx.Afid != plan9.NOFID {
		var ok bool
		afid, ok = c.fids[rx.Afid]
		if !ok {
			return errFcall(errUnknownFid)
		}
	}
	if _, ok := c.fids[rx.Fid]; ok {
		return errFcall(errDupFid)
	}
	newfid, qid, err := c.s.Attach(afid, rx.Uname, rx.Aname)
	if err != nil {
		return errFcall(err)
	}
	c.fids[rx.Fid] = newfid
	return &Fcall{Type: plan9.Rattach, Qid: qid}
}

func (c *conn) walk(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	if _, ok = c.fids[rx.Newfid]; ok && rx.Fid != rx.Newfid {
		return errFcall(errDupFid)
	}
	newfid, qids, err := f.Walk(rx.Wname)
	if err != nil {
		return errFcall(err)
	}
	if len(rx.Wname) == len(qids) {
		c.fids[rx.Newfid] = newfid
	}
	return &Fcall{Type: plan9.Rwalk, Wqid: qids}
}

func (c *conn) open(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	qid, iounit, err := f.Open(rx.Mode)
	if err != nil {
		return errFcall(err)
	}
	return &Fcall{Type: plan9.Ropen, Iounit: iounit, Qid: qid}
}

func (c *conn) create(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	qid, iounit, err := f.Create(rx.Name, rx.Mode, rx.Perm)
	if err != nil {
		return errFcall(err)
	}
	return &Fcall{Type: plan9.Rcreate, Iounit: iounit, Qid: qid}
}

func (c *conn) read(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	b := make([]byte, int(rx.Count))
	n, err := f.ReadAt(b, int64(rx.Offset))
	if err != io.EOF && err != nil {
		return errFcall(err)
	}
	return &Fcall{Type: plan9.Rread, Data: b[:n]}
}

func (c *conn) write(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	n, err := f.WriteAt(rx.Data, int64(rx.Offset))
	if err != nil {
		return errFcall(err)
	}
	return &Fcall{Type: plan9.Rwrite, Count: uint32(n)}
}

func (c *conn) clunk(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	delete(c.fids, rx.Fid)
	err := f.Close()
	if err != nil {
		return errFcall(err)
	}
	return &Fcall{Type: plan9.Rclunk}
}

func (c *conn) remove(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	delete(c.fids, rx.Fid)
	err := f.Remove()
	if err != nil {
		return errFcall(err)
	}
	return &Fcall{Type: plan9.Rremove}
}

func (c *conn) stat(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	dir, err := f.Stat()
	if err != nil {
		return errFcall(err)
	}
	b, _ := dir.Bytes()
	return &Fcall{Type: plan9.Rstat, Stat: b}
}

func (c *conn) wstat(rx *Fcall) *Fcall {
	f, ok := c.fids[rx.Fid]
	if !ok {
		return errFcall(errUnknownFid)
	}
	dir, err := plan9.UnmarshalDir(rx.Stat)
	if err != nil {
		return errFcall(err)
	}
	err = f.WStat(dir)
	if err != nil {
		return errFcall(err)
	}
	return &Fcall{Type: plan9.Rwstat}
}
