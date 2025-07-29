package cmd

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	"github.com/urfave/cli/v3"
	"github.com/wayan/oc-mergexp-gl/cmd/flags"
	"github.com/wayan/oc-mergexp-gl/gitdir"
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

	shaMaster, err := gitLsRemote(gd, sshURL, "master")
	if err != nil {
		return err
	}

	tag, err := gitHighestVersionTag(gd, sshURL)
	if err != nil {
		return err
	}
	if tag == nil {
		return errors.New("missing highest tag")
	}

	fmt.Println(tag)

	newTag := fmt.Sprintf("%d.%d.%d", tag.Major, tag.Minor, tag.Patch)
	if tag.SHA == shaMaster {
		fmt.Println("master has the highest version")
		if err := gd.Command("git", "tag", "-f", newTag, shaMaster).Run(); err != nil {
			return err
		}
	} else {
		newTag = fmt.Sprintf("%d.%d.%d", tag.Major, tag.Minor, tag.Patch+1)
		fmt.Printf("master is not tagged ve use tag %s\n", newTag)

		if err := gd.Command("git", "fetch", sshURL, shaMaster).Run(); err != nil {
			return err
		}

		if err := gd.Command("git", "tag", "-f", newTag, shaMaster).Run(); err != nil {
			return err
		}

		if err := gd.Command("git", "push", sshURL, newTag).Run(); err != nil {
			return err
		}
	}

	// pushing to production
	prodURL := cmd.String(flags.ProductionURL)
	if err := gd.Command("git", "push", prodURL, newTag).Run(); err != nil {
		return err
	}
	if err := gd.Command("git", "push", prodURL, shaMaster+":"+"refs/heads/xmaster").Run(); err != nil {
		return err
	}

	return nil
}
