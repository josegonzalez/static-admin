package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/josegonzalez/static-admin/markdown"
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

// RenderEditorPage serves the editor HTML page
func RenderEditorPage(c *gin.Context) {
	// Parse the HTML template
	tmpl, err := template.ParseFS(staticFiles, "assets/edit.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse template: %s", err.Error())
		return
	}

	// read example.md from the static files
	source, err := staticFiles.ReadFile("assets/example.md")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read example.md: %s", err.Error())
		return
	}

	frontmatter, content, err := ExtractFrontMatter([]byte(source))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to extract frontmatter: %s", err.Error())
		return
	}
	blocks, err := markdown.ParseMarkdownToBlocks(content)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse markdown: %s", err.Error())
		return
	}

	// Example of injecting data into the HTML
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(c.Writer, struct {
		Frontmatter string
		Blocks      string
	}{
		Frontmatter: toJSONString(frontmatter),
		Blocks:      toJSONString(blocks),
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to render template: %s", err.Error())
	}
}

// Helper to convert data to a JSON string for embedding in the template
func toJSONString(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonData)
}
