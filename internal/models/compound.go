package models

// Compound represents a chemical compound with its formula and constituent elements.
type Compound struct {
	Formula    string         // The chemical formula of the compound, e.g., "H2O" for water
	Name       string         // The name of chemical compound
	Appearance string         // The appearance of chemical compound
	Data       map[string]int // A map containing the elements and their respective counts in the compound
}
