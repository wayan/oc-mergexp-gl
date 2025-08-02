package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/wayan/oc-mergexp-gl/cmd/flags"
)

// CliMergeexp returns
func (s System) CliMergexp() (*cli.Command, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %v", err)
	}

	// building flags
	varPrefix := ocpCowValue(s, "OCP_MERGEXP_", "COW_MERGEXP_")
	serviceName := ocpCowValue(s, "ocp-mergexp-gl", "cow-mergexp-gl")
	flgs := []cli.Flag{
		&cli.StringFlag{
			Name:    flags.Workdir,
			Usage:   "The directory with git repo where the branch is built.",
			Value:   homeDir + ocpCowValue(s, "/.ocp-mergexp", "/.cow-mergexp"),
			Sources: cli.EnvVars(varPrefix + "DIR"),
		},
		&cli.StringFlag{
			Name:     flags.PrivateToken,
			Usage:    "Private token to access GitLab REST API",
			Sources:  cli.EnvVars(varPrefix + "PRIVATE_TOKEN"),
			Required: true,
		},
		&cli.IntFlag{
			Name:    flags.TargetProjectID,
			Usage:   "The id of the main GitLab project",
			Value:   ocpCowValue(s, OCPTargetProjectID, CowTargetProjectID),
			Sources: cli.EnvVars(varPrefix + "PROJECT_ID"),
		},
		&cli.StringFlag{
			Name:    flags.GitLabAPIURL,
			Usage:   "GitLab REST API URL",
			Value:   GitLabAPIURL,
			Sources: cli.EnvVars(varPrefix + "GITLAB_API_URL"),
		},
		&cli.StringFlag{
			Name:  flags.StartBranch,
			Usage: "branch to start the building",
			Value: ocpCowValue(s, "develop", "master"),
		},
		&cli.IntSliceFlag{
			Name:    flags.SkipMergeRequests,
			Aliases: []string{"s"},
			Usage:   "Id of merge requests to be skipped from building the branch",
		},
		&cli.StringFlag{
			Name:    flags.DeployKey,
			Usage:   "Path to deploy key for GitLab",
			Value:   findDefaultDeployKey(),
			Sources: cli.EnvVars(varPrefix + "DEPLOY_KEY"),
		},
		&cli.StringFlag{
			Name:  flags.Test1URL,
			Usage: "URL of TEST1 environment",
			Value: ocpCowValue(s, OCPTest1URL, CowTest1URL),
		},
	}

	if s == OCP {
		flgs = append(flgs,
			&cli.StringFlag{
				Name:  flags.Test2URL,
				Usage: "URL of TEST2 environment",
				Value: OCPTest2URL,
			},
		)
	}

	return &cli.Command{
		Version:        Version,
		Name:           serviceName,
		Usage:          "building and deploying experimental branch from GitLabl merge requests",
		Flags:          flgs,
		DefaultCommand: "build",
		Action:         ActionMergexp,
	}, nil

}
