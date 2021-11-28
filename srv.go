package ya9p

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
)

var (
	errUsedFid = errors.New("used fid")
	errBadFid  = errors.New("bad fid")
)

type fid struct {
	path  string
	file  fs.File
	omode int
	info  fs.FileInfo
	dirErr error
	dirOffset int
	dirEntries []fs.DirEntry
}

type srv struct {
	rwc    io.ReadWriteCloser
	logger *log.Logger
	fsys   fs.FS
	fids   map[uint32]*fid
}

func ListenSrv(network, addr string, fsys fs.FS) error {
	logger := log.Default()
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		s := &srv{
			rwc:    conn,
			logger: logger,
			fsys:   fsys,
			fids:   make(map[uint32]*fid),
		}
		go s.srv()
	}
}

func (s *srv) srv() {
	defer s.rwc.Close()
	// auth and attach

	for {
	}
}

func realpath(pwd, relpath string) string {
	panic("not implement")
}

func (s *srv) walk(rx *Fcall, tx *Fcall) {
	var tag, nwname uint16
	var fi, newfi uint32
	var b []byte
	_, err := unpack(tx.Args, &tag, &fi, &newfi, &nwname, &b)
	if err != nil {
		s.logger.Println(errBadFcall)
		rx.SetErr(errBadFcall)
		return
	}

	_, ok := s.fids[newfi]
	if (fi != newfi) && !ok {
		s.logger.Println(errUsedFid)
		rx.SetErr(errUsedFid)
		return
	}

	var qbytes []byte
	var info fs.FileInfo
	var path string = s.fids[fi].path
	var nwqid uint16 = 0
	n := 0
	for ; nwqid < nwname; nwqid++ {
		var wname string
		n, err = unpack(b[n:], &wname)
		if err != nil {
			s.logger.Println(errBadFcall)
			rx.SetErr(errBadFcall)
			return
		}
		path = realpath(path, wname)

		info, err = fs.Stat(s.fsys, path)
		if err != nil {
			s.logger.Printf("srv walk: %v", err)
			break
		}
		qbytes = append(qbytes, infoToQid(info).Bytes()...)
	}

	if nwqid == nwname {
		f := new(fid)
		f.omode = -1
		f.path = path
		f.info = info
		s.fids[newfi] = f
	} else {
		if nwqid == 0 {
			rx.SetErr(errors.New("cannot walk"))
			return
		}
	}

	rx.Args = mustPack(nwqid, qbytes)
}

func (s *srv) open(rx *Fcall, tx *Fcall) {
	var tag uint16
	var fi uint32
	var omode uint8
	_, err := unpack(tx.Args, &tag, &fi, &omode)
	if err != nil {
		s.logger.Println(errBadFcall)
		rx.SetErr(errBadFcall)
	}

	f, ok := s.fids[fi]
	if !ok {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}

	if f.omode != -1 {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}

	var o int
	switch omode & 3 {
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

	var file fs.File
	file, err = openFile(s.fsys, f.path, o, 0)
	if err != nil {
		s.logger.Print(err)
		rx.SetErr(err)
		return
	}

	info, statErr := file.Stat()
	if statErr != nil {
		err := fmt.Errorf("srv open: %v", statErr)
		s.logger.Print(err)
		rx.SetErr(err)
		return
	}

	if info.IsDir() {
		file = toReadDirFile(file, s.fsys, f.path)
	}
	f.omode = int(omode)
	f.file = file
	f.info = info

	rx.Args = mustPack(infoToQid(info).Bytes(), 0)
}

func (s *srv) create(rx *Fcall, tx *Fcall) {
	var fi, perm uint32
	var name string
	var mode uint8
	_, err := unpack(tx.Args, &fi, &name, &perm, &mode)
	if err != nil {
		s.logger.Println(errBadFcall)
		rx.SetErr(errBadFcall)
		return
	}

	f, ok := s.fids[fi]
	if !ok {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}

	if f.omode != -1 {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}
}

/*

func (s *srv) read(rx *Fcall, tx *Fcall) {
	var fid, count uint32
	var offset uint64
	_, err := unpack(tx.Args, &fid, &offset, &count)
	if err != nil {
		s.logger.Println(errBadFcall)
		rx.SetErr(errBadFcall)
		return
	}

	f, ok := s.fids[fi]
	if !ok {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}

	if f.omode != -1 {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}

	p := make([]byte, count)
	c := 0
	if f.info.IsDir() {

	} else {
		if file, ok := f.File.(io.ReaderAt); ok {
			c, err := file.ReadAt(p, offset)
		} else {
			c, err := f.file.Read(p)
		}
		if err == io.EOF {
			p = []byte{}
			c = 0
			err = nil
		}
		if err != nil {
			rx.SetErr(err)
			s.logger.Println(err)
			return
		}
	}
	rx.Args = pack(uint32(rc), p)
}

func (s *srv) write(rx *Fcall, tx *Fcall) {
	var fid, count uint32
	var offset uint64
	var p []byte
	_, err := unpack(tx.Args, &fid, &offset, &count, &p)
	if err != nil {
		s.logger.Println(errBadFcall)
		rx.SetErr(errBadFcall)
		return
	}

	f, ok := s.fids[fi]
	if !ok {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}

	if f.omode != -1 {
		s.logger.Print(errBadFid)
		rx.SetErr(errBadFid)
		return
	}

	var c int
	if file, ok := f.File.(io.WriterAt); ok {
		c, err = file.WriteAt(p, offset)
	} else if file, ok := f.File.(io.Write); ok {
		c, err = file.Write(p)
	} else {
		err = &fs.PathError{Op: "write", Path: f.path, Err: errors.New("not implemented")}
		s.logger.Println(err)
		rx.SetErr(err)
		return
	}

	if err != nil {
		s.logger.Println(err)
		rx.SetErr(err)
		return
	}

	rx.Args = pack(uint32(c))
}

func (s *srv) clunk(rx *Fcall, tx *Fcall) {

}
*/