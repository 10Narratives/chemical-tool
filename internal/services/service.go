package services

import (
	"chemical-tool/internal/database"
	"chemical-tool/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ChemicalService provides methods to interact with chemical data,
// including parsing chemical formulas and managing elements.
//
// It holds a reference to a database.Store which is used to perform
// operations related to the underlying storage of chemical elements
// and compounds.
type ChemicalService struct {
	store database.Store
}

// ParseCompound parses a chemical formula and returns a Compound object
// along with any potential error encountered during the parsing process.
//
// The function utilizes regular expressions to handle nested groups
// and individual elements in the formula. It expands any grouped
// elements according to their multipliers and counts the occurrences
// of each element in the entire formula.
//
// Arguments:
//   - formula: A string representing the chemical formula to parse.
//
// Returns:
//   - models.Compound: A structure containing the final formula and
//     a map of elements with their respective counts.
//   - error: An error value that will be non-nil if any issues were
//     encountered during parsing, including unknown elements or
//     conversion errors.
//
// Example:
//
//	compound, err := service.ParseCompound("C6H12O6")
//	if err != nil {
//	    // handle error
//	}
func (service ChemicalService) ParseCompound(formula string) (models.Compound, error) {
	var err error
	elementCounts := make(map[string]int)

	elementPattern := regexp.MustCompile(`([A-Z][a-z]*)(\d*)`)
	groupPattern := regexp.MustCompile(`\(([^()]+)\)(\d*)`)

	for {
		matches := groupPattern.FindStringSubmatch(formula)
		if matches == nil {
			break
		}

		group := matches[1]
		multiplier := 1
		if matches[2] != "" {
			multiplier, err = strconv.Atoi(matches[2])
			if err != nil {
				return models.Compound{}, err
			}
		}

		expanded := ""
		for _, match := range elementPattern.FindAllStringSubmatch(group, -1) {
			element := match[1]
			count := 1
			if match[2] != "" {
				count, err = strconv.Atoi(match[2])
				if err != nil {
					return models.Compound{}, err
				}
			}
			expanded += fmt.Sprintf("%s%d", element, count*multiplier)
		}

		formula = strings.Replace(formula, matches[0], expanded, 1)
	}

	for _, match := range elementPattern.FindAllStringSubmatch(formula, -1) {
		element := match[1]
		count := 1
		if match[2] != "" {
			count, err = strconv.Atoi(match[2])
			if err != nil {
				return models.Compound{}, err
			}
		}
		_, err := service.store.GetElement(element)
		if err == sql.ErrNoRows {
			err = errors.New("Gotten unknown element in compound")
			return models.Compound{}, err
		}
		if err != nil {
			return models.Compound{}, nil
		}
		elementCounts[element] += count
	}

	return models.Compound{
		Formula: formula,
		Data:    elementCounts,
	}, nil
}
