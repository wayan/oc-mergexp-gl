package cmd

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	"github.com/urfave/cli/v3"
	"github.com/wayan/mergeexp/git"
	"github.com/wayan/mergeexp/gitdir"
	"github.com/wayan/oc-mergexp-gl/cmd/flags"
)

func ActionDeployHotfix(ctx context.Context, cmd *cli.Command) error {
	workdir := cmd.String(flags.Workdir)
	if workdir == "" {
		return fmt.Errorf("no workdir set to build the branches")
	}
	if err := createDirIfNotExists(workdir); err != nil {
		return err
	}
	gd, err := gitdir.New(workdir)
	if err != nil {
		return err
	}

	slog.Info("entering work dir", "dir", workdir)
	if err := gd.GitInit(); err != nil {
		return err
	}

	sshURL := cmd.String(flags.TargetProjectSSHURL)

	shaMaster, err := git.LsRemote(gd, sshURL, "master")
	if err != nil {
		return err
	}

	tag, err := git.HighestVersionTag(gd, sshURL)
	if err != nil {
		return err
	}
	if tag == nil {
		return errors.New("missing highest tag")
	}

	masterTagged := tag.SHA == shaMaster
	if !masterTagged {
		tag.Patch++
	}
	if err := gd.Command("git", "tag", "-f", tag.String(), shaMaster).Run(); err != nil {
		// creating forcefully the local tag
		return err
	}

	if tag.SHA == shaMaster {
		slog.Info("master has the highest version tag", "tag", tag.String())
	} else {
		slog.Info("tagging master", "tag", tag.String())

		if err := gd.Command("git", "push", sshURL, tag.String()).Run(); err != nil {
			return err
		}
	}

	// fetching the tag
	if gd.Command("git", "fetch", sshURL, shaMaster).Run(); err != nil {
		return err
	}

	// pushing to production
	prodURL := cmd.String(flags.ProductionURL)
	if err := gd.Command("git", "push", prodURL, tag.String()).Run(); err != nil {
		return err
	}
	if err := gd.Command("git", "push", prodURL, shaMaster+":"+"refs/heads/master").Run(); err != nil {
		return err
	}

	return nil
}
