package ya9p

import (
	"os"
	"testing"

	"9fans.net/go/plan9"
)

func TestServeSrv(t *testing.T) {
	transmits := []*Fcall{
		{Type: plan9.Tattach, Fid: 0, Afid: plan9.NOFID},
		{Type: plan9.Twalk, Fid: 0, Newfid: 1, Wname: []string{"srv_test.go"}},
		{Type: plan9.Topen, Fid: 1, Mode: plan9.OREAD},
		{Type: plan9.Tread, Fid: 1, Offset: 0, Count: 100},
		{Type: plan9.Tclunk, Fid: 1},
	}
	s := &serveSrv{s: FS(os.DirFS("."))}
	for _, tx := range transmits {
		t.Log(tx)
		rx := s.transmit(tx)
		t.Log(rx)
		if rx.Type == plan9.Rerror {
			t.Fatal(rx.Ename)
		}
	}
}