package database

import (
	"chemical-tool/internal/models"
	"database/sql"
	"errors"
)

// Store represents a storage system that interacts with a database.
//
// It contains a single field `db` which is a pointer to an sql.DB instance,
// allowing for database operations such as querying and transactions.
type Store struct {
	db *sql.DB
}

// GetElement retrieves an element from the periodic table by its symbol.
//
// It queries the database for the element with the specified symbol,
// scanning the result into a models.Element struct.
//
// Parameters:
//
//	symbol (string): The chemical symbol of the element to retrieve.
//
// Returns:
//
//	models.Element: The element corresponding to the provided symbol.
//	error: An error, if any occurred during the database query.
//	       If no element is found, the returned element will be empty
//	       and the error will be nil.
func (store Store) GetElement(symbol string) (models.Element, error) {
	row := store.db.QueryRow("SELECT symbol atomic_weight FROM periodic_table WHERE symbol = ?", symbol)

	gottenElement := models.Element{}
	err := row.Scan(&gottenElement.Symbol, &gottenElement.AtomicWeight)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Element{}, nil
	}
	if err != nil {
		return models.Element{}, err
	}

	return gottenElement, nil
}
