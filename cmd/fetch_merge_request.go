package cmd

import (
	"context"
	"fmt"

	"github.com/wayan/mergeexp/gitdir"
	"github.com/wayan/mergeexp/gitlab"
)

// fetches SHA for all MergeRequests
func FetchMergeRequest(ctx context.Context, gc *gitlab.Client, wd *gitdir.Dir, mr gitlab.MergeRequest) error {
	// sshURL for remote project
	// if the SHA of the merge request is already present, no need to fetch the repo
	if wd.ShaExists(mr.Sha) {
		return nil
	}

	// trying to fetch the url of repo
	sshURL, err := gc.ProjectSSHUrl(ctx, mr.SourceProjectId)
	if err != nil {
		return fmt.Errorf("fetching project %d: %w", mr.SourceProjectId, err)
	}

	return fetchSHA(wd, sshURL, mr.Sha)
}
