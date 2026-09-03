package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestWithTx_Success(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer database.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	called := false

	err = WithTx(
		context.Background(),
		database,
		func(tx *sql.Tx) error {
			called = true

			if tx == nil {
				t.Fatal("expected transaction")
			}

			return nil
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !called {
		t.Fatal("expected callback to run")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestWithTx_CallbackErrorRollsBack(
	t *testing.T,
) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer database.Close()

	expectedErr := errors.New("callback failed")

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = WithTx(
		context.Background(),
		database,
		func(tx *sql.Tx) error {
			return expectedErr
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestWithTx_BeginError(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer database.Close()

	expectedErr := errors.New("begin failed")

	mock.ExpectBegin().
		WillReturnError(expectedErr)

	called := false

	err = WithTx(
		context.Background(),
		database,
		func(tx *sql.Tx) error {
			called = true
			return nil
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if called {
		t.Fatal(
			"callback must not run when begin fails",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestWithTx_CommitError(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer database.Close()

	expectedErr := errors.New("commit failed")

	mock.ExpectBegin()
	mock.ExpectCommit().
		WillReturnError(expectedErr)

	err = WithTx(
		context.Background(),
		database,
		func(tx *sql.Tx) error {
			return nil
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestWithTx_PanicRollsBack(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer database.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	defer func() {
		rec := recover()

		if rec == nil {
			t.Fatal("expected panic")
		}

		if rec != "boom" {
			t.Fatalf(
				"expected panic boom, got %v",
				rec,
			)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf(
				"unmet expectations: %v",
				err,
			)
		}
	}()

	_ = WithTx(
		context.Background(),
		database,
		func(tx *sql.Tx) error {
			panic("boom")
		},
	)
}

func TestWithTx_ContextCancelled(
	t *testing.T,
) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer database.Close()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err = WithTx(
		ctx,
		database,
		func(tx *sql.Tx) error {
			t.Fatal(
				"callback must not run",
			)
			return nil
		},
	)

	if err == nil {
		t.Fatal(
			"expected cancelled context error",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}
