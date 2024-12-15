package services_test

import (
	"chemical-tool/internal/models"
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

func TestComputeData(t *testing.T) {
	type testCase struct {
		name           string
		compound       models.Compound
		elements       []models.Element
		expectedResult services.MolarMassResponse
	}

	tests := []testCase{
		{
			name: "Simple compound: H2O",
			compound: models.Compound{
				Formula: "H2O",
				Data: map[string]int{
					"H": 2,
					"O": 1,
				},
			},
			elements: []models.Element{
				{Name: "Hydrogen", Symbol: "H", AtomicWeight: 1.008},
				{Name: "Oxygen", Symbol: "O", AtomicWeight: 16.00},
			},
			expectedResult: services.MolarMassResponse{
				GeneralWeight: (2.016 + 16.00) / 2, // Sum of element weights divided by the number of elements
				ElementsInfo: []services.MolarMassElementInfo{
					{Name: "Hydrogen", AtomsCount: 2, WeightInCompound: 2.016, WeightPercent: 22.4},
					{Name: "Oxygen", AtomsCount: 1, WeightInCompound: 16.00, WeightPercent: 177.6},
				},
			},
		},
		{
			name: "Compound with multiple elements: C6H12O6",
			compound: models.Compound{
				Formula: "C6H12O6",
				Data: map[string]int{
					"C": 6,
					"H": 12,
					"O": 6,
				},
			},
			elements: []models.Element{
				{Name: "Carbon", Symbol: "C", AtomicWeight: 12.011},
				{Name: "Hydrogen", Symbol: "H", AtomicWeight: 1.008},
				{Name: "Oxygen", Symbol: "O", AtomicWeight: 16.00},
			},
			expectedResult: services.MolarMassResponse{
				GeneralWeight: (72.066 + 12.096 + 96.00) / 3,
				ElementsInfo: []services.MolarMassElementInfo{
					{Name: "Carbon", AtomsCount: 6, WeightInCompound: 72.066, WeightPercent: 33.3333},
					{Name: "Hydrogen", AtomsCount: 12, WeightInCompound: 12.096, WeightPercent: 5.5882},
					{Name: "Oxygen", AtomsCount: 6, WeightInCompound: 96.00, WeightPercent: 45.683},
				},
			},
		},
		{
			name: "Single element compound: O2",
			compound: models.Compound{
				Formula: "O2",
				Data: map[string]int{
					"O": 2,
				},
			},
			elements: []models.Element{
				{Name: "Oxygen", Symbol: "O", AtomicWeight: 16.00},
			},
			expectedResult: services.MolarMassResponse{
				GeneralWeight: 32.00,
				ElementsInfo: []services.MolarMassElementInfo{
					{Name: "Oxygen", AtomsCount: 2, WeightInCompound: 32.00, WeightPercent: 100.0},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := services.MolarMassService{}
			result := service.ComputeData(tc.compound, tc.elements)

			// Check the general weight
			assert.InEpsilon(t, tc.expectedResult.GeneralWeight, result.GeneralWeight, 0.0001, "GeneralWeight mismatch")

			// Check elements info
			assert.Len(t, result.ElementsInfo, len(tc.expectedResult.ElementsInfo), "ElementsInfo length mismatch")
			for i, expectedInfo := range tc.expectedResult.ElementsInfo {
				assert.Equal(t, expectedInfo.Name, result.ElementsInfo[i].Name, "Name mismatch")
				assert.Equal(t, expectedInfo.AtomsCount, result.ElementsInfo[i].AtomsCount, "AtomsCount mismatch")
				assert.InEpsilon(t, expectedInfo.WeightInCompound, result.ElementsInfo[i].WeightInCompound, 0.0001, "WeightInCompound mismatch")
			}
		})
	}
}
