package blocks

import (
	"bytes"
	"fmt"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	elem "github.com/chasefleming/elem-go"
	"github.com/chasefleming/elem-go/attrs"
)

// Block represents a structured block of content.
type Block struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// MarkdownOptions centralizes all options for markdown conversion
type MarkdownOptions struct {
	WithImageBorder     bool
	WithImageBackground bool
	WithImageStretched  bool
	// Add other block-specific options here as needed
}

// Functional option for configuring markdown parsing
type MarkdownOption func(*MarkdownOptions)

// DefaultMarkdownOptions returns default settings
func DefaultMarkdownOptions() *MarkdownOptions {
	return &MarkdownOptions{
		WithImageBorder:     false,
		WithImageBackground: false,
		WithImageStretched:  false,
	}
}

// Functions to configure image-specific options
func WithImageBorder(withBorder bool) MarkdownOption {
	return func(opts *MarkdownOptions) {
		opts.WithImageBorder = withBorder
	}
}

func WithImageBackground(withBackground bool) MarkdownOption {
	return func(opts *MarkdownOptions) {
		opts.WithImageBackground = withBackground
	}
}

func WithImageStretched(stretched bool) MarkdownOption {
	return func(opts *MarkdownOptions) {
		opts.WithImageStretched = stretched
	}
}

// ParseBlocksToMarkdown converts a list of Block objects into a markdown string.
func ParseBlocksToMarkdown(blocks []Block, options ...MarkdownOption) (string, error) {
	mdOptions := DefaultMarkdownOptions()
	for _, opt := range options {
		opt(mdOptions)
	}
	var buffer bytes.Buffer

	for _, block := range blocks {
		handler, ok := blockHandlers[block.Type]
		if !ok {
			return "", fmt.Errorf("no handler found for block type %s", block.Type)
		}

		err := handler(&buffer, block.Data, mdOptions)
		if err != nil {
			return "", fmt.Errorf("error handling block type %s: %w", block.Type, err)
		}
	}

	return strings.TrimSpace(buffer.String()), nil
}

// blockHandlers maps block types to their corresponding handler functions.
var blockHandlers = map[string]func(*bytes.Buffer, map[string]interface{}, *MarkdownOptions) error{
	"image":     handleImage,
	"paragraph": handleParagraph,
	"header":    handleHeader,
	"list":      handleList,
	"code":      handleCode,
	"quote":     handleQuote,
	"table":     handleTable,
	"alert":     handleAlert,
	"delimiter": handleDelimiter,
}

// Handlers for individual block types
func handleParagraph(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	if text, ok := data["text"].(string); ok {
		markdownText := convertToMarkdown(strings.TrimSpace(text))
		buffer.WriteString(markdownText + "\n\n")
	}
	return nil
}

func handleHeader(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	text, textOk := data["text"].(string)
	level, levelOk := data["level"].(float64)
	if !textOk {
		return fmt.Errorf("missing or invalid 'text' in heading block")
	}
	if !levelOk {
		return fmt.Errorf("missing or invalid 'level' in heading block")
	}

	markdownText := convertToMarkdown(text)
	buffer.WriteString(strings.Repeat("#", int(level)) + " " + markdownText + "\n\n")
	return nil
}

func handleList(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	// Extract list properties
	items, ok := data["items"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid or missing 'items' in list data")
	}

	style := "unordered"
	if s, exists := data["style"].(string); exists {
		style = s
	}

	meta, _ := data["meta"].(map[string]interface{})

	// Process the list items based on style
	if err := processList(buffer, items, style, meta, 0); err != nil {
		return fmt.Errorf("error processing list: %w", err)
	}

	buffer.WriteString("\n")
	return nil
}

// Helper to process a list
func processList(buffer *bytes.Buffer, items []interface{}, style string, meta map[string]interface{}, depth int) error {
	for _, item := range items {
		// Parse item based on its type
		switch v := item.(type) {
		case map[string]interface{}:
			if err := processListItem(buffer, v, style, meta, depth); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected item type in list: %T", v)
		}
	}
	return nil
}

// Helper to process a single list item
func processListItem(buffer *bytes.Buffer, item map[string]interface{}, style string, meta map[string]interface{}, depth int) error {
	prefix := strings.Repeat("    ", depth) // Indentation for nested lists

	// Extract content and metadata
	content, _ := item["content"].(string)
	markdownContent := convertToMarkdown(content)
	itemMeta, _ := item["meta"].(map[string]interface{})
	nestedItems, _ := item["items"].([]interface{})

	switch style {
	case "checklist":
		// Handle checklist
		checked := false
		if itemMeta != nil {
			checked, _ = itemMeta["checked"].(bool)
		}
		box := "[ ]"
		if checked {
			box = "[x]"
		}
		buffer.WriteString(fmt.Sprintf("%s%s %s\n", prefix, box, markdownContent))

	case "ordered":
		// Handle ordered list
		start := 1
		counterType := "numeric"
		if meta != nil {
			if s, ok := meta["start"].(float64); ok {
				start = int(s)
			}
			if c, ok := meta["counterType"].(string); ok {
				counterType = c
			}
		}
		buffer.WriteString(fmt.Sprintf("%s%s%s\n", prefix, getOrderedPrefix(start, counterType, depth), markdownContent))

	default:
		// Default to unordered list
		buffer.WriteString(fmt.Sprintf("%s- %s\n", prefix, markdownContent))
	}

	// Process nested items
	if len(nestedItems) > 0 {
		if err := processList(buffer, nestedItems, style, meta, depth+1); err != nil {
			return fmt.Errorf("error processing nested list: %w", err)
		}
	}

	return nil
}

func getOrderedPrefix(index int, counterType string, depth int) string {
	adjustedIndex := index
	for i := 0; i < depth; i++ {
		adjustedIndex += index - 1 // Adjust index for nested levels
	}
	switch counterType {
	case "lower-roman":
		return fmt.Sprintf("%s. ", toRoman(adjustedIndex, false))
	case "upper-roman":
		return fmt.Sprintf("%s. ", toRoman(adjustedIndex, true))
	case "lower-alpha":
		return fmt.Sprintf("%c. ", 'a'+adjustedIndex-1)
	case "upper-alpha":
		return fmt.Sprintf("%c. ", 'A'+adjustedIndex-1)
	default:
		return fmt.Sprintf("%d. ", adjustedIndex)
	}
}

// Utility function to convert an integer to a Roman numeral
func toRoman(num int, uppercase bool) string {
	var romanMap = []struct {
		Value  int
		Symbol string
	}{
		{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
		{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
	}

	var result strings.Builder
	for _, entry := range romanMap {
		for num >= entry.Value {
			num -= entry.Value
			if uppercase {
				result.WriteString(strings.ToUpper(entry.Symbol))
			} else {
				result.WriteString(strings.ToLower(entry.Symbol))
			}
		}
	}
	return result.String()
}
func handleCode(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	if code, ok := data["code"].(string); ok {
		if language, exists := data["language"].(string); exists && language != "" {
			// Use fenced code block with language
			buffer.WriteString("```" + language + "\n" + code + "\n```\n\n")
		} else {
			// Use indented code block (4 spaces for each line)
			lines := strings.Split(code, "\n")
			for _, line := range lines {
				buffer.WriteString("    " + line + "\n")
			}
			buffer.WriteString("\n")
		}
	}
	return nil
}

func handleQuote(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	if text, ok := data["text"].(string); ok {
		markdownText := convertToMarkdown(text)
		buffer.WriteString("> " + markdownText + "\n")
		if caption, exists := data["caption"].(string); exists && caption != "" {
			captionMarkdown := convertToMarkdown(caption)
			buffer.WriteString("> \n> -- <caption>" + captionMarkdown + "</caption>\n")
		}
		buffer.WriteString("\n")
	}
	return nil
}

func handleTable(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	if rows, ok := data["rows"].([]interface{}); ok {
		withHeadings := false
		if val, exists := data["withHeadings"].(bool); exists {
			withHeadings = val
		}

		for i, row := range rows {
			if cols, ok := row.([]interface{}); ok {
				var line []string
				for _, col := range cols {
					if cell, ok := col.(string); ok {
						line = append(line, convertToMarkdown(cell))
					}
				}

				if withHeadings && i == 0 {
					buffer.WriteString("| " + strings.Join(line, " | ") + " |\n")
					buffer.WriteString("|" + strings.Repeat(" --- |", len(line)) + "\n")
				} else {
					buffer.WriteString("| " + strings.Join(line, " | ") + " |\n")
				}
			}
		}
		buffer.WriteString("\n")
	}
	return nil
}

func handleDelimiter(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	buffer.WriteString("---\n\n")
	return nil
}

func convertToMarkdown(htmlText string) string {
	markdown, err := htmltomarkdown.ConvertString(
		htmlText,
	)
	if err != nil {
		println("Error converting to markdown:", err)
	}
	return markdown
}

func handleImage(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	// Extract required fields
	url, caption, err := extractImageData(data)
	if err != nil {
		return err
	}

	// Determine image classes
	classes := generateImageClasses(opts)

	// Render HTML if caption is present, otherwise render Markdown
	if caption != "" {
		renderImageAsHTML(buffer, url, caption, classes)
	} else {
		renderImageAsMarkdown(buffer, url, caption, opts)
	}

	return nil
}

// Helper to extract image data
func extractImageData(data map[string]interface{}) (string, string, error) {
	file, ok := data["file"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("missing or invalid 'file' field in image block")
	}

	url, ok := file["url"].(string)
	if !ok {
		return "", "", fmt.Errorf("missing or invalid 'url' in image file")
	}

	caption, _ := data["caption"].(string)
	return url, caption, nil
}

// Helper to generate image classes based on options
func generateImageClasses(opts *MarkdownOptions) string {
	var classes []string
	if opts.WithImageBorder {
		classes = append(classes, "with-border")
	}
	if opts.WithImageBackground {
		classes = append(classes, "with-background")
	}
	if opts.WithImageStretched {
		classes = append(classes, "stretched")
	}
	return strings.Join(classes, " ")
}

// Helper to render image as HTML
func renderImageAsHTML(buffer *bytes.Buffer, url, caption, classes string) {
	figure := elem.Figure(nil,
		elem.Img(attrs.Props{
			attrs.Src:   url,
			attrs.Alt:   caption,
			attrs.Class: classes,
		}),
		elem.FigCaption(nil, elem.Text(caption)),
	)
	buffer.WriteString(figure.Render() + "\n\n")
}

// Helper to render image as Markdown
func renderImageAsMarkdown(buffer *bytes.Buffer, url, caption string, opts *MarkdownOptions) {
	buffer.WriteString(fmt.Sprintf("![%s](%s)\n", convertToMarkdown(caption), url))

	// Append options as a comment
	buffer.WriteString("<!-- Options: ")
	if opts.WithImageBorder {
		buffer.WriteString("border, ")
	}
	if opts.WithImageBackground {
		buffer.WriteString("background, ")
	}
	if opts.WithImageStretched {
		buffer.WriteString("stretched, ")
	}
	buffer.WriteString("-->\n\n")
}

func handleAlert(buffer *bytes.Buffer, data map[string]interface{}, opts *MarkdownOptions) error {
	// Supported types and their corresponding alert headers
	alertTypes := map[string]string{
		"primary":   "[!NOTE]",
		"success":   "[!TIP]",
		"secondary": "[!IMPORTANT]",
		"warning":   "[!WARNING]",
		"danger":    "[!CAUTION]",
	}

	// Extract data
	alertType, _ := data["type"].(string)
	align, _ := data["align"].(string)
	text, ok := data["text"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'text' field in alert block")
	}

	// Determine the alert header
	header, exists := alertTypes[alertType]
	if !exists {
		return fmt.Errorf("unsupported alert type: %s", alertType)
	}

	// Process text
	var processedText string
	if align == "center" {
		// Use raw HTML text for center alignment
		processedText = text
	} else {
		// Convert text to Markdown for other alignments
		processedText = convertToMarkdown(text)
	}

	// Split the text into lines
	lines := strings.Split(processedText, "\n")

	// Write the alert block
	buffer.WriteString("> " + header + "\n")
	if align == "center" {
		buffer.WriteString("> <div align='center'>\n")
		for _, line := range lines {
			buffer.WriteString("> " + line + "\n")
		}
		buffer.WriteString("> </div>\n")
	} else {
		for _, line := range lines {
			buffer.WriteString("> " + line + "\n")
		}
	}
	buffer.WriteString("\n")

	return nil
}
