package channel

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// HashChannel returns a SHA256 hex digest of the canonical YAML representation
// of the given channel. The canonical form orders map keys alphabetically so
// that semantically identical channels always produce the same hash.
func HashChannel(channel *Channel) (string, error) {
	if channel == nil {
		return "", fmt.Errorf("channel is nil")
	}

	node, err := toYAMLNode(channel)
	if err != nil {
		return "", fmt.Errorf("canonicalize channel: %w", err)
	}

	// Sort all mapping nodes recursively for determinism.
	sortMappingKeys(node)

	canonical, err := yaml.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("marshal canonical yaml: %w", err)
	}

	digest := sha256.Sum256(canonical)
	return fmt.Sprintf("%x", digest), nil
}

// toYAMLNode converts any Go value into a *yaml.Node tree.
func toYAMLNode(v any) (*yaml.Node, error) {
	// Round-trip through yaml.v3 to obtain a node tree.
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(bytes, &node); err != nil {
		return nil, err
	}
	if len(node.Content) == 0 {
		return &yaml.Node{Kind: yaml.MappingNode}, nil
	}
	return node.Content[0], nil
}

// sortMappingKeys recursively sorts the keys of every MappingNode in the tree.
func sortMappingKeys(n *yaml.Node) {
	if n == nil {
		return
	}
	if n.Kind == yaml.MappingNode && len(n.Content)%2 == 0 {
		pairs := make([][2]*yaml.Node, 0, len(n.Content)/2)
		for i := 0; i < len(n.Content); i += 2 {
			pairs = append(pairs, [2]*yaml.Node{n.Content[i], n.Content[i+1]})
		}
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i][0].Value < pairs[j][0].Value
		})
		n.Content = n.Content[:0]
		for _, p := range pairs {
			n.Content = append(n.Content, p[0], p[1])
		}
		for i := 1; i < len(n.Content); i += 2 {
			sortMappingKeys(n.Content[i])
		}
		return
	}
	for _, child := range n.Content {
		sortMappingKeys(child)
	}
}
