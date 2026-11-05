package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// PollingStates captures pending/target state slices used by the provider for
// state polling after an asynchronous API operation.
type PollingStates struct {
	Pending []string `yaml:"pending"`
	Target  []string `yaml:"target"`
}

// Operation describes how a single operation (create/update/delete) is handled
// by the provider: asynchronous or not, retry pattern, timeout.
type Operation struct {
	Async               bool          `yaml:"async"`
	PollingStates       PollingStates `yaml:"polling_states"`
	RetryPattern        string        `yaml:"retry_pattern"`
	TimeoutDefault      string        `yaml:"timeout_default"`
	TimeoutConfigurable bool          `yaml:"timeout_configurable"`
}

// Resource describes one Terraform resource exposed by the Numspot provider.
type Resource struct {
	Name            string               `yaml:"name"`
	TerraformName   string               `yaml:"terraform_name"`
	Category        string               `yaml:"category"`
	OwnerTeam       string               `yaml:"owner_team"`
	Scope           string               `yaml:"scope"`
	Operations      map[string]Operation `yaml:"operations"`
	BackendStatuses []string             `yaml:"backend_statuses"`
	Notes           string               `yaml:"notes"`
}

// Registry is the top-level shape of resources-registry.yaml.
type Registry struct {
	Resources []Resource `yaml:"resources"`
}

const (
	registryPath     = "docs/internal/resources-registry.yaml"
	outputPath       = "docs/internal/RESOURCE_REGISTRY.md"
	outputPathBackup = "docs/internal/RESOURCE_REGISTRY.md.bak"
	ownersPath       = "OWNERS.md"
	codeownersPath   = ".github/CODEOWNERS"
)

// teamHandles maps the registry owner_team values to GitHub team handles used
// in CODEOWNERS. Teams must exist in the GitHub org and have explicit write
// access to the repository, otherwise GitHub silently skips the CODEOWNERS
// line (no owner is assigned).
var teamHandles = map[string]string{
	"numspot-iaas": "@numspot/numspot-iaas",
	"numspot-paas": "@numspot/numspot-paas",
}

// handleFor returns the CODEOWNERS handle for an owner_team, plus a marker when
// the team is unknown so the generated file flags it instead of silently
// dropping ownership.
func handleFor(ownerTeam string) string {
	if h, ok := teamHandles[ownerTeam]; ok {
		return h
	}
	return fmt.Sprintf("# TODO unknown team %q", ownerTeam)
}

// folderFor derives the Go service folder from a resource name, mirroring
// Descriptor.PackageName() in cmd/scaffold (concatenated lowercase, no
// underscores), e.g. kubernetes_cluster -> kubernetescluster.
func folderFor(name string) string {
	return strings.ReplaceAll(name, "_", "")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "resource-registry: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	registry, err := parseRegistry(registryPath)
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	if err := backupExisting(outputPath); err != nil {
		return fmt.Errorf("backing up previous registry: %w", err)
	}

	markdown := generateMarkdown(registry)
	if err := os.WriteFile(outputPath, []byte(markdown), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outputPath, err)
	}
	fmt.Printf("Generated %s\n", outputPath)

	if err := os.WriteFile(ownersPath, []byte(generateOwnersMarkdown(registry)), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", ownersPath, err)
	}
	fmt.Printf("Generated %s\n", ownersPath)

	if err := os.MkdirAll(filepath.Dir(codeownersPath), 0o755); err != nil {
		return fmt.Errorf("creating %s dir: %w", filepath.Dir(codeownersPath), err)
	}
	if err := os.WriteFile(codeownersPath, []byte(generateCodeowners(registry)), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", codeownersPath, err)
	}
	fmt.Printf("Generated %s\n", codeownersPath)

	return nil
}

func parseRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	// Stable order within categories.
	sort.SliceStable(reg.Resources, func(i, j int) bool {
		if reg.Resources[i].Category != reg.Resources[j].Category {
			return reg.Resources[i].Category < reg.Resources[j].Category
		}
		return reg.Resources[i].Name < reg.Resources[j].Name
	})

	return &reg, nil
}

func backupExisting(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := os.Rename(path, outputPathBackup); err != nil {
		return fmt.Errorf("failed to rename existing file: %w", err)
	}
	if err := os.Remove(outputPathBackup); err != nil {
		return fmt.Errorf("failed to remove backup file: %w", err)
	}
	return nil
}

//nolint:misspell // Generated French documentation contains valid French words flagged by the English misspell linter.
func generateMarkdown(reg *Registry) string {
	var buf bytes.Buffer

	buf.WriteString("# Référentiel des ressources Terraform Numspot\n\n")
	buf.WriteString("> Ce document est **généré automatiquement** depuis ")
	buf.WriteString("[`docs/internal/resources-registry.yaml`](resources-registry.yaml).\n")
	buf.WriteString("> Ne pas éditer manuellement : exécuter `make generate-resource-registry`.\n\n")

	buf.WriteString("## Vue d'ensemble\n\n")
	buf.WriteString(
		"Ce référentiel centralise, par catégorie de la console Numspot, les informations suivantes pour chaque ressource exposée par le provider Terraform :\n\n",
	)
	buf.WriteString("- **Async create/delete** : indique si le provider effectue du *state polling* après l'appel API.\n")
	buf.WriteString("- **États possibles (polling)** : états `pending` → `target` surveillés par le provider.\n")
	buf.WriteString("- **Retry pattern** : fonction(s) de retry utilisées.\n")
	buf.WriteString("- **Timeout configurable** : présence d'un champ `timeouts` ou `timeout` dans le schéma Terraform.\n")
	buf.WriteString("- **Backend statuses** : statuts réels du backend Numspot (à compléter depuis les repos backend).\n\n")

	buf.WriteString("## Légende\n\n")
	buf.WriteString("- `Oui` : le provider attend que la ressource atteigne un état final.\n")
	buf.WriteString("- `Non` : pas de polling côté provider (opération considérée comme synchrone par le provider).\n")
	buf.WriteString("- `N/A` : opération non applicable ou non implémentée.\n\n")

	// Group by category preserving a console-friendly order.
	categories := orderedCategories(reg.Resources)
	for _, category := range categories {
		resources := resourcesInCategory(reg.Resources, category)
		if category == "IAM" && len(resources) == 0 {
			fmt.Fprintf(&buf, "## %s\n\n", category)
			buf.WriteString("_Aucune ressource IAM n'est exposée actuellement par le provider._\n\n")
			continue
		}

		fmt.Fprintf(&buf, "## %s\n\n", category)
		buf.WriteString(
			"| Ressource | Scope | Owner | Async create | Async delete | États provider (pending → target) | Retry pattern | Timeout configurable | Backend statuses | Notes |\n",
		)
		buf.WriteString("|---|---|---|---|---|---|---|---|---|---|\n")

		for i := range resources {
			buf.WriteString(resourceRow(&resources[i]))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("---\n\n")

	buf.WriteString("## Notes complémentaires\n\n")
	buf.WriteString(
		"- Le timeout par défaut des fonctions de retry est fixé à `5m` pour les opérations API et `15m` pour le polling d'état (`TfRequestStateRetryTimeout`).\n",
	)
	buf.WriteString(
		"- La majorité des suppressions utilisent `RetryDeleteUntilResourceAvailable`, qui est actuellement impacté par le bug B1 du plan de standardisation (codes de retry de création utilisés à la place de ceux de suppression).\n",
	)
	buf.WriteString("- Voir également : [RETRY_ASYNC_STANDARDIZATION_PLAN.md](RETRY_ASYNC_STANDARDIZATION_PLAN.md).\n")

	return buf.String()
}

// orderedTeams returns the distinct owner_team values across resources, sorted
// for deterministic output.
func orderedTeams(resources []Resource) []string {
	seen := make(map[string]bool)
	var teams []string
	for i := range resources {
		if resources[i].OwnerTeam != "" && !seen[resources[i].OwnerTeam] {
			seen[resources[i].OwnerTeam] = true
			teams = append(teams, resources[i].OwnerTeam)
		}
	}
	sort.Strings(teams)
	return teams
}

func resourcesForTeam(resources []Resource, team string) []Resource {
	var out []Resource
	for i := range resources {
		if resources[i].OwnerTeam == team {
			out = append(out, resources[i])
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// generateOwnersMarkdown renders the human-readable root OWNERS.md: which team
// owns which Terraform resources (from owner_team).
func generateOwnersMarkdown(reg *Registry) string {
	var buf bytes.Buffer

	buf.WriteString("# Owners\n\n")
	buf.WriteString("> This file is **auto-generated** from ")
	buf.WriteString("[`docs/internal/resources-registry.yaml`](docs/internal/resources-registry.yaml).\n")
	buf.WriteString("> Do not edit by hand: run `make generate-resource-registry`.\n\n")

	buf.WriteString("## Resources by team\n\n")
	buf.WriteString("Each Terraform resource is owned by the team declared in its `owner_team` field in the registry.\n\n")
	for _, team := range orderedTeams(reg.Resources) {
		fmt.Fprintf(&buf, "### %s\n\n", team)
		teamResources := resourcesForTeam(reg.Resources, team)
		for i := range teamResources {
			fmt.Fprintf(&buf, "- `%s` — `internal/services/%s/`\n", teamResources[i].TerraformName, folderFor(teamResources[i].Name))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// generateCodeowners renders the functional GitHub .github/CODEOWNERS: each
// service folder -> its owner_team handle.
func generateCodeowners(reg *Registry) string {
	var buf bytes.Buffer

	buf.WriteString("# CODEOWNERS — auto-generated from docs/internal/resources-registry.yaml\n")
	buf.WriteString("# Do not edit by hand: run `make generate-resource-registry`.\n\n")

	buf.WriteString("# Per-resource ownership (from owner_team)\n")
	for _, team := range orderedTeams(reg.Resources) {
		teamResources := resourcesForTeam(reg.Resources, team)
		for i := range teamResources {
			path := fmt.Sprintf("/internal/services/%s/", folderFor(teamResources[i].Name))
			fmt.Fprintf(&buf, "%-36s %s\n", path, handleFor(team))
		}
	}

	return buf.String()
}

func orderedCategories(resources []Resource) []string {
	seen := make(map[string]bool)
	order := []string{}

	consoleOrder := []string{
		"Services Managés",
		"Compute",
		"Réseau",
		"Stockage",
		"VPN",
		"Direct Link",
		"IAM",
	}

	for _, c := range consoleOrder {
		for i := range resources {
			if resources[i].Category == c && !seen[c] {
				seen[c] = true
				order = append(order, c)
			}
		}
	}

	// Append any category not explicitly listed above.
	for i := range resources {
		if !seen[resources[i].Category] {
			seen[resources[i].Category] = true
			order = append(order, resources[i].Category)
		}
	}

	return order
}

func resourcesInCategory(resources []Resource, category string) []Resource {
	var out []Resource
	for i := range resources {
		if resources[i].Category == category {
			out = append(out, resources[i])
		}
	}
	return out
}

func resourceRow(r *Resource) string {
	create := r.Operations["create"]
	deleteOp := r.Operations["delete"]

	return fmt.Sprintf(
		"| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
		fmt.Sprintf("%#q", r.TerraformName),
		r.Scope,
		r.OwnerTeam,
		asyncText(create.Async),
		asyncText(deleteOp.Async),
		statesText(&create),
		create.RetryPattern,
		timeoutText(create.TimeoutConfigurable),
		backendStatusesText(r.BackendStatuses),
		strings.TrimSpace(r.Notes),
	)
}

func asyncText(async bool) string {
	if async {
		return "Oui"
	}
	return "Non"
}

func statesText(op *Operation) string {
	if !op.Async {
		return "—"
	}
	pending := strings.Join(op.PollingStates.Pending, ", ")
	target := strings.Join(op.PollingStates.Target, ", ")
	return fmt.Sprintf("%s → %s", pending, target)
}

func timeoutText(configurable bool) string {
	if configurable {
		return "Oui"
	}
	return "Non"
}

func backendStatusesText(statuses []string) string {
	if len(statuses) == 0 {
		return "_à compléter_"
	}
	return strings.Join(statuses, ", ")
}
