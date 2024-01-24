package atomicredteam

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func ExpandStringSlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}

	var r []string

	for _, e := range s {
		t := strings.Split(e, ",")
		r = append(r, t...)
	}

	return r
}

func Configure(ctx *cli.Context) {
	if HasBundledTechniques() || REPO != "" {
		BUNDLED = true
	} else {
		REPO = ctx.String("repo")
	}

	if local := ctx.String("local-atomics-path"); local != "" {
		LOCAL = local
	}

	TEMPDIR, _ = os.MkdirTemp(os.TempDir(), "goart-")

	Println(string(Logo()))
}

func Teardown(ctx *cli.Context) {
	os.RemoveAll(TEMPDIR)
}
