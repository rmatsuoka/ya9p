package ya9p

import (
	"errors"
	"io"
	"io/fs"
	"math/rand"
	pathpkg "path"
	"strings"

	"9fans.net/go/plan9"
)

type srvFS struct {
	fsys fs.FS
}

type fidFS struct {
	omode         int
	fsys          fs.FS
	file          fs.File
	path          string
	isDir         bool
	dirEntries    []fs.DirEntry
	readDirOffset int
	readOffset    int64
}

func ServeFS(rw io.ReadWriter, fsys fs.FS) {
	Serve(rw, &srvFS{fsys})
}

func (s *srvFS) Auth(user, aname string) (Fid, Qid, error) {
	return nil, Qid{}, NoAuthRequired
}

func (s *srvFS) Attach(afid Fid, user, aname string) (Fid, Qid, error) {
	info, err := fs.Stat(s.fsys, ".")
	if err != nil {
		return nil, Qid{}, err
	}
	return newFidFS(s.fsys, "."), FileInfoToQid(info), nil
}

func (s *srvFS) End(afid Fid) error {
	return nil
}

func newFidFS(fsys fs.FS, path string) *fidFS {
	return &fidFS{omode: -1, fsys: fsys, path: path}
}

func cleanPath(p string) string {
	s := pathpkg.Clean(p)
	if strings.HasPrefix(s, "../") {
		return "."
	}
	return s
}

func (f *fidFS) Walk(names []string) (Fid, []Qid, error) {
	if f.omode != -1 {
		return nil, nil, ErrBadUseFid
	}

	info, err := fs.Stat(f.fsys, f.path)
	if err != nil {
		return nil, nil, err
	}

	if !info.IsDir() {
		return nil, nil, ErrWalkNoDir
	}

	if len(names) == 0 {
		return newFidFS(f.fsys, f.path), []Qid{}, nil
	}

	var qids []Qid
	path := f.path
	for _, name := range names {
		path = cleanPath(path + "/" + name)
		info, err = fs.Stat(f.fsys, path)
		if err != nil {
			break
		}
		qids = append(qids, FileInfoToQid(info))
	}
	if len(qids) == 0 {
		if err == nil {
			err = errors.New("unknown errors")
		}
		return nil, nil, err
	}
	return newFidFS(f.fsys, path), qids, nil
}

func (f *fidFS) Open(mode uint8) (Qid, uint32, error) {
	var err error
	if f.omode != -1 {
		return Qid{}, 0, ErrBadUseFid
	}

	// no support ORCLOSE
	if mode&plan9.ORCLOSE != 0 {
		err = ErrPerm
	}

	// currently, open only
	if mode&3 != plan9.OREAD {
		err = ErrPerm
	}

	// no support OTRUNC
	if mode&plan9.OTRUNC != 0 {
		err = ErrPerm
	}

	if err != nil {
		return Qid{}, 0, err
	}

	file, err := f.fsys.Open(f.path)
	if err != nil {
		return Qid{}, 0, err
	}

	var info fs.FileInfo
	info, err = file.Stat()
	if err != nil {
		return Qid{}, 0, err
	}

	f.file = file
	f.omode = int(mode)
	if info.IsDir() {
		f.isDir = true
		if file, ok := file.(fs.ReadDirFile); ok {
			f.dirEntries, _ = file.ReadDir(-1)
		} else if fsys, ok := f.fsys.(fs.ReadDirFS); ok {
			f.dirEntries, _ = fsys.ReadDir(f.path)
		}
	}
	return FileInfoToQid(info), 0, nil
}

func (f *fidFS) Create(name string, mode uint8, perm Perm) (Qid, uint32, error) {
	return Qid{}, 0, ErrNoCreate
}

func (f *fidFS) ReadAt(p []byte, off int64) (int, error) {
	if f.omode == -1 {
		return 0, ErrBadUseFid
	}
	if file, ok := f.file.(io.ReaderAt); ok && !f.isDir {
		return file.ReadAt(p, off)
	}
	if off != f.readOffset {
		return 0, ErrBadOffset
	}
	var n int
	var err error
	if f.isDir {
		if f.readDirOffset == len(f.dirEntries) {
			return 0, io.EOF
		}
		l := len(p)
		for i := f.readDirOffset; i < len(f.dirEntries); i++ {
			info, err := f.dirEntries[i].Info()
			if err != nil {
				f.readDirOffset++
				continue
			}
			b, _ := FileInfoToDir(info).Bytes()
			if n+len(b) > l {
				break
			}
			copy(p[n:], b)
			n += len(b)
			f.readDirOffset++
		}
	} else {
		n, err = f.file.Read(p)
	}
	f.readOffset += int64(n)
	return n, err
}

func (f *fidFS) WriteAt(p []byte, off int64) (int, error) {
	return 0, ErrNoWrite
}

func (f *fidFS) Close() error {
	if f.omode == -1 {
		return nil
	}
	f.omode = -1
	return f.file.Close()
}

func (f *fidFS) Remove() error {
	return ErrNoRemove
}

func (f *fidFS) Stat() (*Dir, error) {
	var info fs.FileInfo
	var err error
	if f.file != nil {
		info, err = f.file.Stat()
	} else {
		info, err = fs.Stat(f.fsys, f.path)
	}
	if err != nil {
		return nil, err
	}
	return FileInfoToDir(info), nil
}

func (f *fidFS) WStat(*Dir) error {
	return ErrNoWstat
}

func FileInfoToQid(info fs.FileInfo) Qid {
	return Qid{
		Type: uint8(Plan9FileMode(info.Mode()) >> 24),
		Vers: uint32(info.ModTime().Unix()) ^ uint32(info.Size()<<8), // from u9fs
		Path: rand.Uint64(),
	}
}

func FileInfoToDir(info fs.FileInfo) *Dir {
	name := info.Name()
	if name == "." {
		name = "/"
	}
	return &Dir{
		Qid:    FileInfoToQid(info),
		Mode:   Plan9FileMode(info.Mode()),
		Atime:  uint32(info.ModTime().Unix()),
		Mtime:  uint32(info.ModTime().Unix()),
		Length: uint64(info.Size()),
		Name:   name,
		Uid:    "glenda",
		Gid:    "glenda",
		Muid:   "glenda",
	}
}

var go9modes = []struct {
	fm fs.FileMode
	p9 Perm
}{
	{fs.ModeDir, plan9.DMDIR},
	{fs.ModeAppend, plan9.DMAPPEND},
	{fs.ModeExclusive, plan9.DMEXCL},
	{fs.ModeTemporary, plan9.DMTMP},
	{fs.ModeSymlink, plan9.DMSYMLINK},
	{fs.ModeDevice, plan9.DMDEVICE},
	{fs.ModeNamedPipe, plan9.DMNAMEDPIPE},
	{fs.ModeSocket, plan9.DMSOCKET},
	{fs.ModeSetuid, plan9.DMSETUID},
	{fs.ModeSetgid, plan9.DMSETGID},
}

func Plan9FileMode(m fs.FileMode) Perm {
	var p Perm
	for _, g9 := range go9modes {
		if m&g9.fm != 0 {
			p |= g9.p9
		}
	}
	p |= Perm(m.Perm())
	return p
}
