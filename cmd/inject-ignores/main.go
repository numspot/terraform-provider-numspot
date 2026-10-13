package main

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

const yamlStrTag = "!!str"

var scopeParams = map[string]bool{
	"spaceId":        true,
	"organisationId": true,
}

var pathParamRe = regexp.MustCompile(`\{(\w+)\}`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: inject-ignores <spec.yaml>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	var doc yaml.Node
	if err = yaml.Unmarshal(data, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	root := doc.Content[0]

	for _, section := range []string{"resources", "data_sources"} {
		sectionNode := findMappingValue(root, section)
		if sectionNode == nil {
			continue
		}
		for i := 0; i < len(sectionNode.Content); i += 2 {
			resourceNode := sectionNode.Content[i+1]
			scopeParamsInUse := collectScopeParams(resourceNode)
			if len(scopeParamsInUse) > 0 {
				injectScopeIgnores(resourceNode, scopeParamsInUse)
			}
		}
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding YAML: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(out))
}

func collectScopeParams(resourceNode *yaml.Node) map[string]bool {
	seen := map[string]bool{}
	for _, opName := range []string{"create", "delete", "read", "update"} {
		opNode := findMappingValue(resourceNode, opName)
		if opNode == nil {
			continue
		}
		pathNode := findMappingValue(opNode, "path")
		if pathNode == nil {
			continue
		}
		for _, match := range pathParamRe.FindAllStringSubmatch(pathNode.Value, -1) {
			p := match[1]
			if scopeParams[p] && !seen[p] {
				seen[p] = true
			}
		}
	}
	return seen
}

func injectScopeIgnores(resourceNode *yaml.Node, scopeParamsInUse map[string]bool) {
	schemaNode := findMappingValue(resourceNode, "schema")
	if schemaNode == nil {
		schemaNode = &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: []*yaml.Node{},
		}
		resourceNode.Content = append(resourceNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: yamlStrTag, Value: "schema"},
			schemaNode,
		)
	}

	ignoresNode := findMappingValue(schemaNode, "ignores")
	if ignoresNode == nil {
		ignoresNode = &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: []*yaml.Node{},
		}
		schemaNode.Content = append(schemaNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: yamlStrTag, Value: "ignores"},
			ignoresNode,
		)
	}

	existingIgnores := map[string]bool{}
	for _, item := range ignoresNode.Content {
		existingIgnores[item.Value] = true
	}

	for p := range scopeParamsInUse {
		if !existingIgnores[p] {
			ignoresNode.Content = append(ignoresNode.Content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   yamlStrTag,
				Value: p,
			})
		}
	}
}

func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
