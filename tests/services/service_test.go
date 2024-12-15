package services_test

import (
	"chemical-tool/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCompound(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name         string
		formula      string
		expectedData map[string]int
		expectError  bool
	}

	tests := []testCase{
		{
			name:    "Simple compound without multipliers",
			formula: "H2O",
			expectedData: map[string]int{
				"H": 2,
				"O": 1,
			},
			expectError: false,
		},
		{
			name:    "Compound with group multiplier",
			formula: "(OH)2",
			expectedData: map[string]int{
				"O": 2,
				"H": 2,
			},
			expectError: false,
		},
		{
			name:    "Nested groups with multipliers",
			formula: "Mg(OH)2",
			expectedData: map[string]int{
				"Mg": 1,
				"O":  2,
				"H":  2,
			},
			expectError: false,
		},
		{
			name:    "Multiple groups and elements",
			formula: "C6H12O6",
			expectedData: map[string]int{
				"C": 6,
				"H": 12,
				"O": 6,
			},
			expectError: false,
		},
		{
			name:         "Empty formula",
			formula:      "",
			expectedData: map[string]int{},
			expectError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			service := services.ChemicalService{}
			result, err := service.ParseCompound(tc.formula)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData, result.Data)
			}
		})
	}
}
