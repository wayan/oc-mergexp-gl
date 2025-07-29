package cmd

import (
	"cmp"
	"fmt"
	"iter"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/wayan/oc-mergexp-gl/gitdir"
)

func gitLsRemote(gd *gitdir.Dir, args ...string) (string, error) {
	args = append([]string{"ls-remote"}, args...)
	cmd := gd.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("fetching remote failed: %w", err)
	}
	s := string(out)
	idxWhite := strings.IndexFunc(s, unicode.IsSpace)
	if idxWhite < 0 {
		return "", fmt.Errorf("unexpected output from ls-remote: %s", s)
	}

	return s[:idxWhite], nil
}

func parseOutputTable(output []byte) iter.Seq[[]string] {
	s := string(output)

	return func(yield func([]string) bool) {
		for s != "" {
			line := s
			rest := ""
			if idx := strings.Index(s, "\n"); idx >= 0 {
				line = s[:idx]
				rest = s[idx+1:]
			}
			if !yield(strings.Fields(line)) {
				return
			}
			s = rest
		}
	}
}

func parseVersionTag(ref string) *versionTag {

	// Define a regex to capture the version part, assuming it starts after the last slash
	// and consists of digits and dots.
	// 2.5.0^{}
	re := regexp.MustCompile(`^(?:.*/)?(\d+)\.(\d+)\.(\d+)(\{.*)?$`)
	if matches := re.FindStringSubmatch(ref); len(matches) > 1 {
		vt := versionTag{peeled: matches[4] != ""}
		vt.Major, _ = strconv.Atoi(matches[1])
		vt.Minor, _ = strconv.Atoi(matches[2])
		vt.Patch, _ = strconv.Atoi(matches[3])
		return &vt
	}
	return nil
}

type versionTag struct {
	Major, Minor, Patch int
	SHA                 string
	peeled              bool
}

// compareVersionTags compares two versionTag structs.
// It prioritizes Major, then Minor, then Patch.
// If all version numbers are equal, it compares based on the peeled field:
// - If both are peeled or both are not peeled, they are considered equal (0).
// - If t1 is not peeled and t2 is peeled, t1 is considered lesser (-1).
// - If t1 is peeled and t2 is not peeled, t1 is considered greater (1).
func compareVersionTags(t1, t2 versionTag) int {
	if c := cmp.Compare(t1.Major, t2.Major); c != 0 {
		return c
	}
	if c := cmp.Compare(t1.Minor, t2.Minor); c != 0 {
		return c
	}
	if c := cmp.Compare(t1.Patch, t2.Patch); c != 0 {
		return c
	}

	// Logic for peeled comparison (as per your modified requirements)
	// A peeled tag is considered "greater" than a non-peeled tag if versions are equal.
	if t1.peeled == t2.peeled {
		return 0 // Both have same peeled status
	}
	// At this point, t1.peeled != t2.peeled
	if t1.peeled {
		return 1 // t1 is peeled, t2 is not peeled: t1 is greater
	}
	return -1 // t1 is not peeled, t2 is peeled: t1 is lesser
}

func gitHighestVersionTag(gd *gitdir.Dir, url string) (*versionTag, error) {
	cmd := gd.Command("git", "ls-remote", "--tags", "--sort=v:refname", url)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fetching remote failed: %w", err)
	}

	// parsing the output
	var tags []versionTag
	for row := range parseOutputTable(out) {
		if len(row) < 2 {
			continue
		}
		if vt := parseVersionTag(row[1]); vt != nil {
			vt.SHA = row[0]
			tags = append(tags, *vt)
		}
		// 2.5.0^{}
	}
	if len(tags) == 0 {
		return nil, nil
	}
	slices.SortFunc(tags, compareVersionTags)
	return &(tags[len(tags)-1]), nil
}
