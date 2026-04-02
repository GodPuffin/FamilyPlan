package planutil

import (
	"fmt"
	"regexp"
	"strings"
)

// FilterTerm defines a single equality clause in a PocketBase filter.
type FilterTerm struct {
	Field string
	Value string
}

var safeFilterValuePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// BuildEqualsFilter builds a filter with validated equality clauses.
func BuildEqualsFilter(terms ...FilterTerm) (string, error) {
	parts := make([]string, 0, len(terms))

	for _, term := range terms {
		if err := validateFilterValue(term.Field, term.Value); err != nil {
			return "", err
		}

		parts = append(parts, fmt.Sprintf("%s = '%s'", term.Field, term.Value))
	}

	return strings.Join(parts, " && "), nil
}

// BuildContainsFilter builds a validated contains filter.
func BuildContainsFilter(field, value string) (string, error) {
	if err := validateFilterValue(field, value); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s ~ '%s'", field, value), nil
}

func validateFilterValue(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is empty", field)
	}

	if !safeFilterValuePattern.MatchString(value) {
		return fmt.Errorf("%s contains unsupported characters", field)
	}

	return nil
}
