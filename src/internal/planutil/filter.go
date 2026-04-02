package planutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pocketbase/dbx"
)

// FilterTerm defines a single equality clause in a PocketBase filter.
type FilterTerm struct {
	Field string
	Value interface{}
}

// Filter wraps a PocketBase filter expression and its bound params.
type Filter struct {
	Expression string
	Params     dbx.Params
}

var safeFilterFieldPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// BuildEqualsFilter builds a parameterized equality filter.
func BuildEqualsFilter(terms ...FilterTerm) (Filter, error) {
	parts := make([]string, 0, len(terms))
	params := dbx.Params{}

	for i, term := range terms {
		if err := validateFilterField(term.Field); err != nil {
			return Filter{}, err
		}
		if err := validateFilterValue(term.Field, term.Value); err != nil {
			return Filter{}, err
		}

		paramKey := fmt.Sprintf("p%d", i)
		parts = append(parts, fmt.Sprintf("%s = {:%s}", term.Field, paramKey))
		params[paramKey] = term.Value
	}

	return Filter{
		Expression: strings.Join(parts, " && "),
		Params:     params,
	}, nil
}

// BuildContainsFilter builds a validated parameterized contains filter.
func BuildContainsFilter(field, value string) (Filter, error) {
	if err := validateFilterField(field); err != nil {
		return Filter{}, err
	}
	if err := validateFilterValue(field, value); err != nil {
		return Filter{}, err
	}

	return Filter{
		Expression: fmt.Sprintf("%s ~ {:value}", field),
		Params:     dbx.Params{"value": value},
	}, nil
}

func validateFilterField(field string) error {
	if !safeFilterFieldPattern.MatchString(field) {
		return fmt.Errorf("%s is not a supported filter field", field)
	}

	return nil
}

func validateFilterValue(field string, value interface{}) error {
	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("%s is empty", field)
		}
		return nil
	case bool:
		return nil
	default:
		return fmt.Errorf("%s has unsupported filter value type", field)
	}
}
