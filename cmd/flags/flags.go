package flags

// names of flags as constants
const (
	Workdir             = "workdir"
	PrivateToken        = "private-token"
	TargetProjectID     = "target-project-id"
	GitLabAPIURL        = "gitlab-api-url"
	StartBranch         = "start-branch"
	Test1URL            = "test1-url"
	Test2URL            = "test2-url"
	DeployKey           = "deploy-key"
	SkipMergeRequests   = "skip-merge-requests"
	ProductionURL       = "production-url"
	ProductionBranch    = "production-branch"
	TargetProjectSSHURL = "target-project-ssh-url"

	// branch which will be harmonized by merging master^2
	DevelopBranch = "develop-branch"
)
