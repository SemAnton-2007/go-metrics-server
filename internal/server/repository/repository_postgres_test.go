package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("NewPostgresRepository", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Ожидаем запросы на создание таблиц
		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS gauges`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS counters`).WillReturnResult(sqlmock.NewResult(0, 0))

		repo, err := NewPostgresRepository(db)
		require.NoError(t, err)
		require.NotNil(t, repo)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateGauge", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS gauges`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS counters`).WillReturnResult(sqlmock.NewResult(0, 0))

		repo, err := NewPostgresRepository(db)
		require.NoError(t, err)

		mock.ExpectExec(`INSERT INTO gauges`).
			WithArgs("test_gauge", 1.23).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.UpdateGauge(ctx, "test_gauge", 1.23)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateCounter", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS gauges`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS counters`).WillReturnResult(sqlmock.NewResult(0, 0))

		repo, err := NewPostgresRepository(db)
		require.NoError(t, err)

		mock.ExpectExec(`INSERT INTO counters`).
			WithArgs("test_counter", int64(10)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.UpdateCounter(ctx, "test_counter", 10)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetGauge", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS gauges`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS counters`).WillReturnResult(sqlmock.NewResult(0, 0))

		repo, err := NewPostgresRepository(db)
		require.NoError(t, err)

		rows := sqlmock.NewRows([]string{"value"}).AddRow(1.23)
		mock.ExpectQuery(`SELECT value FROM gauges WHERE name = \$1`).
			WithArgs("test_gauge").
			WillReturnRows(rows)

		value, err := repo.GetGauge(ctx, "test_gauge")
		require.NoError(t, err)
		require.Equal(t, 1.23, value)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetCounter", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS gauges`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`CREATE TABLE IF NOT EXISTS counters`).WillReturnResult(sqlmock.NewResult(0, 0))

		repo, err := NewPostgresRepository(db)
		require.NoError(t, err)

		rows := sqlmock.NewRows([]string{"value"}).AddRow(int64(10))
		mock.ExpectQuery(`SELECT value FROM counters WHERE name = \$1`).
			WithArgs("test_counter").
			WillReturnRows(rows)

		value, err := repo.GetCounter(ctx, "test_counter")
		require.NoError(t, err)
		require.Equal(t, int64(10), value)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
