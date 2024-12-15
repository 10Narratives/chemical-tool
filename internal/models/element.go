package models

// Element represents a chemical element with its symbol and atomic weight.
type Element struct {
	Name         string  // The name of the element
	Symbol       string  // The symbol of the element, e.g., "H" for hydrogen
	AtomicWeight float64 // The atomic weight of the element, e.g., 1.008 for hydrogen
}
