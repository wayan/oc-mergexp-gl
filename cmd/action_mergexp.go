package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"slices"

	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v3"
	"github.com/wayan/mergeexp/gitdir"
	"github.com/wayan/mergeexp/gitlab"
	"github.com/wayan/mergeexp/merger"
	"github.com/wayan/oc-mergexp-gl/cmd/flags"
)

func buildResty(cmd *cli.Command) (*resty.Client, error) {
	rc := resty.New()
	rc.SetBaseURL(cmd.String(flags.GitLabAPIURL))
	rc.SetHeader("PRIVATE-TOKEN", cmd.String(flags.PrivateToken))
	rc.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true, // WARNING: Do NOT use in production unless you know the risks!
	})
	return rc, nil
}

func ActionMergexp(ctx context.Context, cmd *cli.Command) error {
	workdir := cmd.String(flags.Workdir)
	if workdir == "" {
		return fmt.Errorf("no workdir set to build the branches")
	}
	if err := createDirIfNotExists(workdir); err != nil {
		return err
	}
	privateToken := cmd.String(flags.PrivateToken)
	if privateToken == "" {
		return errors.New("no private token for access to GitLab REST API")
	}
	deployKey := cmd.String(flags.DeployKey)
	if _, err := os.Stat(deployKey); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %s with deployment key does not exist, either create it or set different name (see deploy-key option)", deployKey)
	}

	gd, err := gitdir.New(workdir)
	if err != nil {
		return err
	}

	slog.Info("entering work dir", "dir", workdir)
	if err := gd.GitInit(); err != nil {
		return err
	}

	rc, err := buildResty(cmd)
	if err != nil {
		return err
	}

	gc := gitlab.NewClient(rc)
	targetProjectID := cmd.Int(flags.TargetProjectID)
	mrs, err := gc.MergeRequests(ctx, targetProjectID)
	if err != nil {
		return err
	}

	if idsToSkip := cmd.IntSlice(flags.SkipMergeRequests); len(idsToSkip) > 0 {
		slog.Info("skipping", "idsToSkip", idsToSkip)
		skipped := func(mr gitlab.MergeRequest) bool {
			for _, id := range idsToSkip {
				if id == mr.ID {
					return true
				}
			}
			return false
		}
		mrs = slices.DeleteFunc(mrs, skipped)
	}

	slog.Info("Merging pull requests")
	startBranch := cmd.String(flags.StartBranch)
	sha, err := gc.BranchSHA(ctx, targetProjectID, startBranch)
	if err != nil {
		return err
	}

	sshURL, err := gc.ProjectSSHUrl(ctx, targetProjectID)
	if err != nil {
		return err
	}

	// GIT_SSH_COMMAND must be at the end of the settings
	// when run go run the GIT_SSH_COMMAND is already set as GIT_SSH_COMMAND=ssh -o ControlMaster=no -o BatchMode=yes
	gd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -o ControlMaster=no -o BatchMode=yes -i "+deployKey)
	if err := gd.Command("git", "fetch", sshURL, sha).Run(); err != nil {
		return fmt.Errorf("fetching '%s' '%s' failed: %w", sshURL, sha, err)
	}

	if err := gd.StartExperimentalBranch(Experimental, sha); err != nil {
		return err
	}

	for _, mr := range mrs {
		if err := FetchMergeRequest(ctx, gc, gd, mr); err != nil {
			return fmt.Errorf("fetching merge request %d %s failed: %w", mr.ID, mr.Title, err)
		}
	}

	// merging the merging requests
	var mergeRefs []merger.MergeRef
	for _, mr := range mrs {
		mergeRefs = append(mergeRefs, mr.MergeRef())
	}

	merger := merger.New(gd)
	if err := merger.MergeBranches(mergeRefs); err != nil {
		return fmt.Errorf("merge branches: %w", err)
	}

	test1URL := cmd.String(flags.Test1URL)

	// branch MUST be force pushed
	slog.Info("push to GitLab", "url", sshURL)
	if err := gd.Command("git", "push", "-f", sshURL, Experimental).Run(); err != nil {
		return fmt.Errorf("push to GitLab failed: %w", err)
	}

	slog.Info("push to TEST1", "url", test1URL)
	if err := gd.Command("git", "push", "-f", test1URL, Experimental+":"+Demo).Run(); err != nil {
		return fmt.Errorf("push to TEST1 environment failed: %w", err)
	}

	if test2URL := cmd.String(flags.Test2URL); test2URL != "" {
		slog.Info("push to TEST2", "url", test2URL)
		if err := gd.Command("git", "push", "-f", test2URL, Experimental+":"+Demo).Run(); err != nil {
			return fmt.Errorf("push to TEST2 environment failed: %w", err)
		}
	}
	return nil
}
