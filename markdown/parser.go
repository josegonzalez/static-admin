package markdown

import (
	"bytes"
	"net/http"
	"regexp"
	"static-admin/blocks"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type ParseOption func(*ParseConfig)

type ParseConfig struct {
	MaxDepth            int
	QuoteCaptionAlign   string
	TableStretched      bool
	ImageStretched      bool
	ImageWithBorder     bool
	ImageWithBackground bool
}

func WithMaxDepth(depth int) ParseOption {
	return func(cfg *ParseConfig) {
		cfg.MaxDepth = depth
	}
}

func WithQuoteCaptionAlign(align string) ParseOption {
	return func(cfg *ParseConfig) {
		cfg.QuoteCaptionAlign = align
	}
}

func WithTableStretched(stretched bool) ParseOption {
	return func(cfg *ParseConfig) {
		cfg.TableStretched = stretched
	}
}

func WithImageStretched(stretched bool) ParseOption {
	return func(cfg *ParseConfig) {
		cfg.ImageStretched = stretched
	}
}

func WithImageWithBorder(border bool) ParseOption {
	return func(cfg *ParseConfig) {
		cfg.ImageWithBorder = border
	}
}

func WithImageWithBackground(background bool) ParseOption {
	return func(cfg *ParseConfig) {
		cfg.ImageWithBackground = background
	}
}

// MarkdownParser initializes a reusable Goldmark instance
var MarkdownParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()), // Allow unsafe HTML
)

// ParseMarkdownToBlocks converts Markdown content into structured blocks.
func ParseMarkdownToBlocks(markdown string, opts ...ParseOption) ([]blocks.Block, error) {
	config := &ParseConfig{
		MaxDepth:            5,      // Default max depth for lists
		QuoteCaptionAlign:   "left", // Default alignment for quote captions
		TableStretched:      false,  // Default for table stretching
		ImageStretched:      false,  // Default for image stretching
		ImageWithBorder:     false,  // Default for image border
		ImageWithBackground: false,  // Default for image background
	}

	for _, opt := range opts {
		opt(config)
	}

	var blocks []blocks.Block

	reader := text.NewReader([]byte(markdown))
	document := MarkdownParser.Parser().Parse(reader)

	// Walk the AST to extract blocks
	err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		block, walkStatus, handled := processNode(node, markdown, config)
		if handled {
			blocks = append(blocks, block)
		}

		return walkStatus, nil
	})

	if err != nil {
		return nil, err
	}

	return blocks, nil
}

// processNode routes the processing of an AST node to the appropriate handler.
func processNode(node ast.Node, markdown string, config *ParseConfig) (blocks.Block, ast.WalkStatus, bool) {
	switch n := node.(type) {
	case *ast.Heading:
		return handleHeading(n, markdown), ast.WalkContinue, true
	case *ast.ThematicBreak:
		return handleDelimiter(), ast.WalkContinue, true
	case *ast.List:
		return handleList(n, markdown, config), ast.WalkSkipChildren, true
	case *ast.CodeBlock:
		return handleCodeBlock(n, markdown), ast.WalkContinue, true
	case *ast.FencedCodeBlock:
		return handleFencedCodeBlock(n, markdown), ast.WalkContinue, true
	case *ast.Blockquote:
		return handleBlockquote(n, markdown, config), ast.WalkSkipChildren, true
	case *ast.HTMLBlock:
		return handleHTMLBlock(n, markdown, config), ast.WalkContinue, true
	case *east.Table:
		return handleTable(n, markdown, config), ast.WalkContinue, true
	case *ast.Paragraph:
		return handleParagraph(n, markdown, config), ast.WalkContinue, true
	default:
		return blocks.Block{}, ast.WalkContinue, false
	}
}

func handleHeading(node *ast.Heading, markdown string) blocks.Block {
	nodeText := extractNodeText(node, markdown)

	text, err := markdownToHTML(nodeText)
	if err != nil {
		text = nodeText
	}

	return blocks.Block{
		Type: "header",
		Data: map[string]interface{}{
			"text":  text,
			"level": node.Level,
		},
	}
}

func handleDelimiter() blocks.Block {
	return blocks.Block{
		Type: "delimiter",
		Data: map[string]interface{}{},
	}
}

func handleList(node *ast.List, markdown string, config *ParseConfig) blocks.Block {
	style := "unordered"
	if node.IsOrdered() {
		style = "ordered"
	}

	items, isChecklist := extractListItems(node, markdown, 0, config.MaxDepth)
	meta := map[string]interface{}{}
	if style == "ordered" {
		meta["start"] = node.Start
		meta["counterType"] = detectCounterType(node)
	}

	if isChecklist {
		style = "checklist"
		delete(meta, "start")
		delete(meta, "counterType")
	}

	return blocks.Block{
		Type: "list",
		Data: map[string]interface{}{
			"style": style,
			"meta":  meta,
			"items": items,
		},
	}
}

func extractListItems(list *ast.List, markdown string, depth, maxDepth int) ([]map[string]interface{}, bool) {
	if depth >= maxDepth {
		return nil, false
	}

	var items []map[string]interface{}

	isChecklist := false
	checkboxPrefixes := []string{"[ ] ", "[x] ", "[X] "}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		if listItem, ok := item.(*ast.ListItem); ok {
			text := extractNodeText(listItem, markdown)

			isChecked := false
			trimmedText := strings.TrimSpace(text)
			for _, prefix := range checkboxPrefixes {
				if strings.HasPrefix(trimmedText, prefix) {
					isChecklist = true
					isChecked = strings.HasPrefix(trimmedText, "[x] ") || strings.HasPrefix(trimmedText, "[X] ")
					text = strings.TrimPrefix(text, prefix)
					break
				}
			}

			children := []map[string]interface{}{}
			for child := listItem.FirstChild(); child != nil; child = child.NextSibling() {
				if nestedList, ok := child.(*ast.List); ok {
					newChildren, _ := extractListItems(nestedList, markdown, depth+1, maxDepth)
					children = append(children, newChildren...)
				}
			}

			htmlText, err := markdownToHTML(text)
			if err == nil {
				text = htmlText
			}

			itemData := map[string]interface{}{
				"content": text,
				"items":   children,
			}

			if isChecklist {
				itemData["checked"] = isChecked
			}

			items = append(items, itemData)
		}
	}

	return items, isChecklist
}

func detectCounterType(list *ast.List) string {
	switch list.Marker {
	case ')':
		return "upper-roman"
	case '.':
		return "decimal"
	default:
		return "unknown"
	}
}

func handleCodeBlock(node *ast.CodeBlock, markdown string) blocks.Block {
	code := extractNodeText(node, markdown)
	return blocks.Block{
		Type: "code",
		Data: map[string]interface{}{
			"code": code,
		},
	}
}

func handleFencedCodeBlock(node *ast.FencedCodeBlock, markdown string) blocks.Block {
	code := extractNodeText(node, markdown)
	language := string(node.Language([]byte(markdown)))
	return blocks.Block{
		Type: "code",
		Data: map[string]interface{}{
			"code":     code,
			"language": language,
		},
	}
}

func handleBlockquote(node *ast.Blockquote, markdown string, config *ParseConfig) blocks.Block {
	content := extractNodeText(node, markdown)
	admonitionRegex := regexp.MustCompile(`^\[\!([A-Z]*)\]\n([\S\s]*)`)
	matches := admonitionRegex.FindStringSubmatch(content)

	if len(matches) > 0 {
		typeMap := map[string]string{
			"CAUTION":   "danger",
			"WARNING":   "warning",
			"IMPORTANT": "secondary",
			"TIP":       "success",
			"NOTE":      "primary",
		}

		alertType, ok := typeMap[matches[1]]
		if !ok {
			alertType = "info"
		}

		htmlContent, err := markdownToHTML(matches[2])
		if err != nil {
			htmlContent = matches[2]
		}
		return blocks.Block{
			Type: "alert",
			Data: map[string]interface{}{
				"type":    alertType,
				"align":   config.QuoteCaptionAlign,
				"message": htmlContent,
			},
		}
	}

	// Handle normal blockquote with citation on the last line
	lines := strings.Split(content, "\n")
	var quoteText string
	var citation string

	quoteText = content
	if len(lines) > 1 {
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		if strings.HasPrefix(lastLine, "--") {
			// Extract citation from the last line
			citation = strings.TrimSpace(strings.TrimPrefix(lastLine, "--"))
			quoteText = strings.Join(lines[:len(lines)-1], "\n")
		}
	}

	// Convert quote text and citation to HTML
	htmlQuoteText, err := markdownToHTML(quoteText)
	if err != nil {
		htmlQuoteText = quoteText
	}
	htmlCitation := ""
	if citation != "" {
		htmlCitation, err = markdownToHTML(citation)
		if err != nil {
			htmlCitation = citation
		}
	}

	return blocks.Block{
		Type: "quote",
		Data: map[string]interface{}{
			"text":      htmlQuoteText,
			"caption":   htmlCitation,
			"alignment": config.QuoteCaptionAlign,
		},
	}
}

func handleHTMLBlock(node *ast.HTMLBlock, markdown string, config *ParseConfig) blocks.Block {
	htmlContent := extractNodeText(node, markdown)

	// Handle `<figure>` blocks
	if figureBlock := handleFigureBlock(node, markdown, config); figureBlock != nil {
		return *figureBlock
	}

	// Handle arbitrary HTML blocks
	return blocks.Block{
		Type: "raw",
		Data: map[string]interface{}{
			"html": htmlContent,
		},
	}
}

func handleFigureBlock(node ast.Node, markdown string, config *ParseConfig) *blocks.Block {
	htmlContent := extractNodeText(node, markdown)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	if len(doc.Nodes) != 1 {
		return nil
	}

	// find the figure block
	figures := doc.Find("figure")
	if figures.Length() != 1 {
		return nil
	}

	// get the image url and caption
	imageURL, _ := figures.Find("img").Attr("src")
	caption := figures.Find("figcaption").Text()

	return &blocks.Block{
		Type: "image",
		Data: map[string]interface{}{
			"file": map[string]interface{}{
				"url": imageURL,
			},
			"caption":        caption,
			"stretched":      config.ImageStretched,
			"withBackground": config.ImageWithBackground,
			"withBorder":     config.ImageWithBorder,
		},
	}
}

func handleTable(node *east.Table, markdown string, config *ParseConfig) blocks.Block {
	var content [][]string
	withHeadings := false

	// Iterate through table rows
	for row := node.FirstChild(); row != nil; row = row.NextSibling() {
		switch row.(type) {
		case *east.TableHeader:
			withHeadings = true
			content = append(content, parseTableRow(row, markdown))
		case *east.TableRow:
			content = append(content, parseTableRow(row, markdown))
		}
	}

	return blocks.Block{
		Type: "table",
		Data: map[string]interface{}{
			"withHeadings": withHeadings,
			"stretched":    config.TableStretched,
			"content":      content,
		},
	}
}

func parseTableRow(row ast.Node, markdown string) []string {
	var rowContent []string

	// Iterate through table cells
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if tableCell, ok := cell.(*east.TableCell); ok {
			text := extractNodeText(tableCell, markdown)
			rowContent = append(rowContent, strings.TrimSpace(text))
		}
	}

	return rowContent
}

func handleParagraph(node *ast.Paragraph, markdown string, config *ParseConfig) blocks.Block {
	nodeText := extractNodeText(node, markdown)
	text, err := markdownToHTML(nodeText)
	if err != nil {
		text = nodeText
	}

	// Handle `<img>` blocks
	if imageBlock := handleImageBlock(node, markdown, text, config); imageBlock != nil {
		return *imageBlock
	}

	// Check if paragraph is a standalone link
	linkRegex := regexp.MustCompile(`^https?:\/\/[^\s]+$`)
	if linkRegex.MatchString(text) {
		return blocks.Block{
			Type: "linkTool",
			Data: map[string]interface{}{
				"link": text,
				"meta": FetchLinkMetadata(text), // Fetch metadata for the link
			},
		}
	}

	return blocks.Block{
		Type: "paragraph",
		Data: map[string]interface{}{
			"text": text,
		},
	}
}

func handleImageBlock(_ ast.Node, _ string, htmlContent string, config *ParseConfig) *blocks.Block {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	if len(doc.Nodes) != 1 {
		return nil
	}

	// find the image block
	images := doc.Find("img")
	if images.Length() != 1 {
		return nil
	}

	// get the image url and caption
	imageURL, _ := images.Attr("src")
	caption := images.AttrOr("alt", "")

	return &blocks.Block{
		Type: "image",
		Data: map[string]interface{}{
			"caption": caption,
			"file": map[string]interface{}{
				"url": imageURL,
			},
			"stretched":      config.ImageStretched,
			"withBackground": config.ImageWithBackground,
			"withBorder":     config.ImageWithBorder,
		},
	}
}

func markdownToHTML(markdown string) (string, error) {
	var buf bytes.Buffer
	err := MarkdownParser.Convert([]byte(markdown), &buf)
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(buf.String())
	if strings.HasPrefix(content, "<p>") && strings.HasSuffix(content, "</p>") {
		// Remove surrounding <p> tags
		content = content[3 : len(content)-4]
	}

	// Add inline-code class to all code elements
	if strings.Contains(content, "<code>") {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err == nil {
			doc.Find("code").AddClass("inline-code")
			if html, err := doc.Html(); err == nil {
				content = html
			}
		}
	}

	return content, nil
}

func extractNodeText(node ast.Node, fullMarkdownDocument string) string {
	var buf bytes.Buffer
	if node.Type() == ast.TypeBlock {
		if node.Lines() == nil {
			return ""
		}

		lines := node.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value([]byte(fullMarkdownDocument)))
		}
		buf.WriteString("\n")
	}

	if node.Kind() != ast.KindList {
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			buf.Write([]byte(extractNodeText(c, fullMarkdownDocument)))
			buf.WriteString("\n")
		}
	}

	return strings.Trim(buf.String(), "\n")
}

// FetchLinkMetadata fetches metadata for a given URL, including the title, description, and an image.
func FetchLinkMetadata(link string) map[string]interface{} {
	resp, err := http.Get(link)
	if err != nil {
		return map[string]interface{}{
			"title":       "",
			"description": "",
			"image":       map[string]string{"url": ""},
			"error":       "failed to fetch URL",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"title":       "",
			"description": "",
			"image":       map[string]string{"url": ""},
			"error":       "non-200 status code",
		}
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"title":       "",
			"description": "",
			"image":       map[string]string{"url": ""},
			"error":       "failed to parse document",
		}
	}

	// Extract metadata
	title := doc.Find("title").First().Text()
	description, _ := doc.Find(`meta[name="description"]`).Attr("content")
	image, _ := doc.Find(`meta[property="og:image"]`).Attr("content")

	// Clean up data
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	image = strings.TrimSpace(image)

	return map[string]interface{}{
		"title":       title,
		"description": description,
		"image":       map[string]string{"url": image},
	}
}
