package database

import (
	"chemical-tool/internal/models"
	"database/sql"
	"errors"
)

// openDB opens a database connection using the provided driver and data source name (DSN).
//
// Parameters:
//   - driver: A string specifying the database driver (e.g., "postgres", "mysql").
//   - dns: A string specifying the data source name (DSN), which contains the
//     connection details such as host, port, username, password, and database name.
//
// Returns:
//   - *sql.DB: A pointer to the database connection object, which can be used to
//     interact with the database.
//   - func(): A cleanup function that closes the database connection. It should
//     be called to release resources when the connection is no longer needed.
//   - error: An error object if the connection fails to open. If nil, the
//     connection was successful.
//
// Usage:
//
//	db, closeFunc, err := openDB("postgres", "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=disable")
//	if err != nil {
//	    log.Fatalf("Failed to open database: %v", err)
//	}
//	defer closeFunc() // Ensure the database connection is closed when done.
//
// Notes:
//   - It is important to call the cleanup function (returned as the second value)
//     to close the database connection and prevent resource leaks.
//   - The `sql.DB` object is thread-safe and can be used concurrently by multiple
//     goroutines.
func openDB(driver, dns string) (*sql.DB, func(), error) {
	db, err := sql.Open(driver, dns)
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() {
		_ = db.Close()
	}

	return db, closeFunc, nil
}

// FIXME DB field is public

// Store represents a storage system that interacts with a database.
//
// It contains a single field `db` which is a pointer to an sql.DB instance,
// allowing for database operations such as querying and transactions.
type Store struct {
	DB *sql.DB
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
	row := store.DB.QueryRow("SELECT name symbol atomic_weight FROM periodic_table WHERE symbol = ?", symbol)

	gottenElement := models.Element{}
	err := row.Scan(&gottenElement.Name, &gottenElement.Symbol, &gottenElement.AtomicWeight)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Element{}, nil
	}
	if err != nil {
		return models.Element{}, err
	}

	return gottenElement, nil
}

// GetElements retrieves the elements that make up a given compound.
//
// It takes a 'target' of type models.Compound, which contains the
// data representing the compound's elements. The function iterates
// over the symbols in the target's data, fetching each corresponding
// element from the store. If any retrieval fails, it returns an error.
//
// Returns:
//   - A slice of models.Element containing the elements of the compound,
//     or nil if an error occurred.
//   - An error indicating what went wrong, if applicable.
func (store Store) GetElements(target models.Compound) ([]models.Element, error) {
	var gottenElements []models.Element = make([]models.Element, 0)
	for symbol := range target.Data {
		element, err := store.GetElement(symbol)
		if err != nil {
			return nil, err
		}
		gottenElements = append(gottenElements, element)
	}
	return gottenElements, nil
}
