package cmd

import (
	"context"
	"fmt"

	"github.com/wayan/mergeexp/gitdir"
)

func MergexpFinalCommit(ctx context.Context, wd *gitdir.Dir, shaExp string) error {
	var err error
	var commitsNotIncluded string

	// test if remote branch exists
	if shaExp != "" {
		// remote branch (experimental) exists
		out, err := wd.Command("git", "log", "--format=%h %ad %an%n     %s", "--no-merges", shaExp+"..").Output()
		if err != nil {
			return err
		}
		commitsNotIncluded = string(out)
	} else {
		commitsNotIncluded = fmt.Sprintf("Differential commits cannot be found, %q does not exist so far", Experimental)
	}

	message := fmt.Sprintf("Experimental merge")
	if true {
		// OCP specific :-(
		message = message + " NOTESTS"
	}
	message = message + "\n\n"

	out, err := wd.Command("git", "log", "--oneline", "--first-parent", shaExp+"..").Output()
	if err != nil {
		return err
	}

	message = message + string(out) + "\n\n" +
		fmt.Sprintf("Commit(s) included in this merge not present in last %q branch:",
			Experimental,
		) +
		"\n\n" + commitsNotIncluded

	if err := wd.Command("git", "commit", "--allow-empty", "--message", message).Run(); err != nil {
		return err
	}
	return nil
}
