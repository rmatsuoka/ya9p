package ya9p

import (
	"9fans.net/go/plan9"
	"io/fs"
)

func GoMode(m plan9.Perm) fs.FileMode {
	return 0
}

func Plan9Mode(m fs.FileMode) plan9.Perm {
	return 0
}
