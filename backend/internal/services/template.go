package services

import (
	"fmt"
	"regexp"
	"strings"
)

var variablePattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// RenderTemplate replaces {{variable_name}} placeholders with values from inputs.
// Returns an error if any variable in the template is not present in inputs.
func RenderTemplate(tmpl string, inputs map[string]interface{}) (string, error) {
	matches := variablePattern.FindAllStringSubmatch(tmpl, -1)
	missing := []string{}
	result := tmpl

	for _, match := range matches {
		varName := match[1]
		val, ok := inputs[varName]
		if !ok {
			missing = append(missing, varName)
			continue
		}
		result = strings.ReplaceAll(result, match[0], fmt.Sprintf("%v", val))
	}

	if len(missing) > 0 {
		return "", fmt.Errorf("undefined variables: %s", strings.Join(missing, ", "))
	}

	return result, nil
}

// ExtractVariables returns deduplicated variable names found in a template.
func ExtractVariables(tmpl string) []string {
	matches := variablePattern.FindAllStringSubmatch(tmpl, -1)
	seen := map[string]bool{}
	var vars []string
	for _, match := range matches {
		name := match[1]
		if !seen[name] {
			vars = append(vars, name)
			seen[name] = true
		}
	}
	return vars
}
