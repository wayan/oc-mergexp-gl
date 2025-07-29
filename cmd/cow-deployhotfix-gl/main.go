package main

import (
	"gitlab.services.itc.st.sk/b2btmcz/ocpdevelopers/oc-mergexp-gl/cmd"
)

func main() {
	cmd.Run(cmd.CliDeployHotfix(cmd.Cow))
}
