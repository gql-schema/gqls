package convert

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"text/template"
)

// getPrimaryLabel retrieves the first label from a label set.
func getPrimaryLabel(labelSet []string) (string, error) {
	if len(labelSet) == 0 {
		return "", errors.New("no labelSet found")
	}
	return labelSet[0], nil
}

// isVersionSupported checks if the current Neo4j version satisfies the minimum requirement.
func isVersionSupported(current, minimum string) (bool, error) {
	if current == "" || current == "latest" {
		return true, nil
	}

	currentParts := strings.Split(current, ".")
	minimumParts := strings.Split(minimum, ".")

	for i := 0; i < len(currentParts) && i < len(minimumParts); i++ {
		currentStr := currentParts[i]
		currentPart, err := strconv.Atoi(currentStr)
		if err != nil {
			return false, fmt.Errorf("invalid current version part %q: %w", currentStr, err)
		}

		minimumStr := minimumParts[i]
		minimumPart, err := strconv.Atoi(minimumStr)
		if err != nil {
			return false, fmt.Errorf("invalid minimum version part %q: %w", minimumStr, err)
		}

		if currentPart < minimumPart {
			return false, nil
		} else if currentPart > minimumPart {
			return true, nil
		}
	}

	return len(currentParts) >= len(minimumParts), nil
}

// mustFormatName formats the name of a constraint or trigger.
func mustFormatName(prefix string, components []string) string {
	if len(components) == 0 {
		panic("mustFormatName requires at least one component")
	}

	return strings.ToLower(
		fmt.Sprintf(
			"%s_%s",
			prefix,
			strings.Join(components, "_"),
		),
	)
}

// getPattern generates a Cypher pattern based on the entity type, variable, and label.
func getPattern(entityType EntityType, variable, label string) string {
	switch entityType {
	case NodeEntityType:
		return fmt.Sprintf("(%s:%s)", variable, label)
	case EdgeEntityType:
		return fmt.Sprintf("()-[%s:%s]->()", variable, label)
	}
	panic("unreachable")
}

// entityTypeToVariables maps an entity type to its corresponding variables.
func entityTypeToVariables(entityType EntityType) (string, string) {
	switch entityType {
	case NodeEntityType:
		return "$createdNodes", "$assignedNodeProperties"
	case EdgeEntityType:
		return "$createdRelationships", "$assignedRelationshipProperties"
	}
	panic("unreachable")
}

// renderTemplate renders a template with the given data.
//
// It reads the template file from the embedded filesystem, parses it, and
// executes it with the provided data.
func renderTemplate(templateName string, data any) (string, error) {
	tmplPath := path.Join("templates", templateName)

	tmplData, err := triggerTemplates.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("error reading template file: %w", err)
	}

	tmpl, err := template.New(templateName).Parse(string(tmplData))
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return buf.String(), nil
}
