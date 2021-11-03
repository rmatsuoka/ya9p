package ya9p

import (
	"io/fs"
	"9fans.net/go/plan9"
)

func GoMode(m plan9.Perm) fs.FileMode {
	return 0
}

func Plan9Mode(m fs.FileMode)plan9.Perm {
	return 0
}
