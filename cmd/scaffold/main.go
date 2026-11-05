package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/numspot/terraform-provider-numspot/internal/cli/ui"
)

var (
	// errDescriptorFlagRequired is returned when the -descriptor flag is missing.
	errDescriptorFlagRequired = errors.New("the -descriptor flag is required")
	// errProjectRootNotFound is returned when no go.mod can be found while
	// searching for the project root.
	errProjectRootNotFound = errors.New("could not find project root (go.mod not found)")
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	sub := args[0]
	rest := args[1:]
	switch sub {
	case "help", "-h", "--help":
		usage()
		return
	case "gen", "audit", "drift":
		// handled below
	default:
		ui.Errorf("unknown subcommand %q", sub)
		usage()
		os.Exit(2)
	}

	projectDir, err := findProjectRoot()
	if err != nil {
		ui.Errorf("%v", err)
		os.Exit(1)
	}

	switch sub {
	case "gen":
		err = runGen(projectDir, rest)
	case "audit":
		err = runAuditCmd(projectDir, rest)
	case "drift":
		err = runDriftCmd(projectDir, rest)
	}
	if err != nil {
		ui.Errorf("%v", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "scaffold - generate, audit and drift-check resource/datasource code from descriptors.")
	fmt.Fprintln(os.Stderr, "\nUsage:")
	fmt.Fprintln(os.Stderr, "  scaffold gen -descriptor <path> [-dry-run]   generate files for a descriptor")
	fmt.Fprintln(os.Stderr, "  scaffold audit [-descriptor <path>]          audit all (or one) descriptors")
	fmt.Fprintln(
		os.Stderr,
		"  scaffold drift [-fix]                        verify committed files match generated output; -fix regenerates drifted resources",
	)
	fmt.Fprintln(os.Stderr, "\nRun 'scaffold <subcommand> -h' for subcommand flags.")
}

// runGen generates the core/resource/datasource files for a single descriptor
// and patches provider.go. It is the consolidated replacement for the previous
// flag-less default mode of this binary.
func runGen(projectDir string, args []string) error {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	descriptorPath := fs.String("descriptor", "", "Path to the descriptor YAML file (required)")
	dryRun := fs.Bool("dry-run", false, "Print generated code without writing files")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse gen flags: %w", err)
	}

	if *descriptorPath == "" {
		fs.Usage()
		return errDescriptorFlagRequired
	}

	d, err := ParseDescriptor(*descriptorPath)
	if err != nil {
		return err
	}

	// Load signatures from the existing core file (read-only) so the generator
	// can use them as defaults.
	if err := d.LoadCoreSignatures(projectDir); err != nil {
		return err
	}

	generator := NewGenerator(d, projectDir, *dryRun)
	if err = generator.Generate(); err != nil {
		return fmt.Errorf("generating files: %w", err)
	}

	patcher := NewProviderPatcher(d, projectDir, *dryRun)
	if err = patcher.Patch(); err != nil {
		return fmt.Errorf("patching provider: %w", err)
	}

	if *dryRun {
		fmt.Println("Dry run complete. No files were written.")
	} else {
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Fill in TODO placeholders in the generated files")
		fmt.Println("  2. Run: make lint && go build ./...")
	}
	return nil
}

// runAuditCmd parses the audit subcommand flags and delegates to runAudit.
func runAuditCmd(projectDir string, args []string) error {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	descriptorPath := fs.String("descriptor", "", "Path to a single descriptor YAML file (optional; audits all when empty)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse audit flags: %w", err)
	}
	return runAudit(projectDir, *descriptorPath)
}

// runDriftCmd parses the drift subcommand flags and delegates to runDrift or runFix.
func runDriftCmd(projectDir string, args []string) error {
	fs := flag.NewFlagSet("drift", flag.ExitOnError)
	fix := fs.Bool("fix", false, "Regenerate drifted resources instead of just reporting")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse drift flags: %w", err)
	}
	if *fix {
		return runFix(projectDir)
	}
	return runDrift(projectDir)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errProjectRootNotFound
		}
		dir = parent
	}
}
