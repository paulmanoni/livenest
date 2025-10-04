package liveview

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Diff represents a Phoenix LiveView-style diff patch
// Format: { "0": { "children": { "1": { "s": ["<span>New</span>"] } } } }
type Diff map[string]interface{}

// ComputeDiff compares two HTML strings and returns Phoenix LiveView-style diffs
func ComputeDiff(oldHTML, newHTML string) (Diff, error) {
	if oldHTML == newHTML {
		return nil, nil
	}

	// For simple comparison, just parse as fragment
	oldNode, err := html.ParseFragment(strings.NewReader(oldHTML), nil)
	if err != nil || len(oldNode) == 0 {
		// If parsing fails, return full replacement
		return Diff{"0": Diff{"s": []string{newHTML}}}, nil
	}

	newNode, err := html.ParseFragment(strings.NewReader(newHTML), nil)
	if err != nil || len(newNode) == 0 {
		// If parsing fails, return full replacement
		return Diff{"0": Diff{"s": []string{newHTML}}}, nil
	}

	// ParseFragment wraps content in <html><body>...</body></html>
	// We need to unwrap to get to the actual content
	oldRoot := unwrapFragment(oldNode[0])
	newRoot := unwrapFragment(newNode[0])

	if oldRoot == nil || newRoot == nil {
		return Diff{"0": Diff{"s": []string{newHTML}}}, nil
	}

	// Compare the unwrapped content
	diff := diffNodes(oldRoot, newRoot, 0)
	if len(diff) == 0 {
		return nil, nil
	}

	return diff, nil
}

// diffNodes recursively diffs two HTML nodes
func diffNodes(oldNode, newNode *html.Node, index int) Diff {
	diff := make(Diff)

	// Handle text nodes differently
	if oldNode.Type == html.TextNode && newNode.Type == html.TextNode {
		if oldNode.Data != newNode.Data {
			diff[toString(index)] = Diff{"s": []string{newNode.Data}}
		}
		return diff
	}

	// If nodes are completely different types or tags, replace entirely
	if oldNode.Type != newNode.Type || oldNode.Data != newNode.Data {
		// Return static replacement "s": [html]
		diff[toString(index)] = Diff{"s": []string{renderNode(newNode)}}
		return diff
	}

	// Check if attributes changed
	if oldNode.Type == html.ElementNode && !sameAttributes(oldNode, newNode) {
		// For now, replace the whole node if attributes differ
		diff[toString(index)] = Diff{"s": []string{renderNode(newNode)}}
		return diff
	}

	// Diff children
	oldChildren := getChildren(oldNode)
	newChildren := getChildren(newNode)

	if len(oldChildren) != len(newChildren) {
		// Different number of children - replace entire node
		diff[toString(index)] = Diff{"s": []string{renderNode(newNode)}}
		return diff
	}

	// Recursively diff each child
	childrenDiff := make(Diff)
	for i := 0; i < len(oldChildren); i++ {
		childDiff := diffNodes(oldChildren[i], newChildren[i], i)
		if len(childDiff) > 0 {
			for k, v := range childDiff {
				childrenDiff[k] = v
			}
		}
	}

	if len(childrenDiff) > 0 {
		diff[toString(index)] = Diff{"children": childrenDiff}
	}

	return diff
}

// getChildren returns all child nodes (element and text)
func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}
	return children
}

// sameAttributes checks if two nodes have the same attributes
func sameAttributes(oldNode, newNode *html.Node) bool {
	if len(oldNode.Attr) != len(newNode.Attr) {
		return false
	}

	oldAttrs := make(map[string]string)
	for _, attr := range oldNode.Attr {
		oldAttrs[attr.Key] = attr.Val
	}

	for _, attr := range newNode.Attr {
		if oldVal, ok := oldAttrs[attr.Key]; !ok || oldVal != attr.Val {
			return false
		}
	}

	return true
}

// renderNode renders an HTML node back to string
func renderNode(node *html.Node) string {
	var sb strings.Builder
	html.Render(&sb, node)
	return sb.String()
}

// toString converts an integer to string for use as map key
func toString(i int) string {
	return strconv.Itoa(i)
}

// MarshalDiff converts a Diff to JSON
func MarshalDiff(diff Diff) ([]byte, error) {
	return json.Marshal(diff)
}

// debugNodeStructure prints the structure of a node tree for debugging
func debugNodeStructure(node *html.Node, depth int) {
	if depth > 3 {
		return
	}
	indent := strings.Repeat("  ", depth)
	nodeType := "?"
	switch node.Type {
	case html.ElementNode:
		nodeType = "Element"
	case html.TextNode:
		nodeType = "Text"
	case html.DocumentNode:
		nodeType = "Document"
	}

	data := node.Data
	if node.Type == html.TextNode {
		data = strings.TrimSpace(data)
		if len(data) > 20 {
			data = data[:20] + "..."
		}
		data = `"` + data + `"`
	}

	childCount := 0
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		childCount++
	}

	log.Printf("%s[%d] %s (%s) children:%d %s", indent, depth, node.Data, nodeType, childCount, data)

	index := 0
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		log.Printf("%s  [%d] %s", indent, index, getNodeName(child))
		if child.Type == html.ElementNode {
			debugNodeStructure(child, depth+1)
		}
		index++
	}
}

func getNodeName(node *html.Node) string {
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		if len(text) > 20 {
			text = text[:20] + "..."
		}
		return "#text \"" + text + "\""
	}
	return node.Data
}

// unwrapFragment extracts the actual content from ParseFragment's html/body wrapper
// ParseFragment returns: <html><head></head><body>CONTENT</body></html>
// We need to extract CONTENT (first child of body)
func unwrapFragment(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}

	// If it's an <html> node, find the <body>
	if node.Type == html.ElementNode && node.Data == "html" {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "body" {
				// Return first non-whitespace child of body
				for bodyChild := child.FirstChild; bodyChild != nil; bodyChild = bodyChild.NextSibling {
					// Skip empty text nodes
					if bodyChild.Type == html.TextNode && strings.TrimSpace(bodyChild.Data) == "" {
						continue
					}
					return bodyChild
				}
			}
		}
	}

	// If not wrapped, return as-is
	return node
}
