package cmd

import (
	"fmt"

	"github.com/wayan/mergeexp/gitdir"
)

func fetchSHA(wd *gitdir.Dir, sshURL, sha string) error {
	if wd.ShaExists(sha) {
		// already exists
		return nil
	}
	if err := wd.Command("git", "fetch", sshURL, sha).Run(); err != nil {
		return fmt.Errorf("fetching %q %q: %w", sshURL, sha, err)
	}
	if !wd.ShaExists(sha) {
		return fmt.Errorf("SHA not available after fetch")
	}
	return nil
}
