package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// ExtractFrontMatter parses the frontmatter and returns it as a map, along with the remaining Markdown content.
func ExtractFrontMatter(content []byte) (map[string]interface{}, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	frontMatter := make(map[string]interface{})
	var markdownContent strings.Builder
	var yamlContent strings.Builder

	// Check the first line for the frontmatter delimiter
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("empty content")
	}
	firstLine := scanner.Text()
	if firstLine != "---" && firstLine != "---yaml" {
		// No frontmatter, return the entire content as markdown
		return nil, string(content), nil
	}

	// Read frontmatter lines
	inFrontMatter := true
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			// End of frontmatter
			inFrontMatter = false
			break
		}
		if inFrontMatter {
			yamlContent.WriteString(line + "\n")
		}
	}

	// Parse the YAML frontmatter
	if inFrontMatter {
		return nil, "", fmt.Errorf("unterminated frontmatter")
	}
	if err := yaml.Unmarshal([]byte(yamlContent.String()), &frontMatter); err != nil {
		return nil, "", fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Append the rest of the content
	for scanner.Scan() {
		markdownContent.WriteString(scanner.Text() + "\n")
	}

	return frontMatter, markdownContent.String(), nil
}
