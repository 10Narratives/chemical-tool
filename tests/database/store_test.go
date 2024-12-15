package database

import (
	"chemical-tool/internal/database"
	"chemical-tool/internal/models"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetElement(t *testing.T) {
	t.Parallel()

	var (
		name         string  = "Водород"
		symbol       string  = "H"
		atomicWeight float64 = 1.0
	)

	tests := []struct {
		name        string
		mocks       func(dbMock sqlmock.Sqlmock)
		symbol      string
		wantElement require.ValueAssertionFunc
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name: "success",
			mocks: func(dbMock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "symbol", "atomic_weight"}).
					AddRow(name, symbol, atomicWeight)
				dbMock.ExpectQuery("SELECT name symbol atomic_weight FROM periodic_table WHERE symbol = ?").
					WithArgs(symbol).
					WillReturnRows(rows)
			},
			symbol: symbol,
			wantElement: func(tt require.TestingT, i1 interface{}, i2 ...interface{}) {
				element, ok := i1.(models.Element)
				require.True(t, ok)
				require.Equal(t, name, element.Name)
				require.Equal(t, symbol, element.Symbol)
				require.Equal(t, atomicWeight, element.AtomicWeight)
			},
			wantErr: require.NoError,
		},
		{
			name: "no rows",
			mocks: func(dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery("SELECT name symbol atomic_weight FROM periodic_table WHERE symbol = ?").
					WithArgs(symbol).
					WillReturnError(sql.ErrNoRows)
			},
			symbol: symbol,
			wantElement: func(tt require.TestingT, got interface{}, i ...interface{}) {
				element, ok := got.(models.Element)
				require.True(t, ok)
				require.Equal(t, models.Element{}, element)
			},
			wantErr: require.NoError,
		},
		{
			name: "database error",
			mocks: func(dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery("SELECT name symbol atomic_weight FROM periodic_table WHERE symbol = ?").
					WithArgs(symbol).
					WillReturnError(errors.New("database error"))
			},
			symbol: symbol,
			wantElement: func(tt require.TestingT, got interface{}, i ...interface{}) {
				element, ok := got.(models.Element)
				require.True(t, ok)
				require.Equal(t, models.Element{}, element)
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.EqualError(t, err, "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, dbMock, err := sqlmock.New()
			require.NoError(t, err)

			store := database.Store{DB: db}
			tt.mocks(dbMock)

			parcel, err := store.GetElement(tt.symbol)
			tt.wantErr(t, err)
			tt.wantElement(t, parcel)

			require.NoError(t, dbMock.ExpectationsWereMet())
		})
	}

}
