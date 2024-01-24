package atomicredteam

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"actshad.dev/go-atomicredteam/types"
)

var (
	LOCAL   string
	REPO    string
	BUNDLED bool
	TEMPDIR string

	AtomicsFolderRegex = regexp.MustCompile(`PathToAtomicsFolder(\\|\/)`)
	BlockQuoteRegex    = regexp.MustCompile(`<\/?blockquote>`)
)

//go:embed include/*
var include embed.FS

func Logo() []byte {
	logo, err := include.ReadFile("include/logo.txt")
	if err != nil {
		panic(err)
	}

	return logo
}

func HasBundledTechniques() bool {
	// will return true if directory exists
	_, err := include.ReadDir("include/atomics")
	return err == nil
}

func Techniques() []string {
	var techniques []string

	entries, err := include.ReadDir("include/atomics")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "T") {
			techniques = append(techniques, entry.Name())
		}
	}

	entries, err = include.ReadDir("include/custom")
	if err != nil {
		return techniques
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "T") {
			techniques = append(techniques, entry.Name())
		}
	}

	return techniques
}

func Technique(tid string) ([]byte, string, error) {
	// Check for a custom atomic first, then public.
	if body, err := include.ReadFile("include/custom/" + tid + "/" + tid + ".yaml"); err == nil {
		return body, "include/custom/", nil
	}

	if body, err := include.ReadFile("include/custom/" + tid + "/" + tid + ".yml"); err == nil {
		return body, "include/custom/", nil
	}

	if body, err := include.ReadFile("include/atomics/" + tid + "/" + tid + ".yaml"); err == nil {
		return body, "include/atomics/", nil
	}

	if body, err := include.ReadFile("include/atomics/" + tid + "/" + tid + ".yml"); err == nil {
		return body, "include/atomics/", nil
	}

	return nil, "", fmt.Errorf("Atomic Test is not currently bundled")
}

func Markdown(tid string) ([]byte, error) {
	var (
		body []byte
		err  error
	)

	// Check for a custom atomic first, then public.
	body, err = include.ReadFile("include/custom/" + tid + "/" + tid + ".md")
	if err != nil {
		body, err = include.ReadFile("include/atomics/" + tid + "/" + tid + ".md")
		if err != nil {
			return nil, fmt.Errorf("Atomic Test is not currently bundled")
		}
	}

	return body, nil
}

func LoadEmulation(emulationPath string) (*types.Emulation, error) {
	body, err := os.ReadFile(emulationPath)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("emulation file is empty")
	}

	var emulation types.Emulation

	if err := yaml.Unmarshal(body, &emulation); err != nil {
		return nil, fmt.Errorf("processing Atomic Emulation YAML file: %w", err)
	}

	return &emulation, nil
}

func InvokeEmulation(ctx *cli.Context) error {
	emulationPath := ctx.String("emulation-path")
	emulation, err := LoadEmulation(emulationPath)
	if err != nil {
		return err
	}

	for _, atomic := range emulation.Atomics {
		for _, test := range atomic.AtomicTests {

			var index int = -1
			var inputs []string = ExpandStringSlice(ctx.StringSlice("input"))

			var env []string
			_, err := Execute(atomic.AttackTechnique, test.Name, index, inputs, env)
			if err != nil {
				fmt.Printf("Error while executing atomic test: %v", err.Error())
			}
		}
	}

	return nil
}

func InvokeAtomic(ctx *cli.Context) error {
	var (
		tid    = ctx.String("technique")
		name   = ctx.String("name")
		index  = ctx.Int("index")
		inputs = ExpandStringSlice(ctx.StringSlice("input"))
		env    = ExpandStringSlice(ctx.StringSlice("env"))
	)

	if tid != "" && (name != "" || index != -1) {
		// Only honor --quiet flag if actually executing a test.
		Quiet = ctx.Bool("quiet")
	}

	Println(string(Logo()))

	if name != "" && index != -1 {
		return cli.Exit("only provide one of 'name' or 'index' flags", 1)
	}

	if tid == "" {
		filter := make(map[string]struct{})

		listTechniques := func() ([]string, error) {
			var (
				techniques   []string
				descriptions []string
			)

			for technique := range filter {
				techniques = append(techniques, technique)
			}

			sort.Strings(techniques)

			for _, tid := range techniques {
				technique, err := GetTechnique(tid)
				if err != nil {
					return nil, fmt.Errorf("unable to get technique %s: %w", tid, err)
				}

				descriptions = append(
					descriptions,
					fmt.Sprintf("%s - %s", tid, technique.DisplayName),
				)
			}

			return descriptions, nil
		}

		getLocalTechniques := func() error {
			files, err := ioutil.ReadDir(LOCAL)
			if err != nil {
				return fmt.Errorf(
					"unable to read contents of provided local atomics path: %w",
					err,
				)
			}

			for _, f := range files {
				if f.IsDir() && strings.HasPrefix(f.Name(), "T") {
					filter[f.Name()] = struct{}{}
				}
			}

			return nil
		}

		if BUNDLED {
			// Get bundled techniques first.
			for _, asset := range Techniques() {
				filter[asset] = struct{}{}
			}

			// We want to get local techniques after getting bundled techniques so
			// the local techniques will replace any bundled techniques with the
			// same ID.
			if LOCAL != "" {
				if err := getLocalTechniques(); err != nil {
					return cli.Exit(err.Error(), 1)
				}
			}

			descriptions, err := listTechniques()
			if err != nil {
				cli.Exit(err.Error(), 1)
			}

			Println("Locally Available Techniques:\n")

			for _, desc := range descriptions {
				Println(desc)
			}

			return nil
		}

		// Even if we're not running in bundled mode, still see if the user
		// wants to load any local techniques.
		if LOCAL != "" {
			if err := getLocalTechniques(); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			descriptions, err := listTechniques()
			if err != nil {
				cli.Exit(err.Error(), 1)
			}

			Println("Locally Available Techniques:\n")

			for _, desc := range descriptions {
				Println(desc)
			}
		}

		orgBranch := strings.Split(REPO, "/")

		if len(orgBranch) != 2 {
			return cli.Exit("repo must be in format <org>/<branch>", 1)
		}

		url := fmt.Sprintf(
			"https://github.com/%s/atomic-red-team/tree/%s/atomics",
			orgBranch[0],
			orgBranch[1],
		)

		Printf("Please see %s for a list of available default techniques", url)

		return nil
	}

	if name == "" && index == -1 {
		if dump := ctx.String("dump-technique"); dump != "" {
			dir, err := DumpTechnique(dump, tid)
			if err != nil {
				return cli.Exit("error dumping technique: "+err.Error(), 1)
			}

			Printf("technique %s files dumped to %s", tid, dir)

			return nil
		}

		technique, err := GetTechnique(tid)
		if err != nil {
			return cli.Exit("error getting details for "+tid, 1)
		}

		Printf("Technique: %s - %s\n", technique.AttackTechnique, technique.DisplayName)
		Println("Tests:")

		for i, t := range technique.AtomicTests {
			Printf("  %d. %s\n", i, t.Name)
		}

		md, err := GetMarkdown(tid)
		if err != nil {
			return cli.Exit("error getting Markdown for "+tid, 1)
		}

		if runtime.GOOS == "windows" {
			Println(string(md))
		} else {
			options := []glamour.TermRendererOption{glamour.WithWordWrap(100)}

			if ctx.Bool("no-color") {
				options = append(options, glamour.WithColorProfile(termenv.Ascii))
			} else {
				options = append(options, glamour.WithStylePath("dark"))
			}

			renderer, err := glamour.NewTermRenderer(options...)
			if err != nil {
				return cli.Exit("error creating new Markdown renderer", 1)
			}

			out, err := renderer.RenderBytes(md)
			if err != nil {
				return cli.Exit("error rendering Markdown for "+tid, 1)
			}

			Print(string(out))
		}

		return nil
	}

	var err error

	TEMPDIR, err = os.MkdirTemp("", "goart-")
	if err != nil {
		return cli.Exit(err, 1)
	}

	defer os.RemoveAll(TEMPDIR)

	test, err := Execute(tid, name, index, inputs, env)
	if err != nil {
		return cli.Exit(err, 1)
	}

	var (
		plan []byte
		ext  = strings.ToLower(ctx.String("results-format"))
	)

	switch ext {
	case "json":
		plan, _ = json.Marshal(test)
	case "yaml":
		plan, _ = yaml.Marshal(test)
	default:
		return cli.Exit("unknown results format provided", 1)
	}

	out := ctx.String("results-file")

	if out == "-" {
		Println()
		fmt.Println(string(plan))
		return nil
	}

	if out == "" {
		now := strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", ".")
		out = fmt.Sprintf("atomic-test-executor-execution-%s-%s.%s", tid, now, ext)
	}

	ioutil.WriteFile(out, plan, 0644)

	return nil
}
