package main

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// Block represents a structured block of content.
type Block struct {
	ID   string      `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// ChecklistBlockData represents checklist block data
type ChecklistBlockData struct {
	Items []ChecklistItem `json:"items"`
}

// ChecklistItem represents a single checklist item
type ChecklistItem struct {
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

// HeaderBlockData represents header block data
type HeaderBlockData struct {
	Text  string `json:"text"`
	Level int    `json:"level"`
}

// ImageBlockData represents the data structure for an image block
type ImageBlockData struct {
	File           ImageFileData `json:"file"`
	Caption        string        `json:"caption"`
	WithBorder     bool          `json:"withBorder"`
	WithBackground bool          `json:"withBackground"`
	Stretched      bool          `json:"stretched"`
}

// ImageFileData represents the file metadata for an image block
type ImageFileData struct {
	URL string `json:"url"`
}

// LinkToolBlockData represents the data structure for a linkTool block
type LinkToolBlockData struct {
	Link string       `json:"link"`
	Meta LinkToolMeta `json:"meta"`
}

// LinkToolMeta represents the metadata for a linkTool block
type LinkToolMeta struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Image       ImageFileData `json:"image"`
}

// ListBlockData represents list block data
type ListBlockData struct {
	Type  string          `json:"type"`
	Items []ListItemBlock `json:"items"`
}

// ListItemBlock represents a list item (may contain nested items)
type ListItemBlock struct {
	Text     string          `json:"text"`
	Children []ListItemBlock `json:"children"`
}

// ParagraphBlockData represents paragraph block data
type ParagraphBlockData struct {
	Text string `json:"text"`
}

// QuoteBlockData represents the data structure for a quote block
type QuoteBlockData struct {
	Text      string `json:"text"`
	Caption   string `json:"caption"`
	Alignment string `json:"alignment"`
}

// MarkdownParser initializes a reusable Goldmark instance
var MarkdownParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()), // Allow unsafe HTML
)

// ConvertFrontmatterToYAML converts frontmatter map to YAML string
func ConvertFrontmatterToYAML(frontmatter map[string]interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to convert frontmatter to YAML: %w", err)
	}
	return "---\n" + string(yamlBytes) + "---\n", nil
}

// ConvertBlocksToMarkdown processes blocks and converts them to Markdown
func ConvertBlocksToMarkdown(blocks []map[string]interface{}, maxDepth int) (string, error) {
	var markdownBuilder strings.Builder

	for _, block := range blocks {
		blockType := block["type"].(string)
		data := block["data"].(map[string]interface{})

		switch blockType {
		case "header":
			text := data["text"].(string)
			level := int(data["level"].(float64))
			markdownBuilder.WriteString(strings.Repeat("#", level) + " " + text + "\n\n")
		case "paragraph":
			htmlContent := data["text"].(string) // Paragraph data is stored as HTML
			markdownContent, err := htmlToMarkdown(htmlContent)
			if err != nil {
				return "", err
			}
			markdownBuilder.WriteString(markdownContent + "\n\n")
		case "list":
			items := data["items"].([]interface{})
			listType := data["type"].(string)
			appendListMarkdown(&markdownBuilder, items, listType, 0, maxDepth)
		case "checklist":
			items := data["items"].([]interface{})
			appendChecklistMarkdown(&markdownBuilder, items, 0, maxDepth)
		case "quote":
			text := data["text"].(string)
			caption := data["caption"].(string)
			markdownBuilder.WriteString("> " + text + "\n")
			if caption != "" {
				markdownBuilder.WriteString("> — " + caption + "\n")
			}
			markdownBuilder.WriteString("\n")
		case "code":
			code := data["code"].(string)
			language, ok := data["language"].(string)
			if ok && language != "" {
				markdownBuilder.WriteString("```" + language + "\n")
			} else {
				markdownBuilder.WriteString("```\n")
			}
			markdownBuilder.WriteString(code + "\n```\n\n")
		case "image":
			url := data["file"].(map[string]interface{})["url"].(string)
			caption := data["caption"].(string)
			markdownBuilder.WriteString("![")
			if caption != "" {
				markdownBuilder.WriteString(caption)
			}
			markdownBuilder.WriteString("](" + url + ")\n\n")
		default:
			return "", fmt.Errorf("unsupported block type: %s", blockType)
		}
	}

	return markdownBuilder.String(), nil
}

func htmlToMarkdown(htmlContent string) (string, error) {
	// Parse HTML to Markdown using Goldmark
	source := []byte(htmlContent)
	reader := text.NewReader(source)

	parsed := MarkdownParser.Parser().Parse(reader)
	var markdownBuffer bytes.Buffer
	if err := MarkdownParser.Renderer().Render(&markdownBuffer, source, parsed); err != nil {
		return "", fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	return markdownBuffer.String(), nil
}

func appendListMarkdown(builder *strings.Builder, items []interface{}, listType string, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}

	for _, item := range items {
		itemData := item.(map[string]interface{})
		text := itemData["text"].(string)

		// Indentation for nested lists
		prefix := strings.Repeat("  ", depth)
		if listType == "ordered" {
			builder.WriteString(prefix + "1. " + text + "\n")
		} else {
			builder.WriteString(prefix + "- " + text + "\n")
		}

		// Recursively process children (either nested lists or checklists)
		if children, ok := itemData["children"].([]interface{}); ok && len(children) > 0 {
			appendListMarkdown(builder, children, listType, depth+1, maxDepth)
		}
		if nestedChecklist, ok := itemData["checklist"].([]interface{}); ok && len(nestedChecklist) > 0 {
			appendChecklistMarkdown(builder, nestedChecklist, depth+1, maxDepth)
		}
	}
	builder.WriteString("\n")
}

func appendChecklistMarkdown(builder *strings.Builder, items []interface{}, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}

	for _, item := range items {
		itemData := item.(map[string]interface{})
		text := itemData["text"].(string)
		checked := itemData["checked"].(bool)

		// Indentation for nested checklists
		prefix := strings.Repeat("  ", depth)
		if checked {
			builder.WriteString(prefix + "- [x] " + text + "\n")
		} else {
			builder.WriteString(prefix + "- [ ] " + text + "\n")
		}

		// Recursively process children (either nested lists or checklists)
		if nestedList, ok := itemData["list"].([]interface{}); ok && len(nestedList) > 0 {
			appendListMarkdown(builder, nestedList, "unordered", depth+1, maxDepth)
		}
		if nestedChecklist, ok := itemData["checklist"].([]interface{}); ok && len(nestedChecklist) > 0 {
			appendChecklistMarkdown(builder, nestedChecklist, depth+1, maxDepth)
		}
	}
	builder.WriteString("\n")
}
