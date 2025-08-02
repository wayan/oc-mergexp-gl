package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

	masterSHA, err := git.LsRemote(gd, sshURL, Master)
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

	masterTagged := tag.SHA == masterSHA
	if !masterTagged {
		tag.Patch++
	}
	if err := gd.Command("git", "tag", "-f", tag.String(), masterSHA).Run(); err != nil {
		// creating forcefully the local tag
		return err
	}

	if tag.SHA == masterSHA {
		slog.Info("master has the highest version tag", "tag", tag.String())
	} else {
		slog.Info("tagging master", "tag", tag.String())

		if err := gd.Command("git", "push", sshURL, tag.String()).Run(); err != nil {
			return err
		}
	}

	// fetching the tag
	if gd.Command("git", "fetch", sshURL, masterSHA).Run(); err != nil {
		return err
	}

	// pushing to production
	productionURL := cmd.String(flags.ProductionURL)
	productionBranch := cmd.String(flags.ProductionBranch)
	if err := gd.Command("git", "push", productionURL, tag.String()).Run(); err != nil {
		return err
	}
	if err := gd.Command("git", "push", productionURL, masterSHA+":"+"refs/heads/"+productionBranch).Run(); err != nil {
		return err
	}

	// harmonization of develop branch, merging second parent, works for OCP only
	if developBranch := cmd.String(flags.DevelopBranch); developBranch != "" {
		if err := harmonizeDevelop(ctx, gd, sshURL, masterSHA, developBranch); err != nil {
			return err
		}

	}

	return nil
}

func harmonizeDevelop(ctx context.Context, gd *gitdir.Dir, sshURL, masterSHA, developBranch string) error {
	// fetching HEAD^2
	out, err := gd.Command("git", "rev-parse", "--verify", "-q", masterSHA+"^2").Output()
	if err != nil {
		return fmt.Errorf("master SHA (%s) apparently has no second parent, nothing to merge to develop", masterSHA)
	}
	secondParentSHA := strings.TrimSpace(string(out))
	developSHA, err := git.LsRemote(gd, sshURL, developBranch)

	// fetching the tag
	if gd.Command("git", "fetch", sshURL, developSHA).Run(); err != nil {
		return err
	}
	slog.Info("secondParent", "sha", secondParentSHA)
	if err := gd.Command("git", "merge-base", "--is-ancestor", secondParentSHA, developSHA).Run(); err == nil {
		slog.Info("develop already contains second parent of master")
		return nil
	}

	// branch has different name
	localBranch := developBranch + "-tmp"
	if err := gd.StartExperimentalBranch(localBranch, developSHA); err != nil {
		return err
	}

	if err := gd.Command("git", "merge", "--no-ff", "-m", "merging second parent of master", secondParentSHA).Run(); err != nil {
		return fmt.Errorf("merge of second parent failed: %w", err)
	}

	// pushing the develop back
	if err := gd.Command("git", "push", sshURL, localBranch+":"+developBranch).Run(); err != nil {
		return fmt.Errorf("push to develop failed: %w", err)
	}

	return nil

}
