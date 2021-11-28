package plan9

import (
	"io/fs"
)

func (m Perm) GoMode() fs.FileMode {
	return 0
}

func Mode(m fs.FileMode) Perm {
	return 0
}
