package ya9p

import (
	"errors"
	"io"
	"log"
	"sync"

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
	Clunk() error
	Remove() error
	Stat() (*Dir, error)
	WStat(*Dir) error
}

type Srv interface {
	Auth(user, aname string) (Fid, Qid, error)
	Attach(afid Fid, user, aname string) (Fid, Qid, error)
}

type conn struct {
	rwc    io.ReadWriteCloser
	logger *log.Logger
	s      *serveSrv
}

func Serve(rwc io.ReadWriteCloser, s Srv) error {
	c := &conn{s: &serveSrv{s: s}, rwc: rwc, logger: log.Default()}
	return c.serve()
}

func (c *conn) serve() error {
	w := make(chan *Fcall)
	var wg sync.WaitGroup

	go func() {
		for rx := range w {
			plan9.WriteFcall(c.rwc, rx) // ignore error
		}
	}()

	for {
		tx, err := plan9.ReadFcall(c.rwc)
		if err != nil {
			break
		}
		wg.Add(1)
		go func(tx *Fcall) {
			defer wg.Done()
			w <- c.s.transmit(tx)
		}(tx)
	}

	go func() {
		wg.Wait()
		close(w)
	}()
	c.rwc.Close()
	return nil
}

// do not copy
type serveSrv struct {
	s    Srv
	fids sync.Map
}

func (s *serveSrv) transmit(tx *Fcall) *Fcall {
	rx := new(Fcall)
	rx.Type = tx.Type + 1
	rx.Tag = tx.Tag

	switch tx.Type {
	case plan9.Tversion:
		s.version(rx, tx)
	case plan9.Tauth:
		s.auth(rx, tx)
	case plan9.Tflush:
		// do nothing
	case plan9.Tattach:
		s.attach(rx, tx)
	case plan9.Twalk:
		s.walk(rx, tx)
	case plan9.Tclunk:
		s.clunk(rx, tx)
	case plan9.Topen:
		s.open(rx, tx)
	case plan9.Tcreate:
		s.create(rx, tx)
	case plan9.Tread:
		s.read(rx, tx)
	case plan9.Twrite:
		s.write(rx, tx)
	case plan9.Tremove:
		s.remove(rx, tx)
	case plan9.Tstat:
		s.stat(rx, tx)
	case plan9.Twstat:
		s.wstat(rx, tx)
	default:
		setError(rx, errors.New("unknown message"))
	}
	return rx
}

func setError(f *Fcall, e error) {
	f.Type = plan9.Rerror
	f.Ename = e.Error()
}

func (s *serveSrv) version(rx, tx *Fcall) {
	rx.Type = plan9.Rversion
	rx.Version = Version
	rx.Msize = tx.Msize
}

func (s *serveSrv) auth(rx, tx *Fcall) {
	if _, ok := s.fids.Load(tx.Afid); ok {
		setError(rx, errDupFid)
		return
	}
	afid, qid, err := s.s.Auth(tx.Uname, tx.Aname)
	if err != nil {
		setError(rx, err)
		return
	}

	s.fids.Store(tx.Afid, afid)

	rx.Aqid = qid
}

func (s *serveSrv) attach(rx, tx *Fcall) {
	var afid Fid
	if tx.Afid != plan9.NOFID {
		a, ok := s.fids.Load(tx.Afid)
		if !ok {
			setError(rx, errUnknownFid)
			return
		}
		afid = a.(Fid)
	}
	if _, ok := s.fids.Load(tx.Fid); ok {
		setError(rx, errDupFid)
		return
	}
	newfid, qid, err := s.s.Attach(afid, tx.Uname, tx.Aname)
	if err != nil {
		setError(rx, err)
		return
	}
	s.fids.Store(tx.Fid, newfid)
	rx.Qid = qid
}

func (s *serveSrv) walk(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	_, ok = s.fids.LoadOrStore(tx.Newfid, nil)
	if ok && tx.Fid != tx.Newfid {
		setError(rx, errDupFid)
		return
	}
	newfid, qids, err := f.(Fid).Walk(tx.Wname)
	if err != nil {
		setError(rx, err)
		return
	}
	if len(tx.Wname) == len(qids) {
		s.fids.Store(tx.Newfid, newfid)
	} else {
		s.fids.Delete(tx.Newfid)
	}
	rx.Wqid = qids
}

func (s *serveSrv) open(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	qid, iounit, err := f.(Fid).Open(tx.Mode)
	if err != nil {
		setError(rx, err)
		return
	}
	rx.Iounit = iounit
	rx.Qid = qid
}

func (s *serveSrv) create(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	qid, iounit, err := f.(Fid).Create(tx.Name, tx.Mode, tx.Perm)
	if err != nil {
		setError(rx, err)
		return
	}
	rx.Iounit = iounit
	rx.Qid = qid
}

func (s *serveSrv) read(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	b := make([]byte, int(tx.Count))
	n, err := f.(Fid).ReadAt(b, int64(tx.Offset))

	// io.ReaderAt returns non-nil error whenever n != len(b).
	// This implement does not treat it error unless n == 0.
	if n == 0 && err != nil && err != io.EOF {
		setError(rx, err)
		return
	}

	rx.Data = b[:n]
}

func (s *serveSrv) write(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	n, err := f.(Fid).WriteAt(tx.Data, int64(tx.Offset))
	if err != nil {
		setError(rx, err)
		return
	}
	rx.Count = uint32(n)
}

func (s *serveSrv) clunk(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	s.fids.Delete(tx.Fid)
	err := f.(Fid).Clunk()
	if err != nil {
		setError(rx, err)
		return
	}
}

func (s *serveSrv) remove(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	s.fids.Delete(tx.Fid)
	err := f.(Fid).Remove()
	if err != nil {
		setError(rx, err)
		return
	}
}

func (s *serveSrv) stat(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	dir, err := f.(Fid).Stat()
	if err != nil {
		setError(rx, err)
		return
	}
	rx.Stat, _ = dir.Bytes()
}

func (s *serveSrv) wstat(rx, tx *Fcall) {
	f, ok := s.fids.Load(tx.Fid)
	if !ok || f == nil {
		setError(rx, errUnknownFid)
		return
	}
	dir, err := plan9.UnmarshalDir(tx.Stat)
	if err != nil {
		setError(rx, err)
		return
	}
	err = f.(Fid).WStat(dir)
	if err != nil {
		setError(rx, err)
		return
	}
}
