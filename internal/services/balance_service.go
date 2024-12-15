package services

import (
	"fmt"
	"strings"

	"gonum.org/v1/gonum/mat"
)

// BalanceService handles chemical balancing operations.  It relies on a ChemicalService for underlying chemical information.
type BalanceService struct {
	ChemicalService
}

// BalanceCompoundInfo represents information about a compound involved in a balanced reaction.
type BalanceCompoundInfo struct {
	Formula    string // Chemical formula (e.g., "H2O")
	Name       string // Name of the compound (e.g., "Water")
	Appearance string // Physical appearance (e.g., "Clear liquid")
}

// BalanceResponse represents the response from a chemical balancing request.
type BalanceResponse struct {
	Result   string                // Balanced equation
	Reagents []BalanceCompoundInfo // Array of reagent compound information.
	Products []BalanceCompoundInfo // Array of product compound information.
}

// fillCompoundInfo retrieves compound information from the data store and converts it to a slice of BalanceCompoundInfo structs.
//
// Args:
//
//	formulas: A slice of chemical formulas (e.g., ["H2O", "C6H12O6"]).
//
// Returns:
//
//	[]BalanceCompoundInfo: A slice of BalanceCompoundInfo structs containing the retrieved compound information.
//	Returns an empty slice if no compounds are found or an error occurs.
//	error: An error object if there's an issue during retrieval from the data store.
func (service BalanceService) fillCompoundInfo(formulas []string) ([]BalanceCompoundInfo, error) {
	compounds, err := service.store.GetCompounds(formulas)
	if err != nil {
		return nil, err
	}

	compoundsInfo := make([]BalanceCompoundInfo, len(compounds))
	for i, compound := range compounds {
		newCompoundInfo := BalanceCompoundInfo{
			Formula:    compound.Formula,
			Name:       compound.Name,
			Appearance: compound.Appearance,
		}
		compoundsInfo[i] = newCompoundInfo
	}

	return compoundsInfo, nil
}

// GetResponse balances a chemical equation represented as a string and returns the balanced equation.
//
// It processes the input `requestedData`, which is expected to be in the form of
// "reagents => products", where both reagents and products are separated by a "+" sign.
// The function solves the system of linear equations formed by the stoichiometry of the chemical compounds
// involved and returns the balanced equation as a string in the form of "reactantSide => productSide",
// along with additional information about reagents and products.
//
// Arguments:
//
//	requestedData (string): The chemical equation to be balanced, formatted as "reagents => products".
//
// Returns:
//
//	BalanceResponse: A struct containing the balanced equation in `Result`, and additional compound information
//	in `Reagents` and `Products`.
//	error: An error, if any, that occurred during the balancing process.
//
// Example:
//
//	requestedData := "H2 + O2 => H2O"
//	response, err := service.GetResponse(requestedData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(response.Result)
func (service BalanceService) GetResponse(requestedData string) (BalanceResponse, error) {
	var response BalanceResponse

	matrix, _, splitIndex := service.buildMatrix(requestedData)
	coefficients, err := service.solveMatrix(matrix)
	if err != nil {
		return response, err
	}

	sides := strings.Split(requestedData, "=>")
	reagents := strings.Split(strings.TrimSpace(sides[0]), "+")
	products := strings.Split(strings.TrimSpace(sides[1]), "+")

	reactantSide := []string{}
	for i, compound := range reagents {
		reactantSide = append(reactantSide, fmt.Sprintf("%d %s", coefficients[i], strings.TrimSpace(compound)))
	}

	productSide := []string{}
	for i, compound := range products {
		productSide = append(productSide, fmt.Sprintf("%d %s", coefficients[splitIndex+i], strings.TrimSpace(compound)))
	}

	response.Result = strings.Join(reactantSide, " + ") + " => " + strings.Join(productSide, " + ")
	response.Reagents, _ = service.fillCompoundInfo(reagents)
	response.Products, _ = service.fillCompoundInfo(products)

	return response, nil
}

// buildMatrix constructs a matrix representing the stoichiometric coefficients of a chemical equation
// by parsing the equation into reagents and products and mapping them to elements.
//
// The function splits the input equation into two parts: reagents and products, and extracts the chemical elements
// involved from both sides of the equation. It then creates a matrix where each row represents an element,
// and each column corresponds to a compound in the equation. The matrix is populated with the stoichiometric coefficients
// of the elements in each compound, with positive values for reagents and negative values for products.
//
// Arguments:
//
//	equation (string): A chemical equation to be parsed, in the form "reagents => products".
//
// Returns:
//
//	*mat.Dense: A matrix representing the stoichiometric coefficients of the equation.
//	[]string: A list of unique elements in the equation.
//	int: The index that separates the reagents and products in the matrix.
//
// Example:
//
//	equation := "H2 + O2 => H2O"
//	matrix, elements, splitIndex := service.buildMatrix(equation)
func (service BalanceService) buildMatrix(equation string) (*mat.Dense, []string, int) {
	sides := strings.Split(equation, "=>")
	reagents := strings.Split(strings.TrimSpace(sides[0]), "+")
	products := strings.Split(strings.TrimSpace(sides[1]), "+")

	compounds := append(reagents, products...)
	allElements := make(map[string]struct{})

	for _, compound := range compounds {
		parsed, _ := service.ParseCompound(strings.TrimSpace(compound))
		for element := range parsed.Data {
			allElements[element] = struct{}{}
		}
	}
	elementList := make([]string, 0, len(allElements))
	for element := range allElements {
		elementList = append(elementList, element)
	}

	rows := len(elementList)
	cols := len(compounds)
	data := make([]float64, rows*cols)

	for i, element := range elementList {
		for j, compound := range reagents {
			parsed, _ := service.ParseCompound(strings.TrimSpace(compound))
			data[i*cols+j] = float64(parsed.Data[element])
		}
		for j, compound := range products {
			parsed, _ := service.ParseCompound(strings.TrimSpace(compound))
			data[i*cols+len(reagents)+j] = float64(-parsed.Data[element])
		}
	}

	return mat.NewDense(rows, cols, data), elementList, len(reagents)
}

// solveMatrix solves the system of linear equations represented by the matrix.
// It uses the Gaussian elimination method to solve for the unknown coefficients
// in the stoichiometric matrix of a chemical equation.
//
// The matrix represents the stoichiometry of a chemical reaction, and the function
// computes the solution vector that contains the coefficients for the chemical compounds.
//
// Arguments:
//
//	matrix (*mat.Dense): The matrix representing the stoichiometric coefficients of the chemical equation,
//	                      where each row represents an element and each column represents a compound.
//
// Returns:
//
//	[]int: A slice of integers representing the normalized stoichiometric coefficients for the compounds.
//	error: An error if the matrix could not be solved (e.g., if the system is singular or inconsistent).
func (service BalanceService) solveMatrix(matrix *mat.Dense) ([]int, error) {
	rows, cols := matrix.Dims()
	rhs := make([]float64, rows)
	coefficients := make([]float64, cols)

	for i := 0; i < rows; i++ {
		rhs[i] = 0
	}

	var x mat.VecDense
	err := x.SolveVec(matrix, mat.NewVecDense(rows, rhs))
	if err != nil {
		return nil, err
	}

	for i := 0; i < cols; i++ {
		coefficients[i] = x.AtVec(i)
	}
	return service.normalize(coefficients), nil
}

// normalize normalizes the coefficients obtained from solving the stoichiometric matrix.
// The function scales the coefficients to the smallest integer values by multiplying
// each coefficient to avoid fractions and finding the least common multiple (LCM) of the coefficients.
//
// Arguments:
//
//	coefficients ([]float64): A slice of floating-point coefficients obtained from solving the matrix.
//
// Returns:
//
//	[]int: A slice of normalized integers representing the stoichiometric coefficients.
func (service BalanceService) normalize(coefficients []float64) []int {
	intCoefficients := make([]int, len(coefficients))
	lcmValue := 1

	for i, value := range coefficients {
		intCoefficients[i] = int(value * 1000000) // Преобразование в целое число.
		if intCoefficients[i] != 0 {
			lcmValue = service.lcm(lcmValue, service.abs(intCoefficients[i]))
		}
	}

	for i := range intCoefficients {
		intCoefficients[i] = intCoefficients[i] / (lcmValue / service.abs(intCoefficients[i]))
	}
	return intCoefficients
}

// gcd calculates the greatest common divisor (GCD) of two integers using the Euclidean algorithm.
//
// The GCD of two numbers is the largest integer that divides both of them without leaving a remainder.
//
// Arguments:
//
//	a (int): The first integer.
//	b (int): The second integer.
//
// Returns:
//
//	int: The greatest common divisor of the two integers.
func (service BalanceService) gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// lcm calculates the least common multiple (LCM) of two integers using the formula:
// LCM(a, b) = (a * b) / GCD(a, b)
//
// The LCM of two numbers is the smallest integer that is divisible by both numbers.
//
// Arguments:
//
//	a (int): The first integer.
//	b (int): The second integer.
//
// Returns:
//
//	int: The least common multiple of the two integers.
func (service BalanceService) lcm(a, b int) int {
	return a / service.gcd(a, b) * b
}

// abs calculates the absolute value of an integer.
//
// It returns the absolute value, which is always non-negative.
//
// Arguments:
//
//	x (int): The integer whose absolute value is to be calculated.
//
// Returns:
//
//	int: The absolute value of the integer.
func (service BalanceService) abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
