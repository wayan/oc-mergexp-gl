package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/wayan/oc-mergexp-gl/cmd/flags"
)

func (s System) CliDeployHotfix() (*cli.Command, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %v", err)
	}

	// building flags
	varPrefix := ocpCowValue(s, "OCP_DEPLOYHOTFIX_", "COW_DEPLOYHOTFIX_")
	serviceName := ocpCowValue(s, "ocp-deployhotfix-gl", "cow-deployhotfix-gl")
	flgs := []cli.Flag{
		&cli.StringFlag{
			Name:    flags.Workdir,
			Usage:   "The directory with git repo where the actions are run",
			Value:   homeDir + ocpCowValue(s, "/.ocp-deployhotfix", "/.cow-deployhotfix"),
			Sources: cli.EnvVars(varPrefix + "DIR"),
		},
		&cli.StringFlag{
			Name:    flags.TargetProjectSSHURL,
			Usage:   "GitLab REST API URL",
			Value:   ocpCowValue(s, OCPTargetProjectSSHURL, CowTargetProjectSSHURL),
			Sources: cli.EnvVars(varPrefix + "GITLAB_SSHURL"),
		},
		&cli.StringFlag{
			Name:    flags.ProductionURL,
			Usage:   "SSH URL to production environment",
			Value:   ocpCowValue(s, OCPProdURL, CowProdURL),
			Sources: cli.EnvVars(varPrefix + "PRODUCTION_URL"),
		},
		&cli.StringFlag{
			Name:    flags.ProductionBranch,
			Usage:   "branch on remote repo to push to",
			Value:   ocpCowValue(s, OCPProductionBranch, CowProductionBranch),
			Sources: cli.EnvVars(varPrefix + "PRODUCTION_BRANCH"),
		},
	}

	if s == OCP {
		flgs = append(
			flgs,
			&cli.StringFlag{
				Name:  flags.DevelopBranch,
				Usage: "into this branch we merge the second parent of master",
				Value: Develop,
			},
		)
	}

	return &cli.Command{
		Version:        Version,
		Name:           serviceName,
		Usage:          "bumping the tag and pushing the master branch from GitLab to production",
		Flags:          flgs,
		DefaultCommand: "build",
		Action:         ActionDeployHotfix,
	}, nil

}
