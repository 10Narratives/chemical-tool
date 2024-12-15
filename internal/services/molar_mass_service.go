package services

import "chemical-tool/internal/models"

// MolarMassService is a service that provides functionalities
// related to calculating the molar mass of chemical compounds.
// It embeds the ChemicalService to utilize its methods and
// properties.
type MolarMassService struct {
	ChemicalService
}

// MolarMassElementInfo holds information about an element in
// a compound, including its name, weight in the compound,
// the count of atoms of that element, and the weight percentage
// of the element in the compound.
type MolarMassElementInfo struct {
	Name             string  // Name of the element
	WeightInCompound float64 // Weight of the element in the compound
	AtomsCount       uint16  // Number of atoms of the element in the compound
	WeightPercent    float64 // Weight percentage of the element in the compound
}

// MolarMassResponse encapsulates the result of a molar mass
// calculation, including the total general weight and a list
// of element information.
type MolarMassResponse struct {
	GeneralWeight float64                // Total weight of the compound
	ElementsInfo  []MolarMassElementInfo // Slice of element information
}

// GetResponse processes the provided requestedData string to
// calculate the molar mass response. It attempts to parse
// a chemical compound from the string, retrieve its elements,
// and compute the molar mass details. It returns a MolarMassResponse
// containing the calculated data and an error if anything goes
// wrong during the process.
//
// Parameters:
//   - requestedData: A string representing the chemical compound
//     for which the molar mass is to be calculated.
//
// Returns:
//   - MolarMassResponse: The response containing general weight
//     and detailed information about the elements in the compound.
//   - error: An error indicator, nil if no errors occurred.
func (service MolarMassService) GetResponse(requestedData string) (MolarMassResponse, error) {
	response := MolarMassResponse{}
	response.ElementsInfo = nil
	compound, err := service.ParseCompound(requestedData)
	if err != nil {
		return response, err
	}
	elements, err := service.store.GetElements(compound)
	if err != nil {
		return response, nil
	}

	response = service.computeData(compound, elements)

	return response, nil
}

// computeData processes the given compound and elements to
// compute the molar mass data. It calculates the general
// weight of the compound and provides detailed information
// about each element in relation to the compound.
//
// Parameters:
//   - compound: The chemical compound whose molar mass needs
//     to be calculated.
//   - elements: A slice of elements involved in the compound.
//
// Returns:
//   - MolarMassResponse: A response containing the general weight
//     and detailed information about the elements in the compound.
func (service MolarMassService) computeData(compound models.Compound, elements []models.Element) MolarMassResponse {
	var (
		generalWeight float64                = 0.0
		elementsInfo  []MolarMassElementInfo = make([]MolarMassElementInfo, len(elements))
	)

	sumWeight := 0.0
	for _, element := range elements {
		elementWeight := element.AtomicWeight * float64(compound.Data[element.Symbol])
		sumWeight += elementWeight
	}
	generalWeight = sumWeight / float64(len(elements))

	for i, element := range elements {
		elementsInfo[i].Name = element.Name
		elementsInfo[i].AtomsCount = uint16(compound.Data[element.Symbol])
		elementsInfo[i].WeightInCompound = element.AtomicWeight * float64(compound.Data[element.Symbol])
		elementsInfo[i].WeightPercent = elementsInfo[i].WeightInCompound / generalWeight * 100
	}

	return MolarMassResponse{
		GeneralWeight: generalWeight,
		ElementsInfo:  elementsInfo,
	}
}
