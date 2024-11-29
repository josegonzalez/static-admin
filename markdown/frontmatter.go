package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type FrontmatterField struct {
	Name             string    `json:"name"`
	StringValue      string    `json:"stringValue"`
	BoolValue        bool      `json:"boolValue"`
	NumberValue      float64   `json:"numberValue"`
	DateTimeValue    time.Time `json:"dateTimeValue"`
	StringSliceValue []string  `json:"stringSliceValue"`
	Type             string    `json:"type"`
}

// ExtractFrontMatter parses the frontmatter and returns it as a map, along with the remaining Markdown content.
func ExtractFrontMatter(content []byte) ([]FrontmatterField, string, error) {
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

	fields, err := parseFrontmatterFields(frontMatter)
	if err != nil {
		return nil, "", err
	}

	return fields, markdownContent.String(), nil
}

func parseFrontmatterFields(frontmatter map[string]interface{}) ([]FrontmatterField, error) {
	fields := map[string]FrontmatterField{}
	for key, value := range frontmatter {
		field := FrontmatterField{
			Name:             key,
			StringValue:      "",
			BoolValue:        false,
			NumberValue:      0,
			StringSliceValue: []string{},
			Type:             reflect.TypeOf(value).String(),
		}

		if field.Name == "date" {
			date, err := time.Parse("2006-01-02 15:04", value.(string))
			if err != nil {
				return nil, fmt.Errorf("failed to parse date: %w", err)
			}
			field.DateTimeValue = date
			field.Type = "dateTime"
		} else {
			switch v := value.(type) {
			case string:
				field.StringValue = v
			case bool:
				field.BoolValue = v
			case float64:
				field.NumberValue = v
			case []interface{}:
				// Convert []interface{} to []string
				sliceParsed := false
				for _, item := range v {
					if str, ok := item.(string); ok {
						sliceParsed = true
						field.Type = "stringSlice"
						field.StringSliceValue = append(field.StringSliceValue, str)
					}
				}

				if !sliceParsed {
					return nil, fmt.Errorf("unsupported slice type: %v", v)
				}
			}
		}
		fields[key] = field
	}

	// sort the fields in a specific order:
	// title, date, description, categories, tags
	// then everything else
	// finally all boolean fields
	orderedFields := make([]FrontmatterField, 0)

	// Add fields in specific order if they exist
	priorityFields := []string{"title", "date", "description", "category", "categories", "tags"}
	for _, key := range priorityFields {
		if field, exists := fields[key]; exists {
			orderedFields = append(orderedFields, field)
			delete(fields, key)
		}
	}

	// Add remaining non-boolean fields in alphabetical order
	remainingFields := make([]string, 0)
	for key, field := range fields {
		if field.Type != "bool" {
			remainingFields = append(remainingFields, key)
		}
	}
	sort.Strings(remainingFields)
	for _, key := range remainingFields {
		orderedFields = append(orderedFields, fields[key])
		delete(fields, key)
	}

	// Add boolean fields last, in alphabetical order
	boolFields := make([]string, 0)
	for key := range fields {
		boolFields = append(boolFields, key)
	}
	sort.Strings(boolFields)
	for _, key := range boolFields {
		orderedFields = append(orderedFields, fields[key])
	}

	return orderedFields, nil
}

// FrontmatterFieldToYaml converts a slice of FrontmatterField to a YAML string
func FrontmatterFieldToYaml(fields []FrontmatterField) (string, error) {
	// Convert fields to a map for yaml marshaling
	frontmatter := make(map[string]interface{})
	for _, field := range fields {
		var value interface{}
		switch field.Type {
		case "string":
			value = field.StringValue
		case "bool":
			value = field.BoolValue
		case "number":
			value = field.NumberValue
		case "dateTime":
			value = field.DateTimeValue.Format("2006-01-02 15:04")
		case "stringSlice":
			value = field.StringSliceValue
		default:
			return "", fmt.Errorf("unknown field type: %s", field.Type)
		}
		frontmatter[field.Name] = value
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Format with delimiters
	return fmt.Sprintf("---\n%s---\n", string(yamlData)), nil
}
