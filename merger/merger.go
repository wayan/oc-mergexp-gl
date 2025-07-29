package merger

import "github.com/wayan/oc-mergexp-gl/gitdir"

type Merger struct {
	dir             *gitdir.Dir
	ConflictRetries int
}

func New(dir *gitdir.Dir) *Merger {
	return &Merger{
		dir:             dir,
		ConflictRetries: 3,
	}
}
