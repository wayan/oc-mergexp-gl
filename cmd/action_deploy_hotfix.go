package cmd

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"log/slog"

	"github.com/urfave/cli/v3"
	"github.com/wayan/mergeexp/git"
	"github.com/wayan/mergeexp/gitdir"
	"github.com/wayan/oc-mergexp-gl/cmd/flags"
)

// deployHotfixReleaseMessage creates new empty commit with changes from previous release
func deployHotfixReleaseMessage(gd *gitdir.Dir, newRelease, prevRelease, newTag string) (string, error) {
	cmd := gd.Command("git", "log", "--merges", `--pretty=format:%x00%h%x00%s%x00%B%x00`, prevRelease+".."+newRelease)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var result = fmt.Sprintf("Release %s\n\nChangelog:\n", newTag)
	for parts := strings.Split(string(out), "\x00"); len(parts) >= 4; parts = parts[4:] {
		// index 0 is ignored it is either empty string or new line
		if line := deployHotfixReleaseMessageLine(parts[1], parts[2], parts[3]); line != "" {
			result += line + "\n"
		}
	}
	return result, nil
}

func deployHotfixReleaseMessageLine(hash, subject, body string) string {
	// if there is line starting with OMCTR- we use it as subject
	if matches := regexp.MustCompile(`(?m)^(OMCTR-.*)$`).FindStringSubmatch(body); matches != nil {
		subject = matches[1]
	}

	if matches := regexp.MustCompile(`See merge request ((\w+/.*?)!(\d+))`).FindStringSubmatch(body); matches != nil {
		const baseUrl = "https://gitlab.services.itc.st.sk"

		mr := matches[1]
		part := matches[2]
		mrid := matches[3]
		//		* 28e6b06f4 OMCTR-14357: HOTFIX - zakládání GP do PK - xml set [b2btmcz/gts-ocp!35] https://gitlab.services.itc.st.sk/b2btmcz/gts-ocp/-/merge_requests/35
		url := fmt.Sprintf("%s/%s/-/merge_requests/%s", baseUrl, part, mrid)
		return fmt.Sprintf("* %s %s [%s] %s", hash, subject, mr, url)
	}
	// without merge request
	return fmt.Sprintf("* %s %s", hash, subject)
}

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

	if err := fetchSHA(gd, sshURL, masterSHA); err != nil {
		return err
	}

	tag, err := git.HighestVersionTag(gd, sshURL)
	if err != nil {
		return err
	}

	if tag == nil {
		return errors.New("missing highest tag")
	}

	// the local tag may not be set, so I create it locally with force
	if err := gd.Command("git", "tag", "-f", tag.String(), tag.SHA).Run(); err != nil {
		// creating forcefully the local tag regardless it was bumped
		return err
	}

	if tag.SHA == masterSHA {
		slog.Info("master was already tagged by the highest version", "tag", tag.String())
	} else {
		// we create local master
		if err := gd.StartExperimentalBranch(Master, masterSHA); err != nil {
			return err
		}

		prevRelease := tag.String()
		tag.Patch++
		msg, err := deployHotfixReleaseMessage(gd, masterSHA, prevRelease, tag.String())
		if err != nil {
			return err
		}

		// create empty release commit
		if err := gd.Command("git", "commit", "-m", msg, "--allow-empty").Run(); err != nil {
			return err
		}

		out, err := gd.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {
			return err
		}

		// new master
		origMasterSHA := masterSHA
		masterSHA = strings.TrimSpace(string(out))

		slog.Info("tagging master", "tag", tag.String())
		if err := gd.Command("git", "tag", "-f", tag.String(), masterSHA).Run(); err != nil {
			// creating forcefully the local tag regardless it was bumped
			return err
		}

		// pushing new master back to GitLab
		if err := gd.Command("git", "push", sshURL, masterSHA+":"+Master).Run(); err != nil {
			return err
		}

		// pushing tag to gitLab
		if err := gd.Command("git", "push", sshURL, tag.String()).Run(); err != nil {
			return err
		}

		// harmonization of develop branch, merging second parent of the original masterSHA
		// works for OCP only
		if developBranch := cmd.String(flags.DevelopBranch); developBranch != "" {
			if err := harmonizeDevelop(ctx, gd, sshURL, origMasterSHA, developBranch); err != nil {
				return err
			}
		}
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

	if err := gd.Command("git", "merge", "--no-ff", "-m", "harmonization of master with develop", secondParentSHA).Run(); err != nil {
		return fmt.Errorf("merge of second parent failed: %w", err)
	}

	// pushing the develop back
	if err := gd.Command("git", "push", sshURL, localBranch+":"+developBranch).Run(); err != nil {
		return fmt.Errorf("push to develop failed: %w", err)
	}

	return nil

}
