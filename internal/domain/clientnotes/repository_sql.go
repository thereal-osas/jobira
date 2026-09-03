package clientnotes

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, note *ClientCleanerNote) error {
	query := `
		INSERT INTO client_cleaner_notes (
			client_id,
			cleaner_id,
			note,
			created_at,
			updated_at
		)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at 
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		note.ClientID,
		note.CleanerID,
		note.Note,
		now,
		now,
	).Scan(
		&note.ID,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
}

func (r *SQLRepository) GetByCleanerID(ctx context.Context, clientID uint, cleanerID uint) ([]ClientCleanerNote, error) {
	query := `
		SELECT 
			id,
			client_id,
			cleaner_id,
			note,
			created_at,
			updated_at
		FROM client_cleaner_notes
		WHERE client_id = $1 
		AND cleaner_id = $2 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID, cleanerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []ClientCleanerNote

	for rows.Next() {
		var note ClientCleanerNote

		err := rows.Scan(
			&note.ID,
			&note.ClientID,
			&note.CleanerID,
			&note.Note,
			&note.CreatedAt,
			&note.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		notes = append(notes, note)
	}

	return notes, rows.Err()
}

func (r *SQLRepository) Update(ctx context.Context, noteID uint, clientID uint, note string) error {
	query := `
	UPDATE client_cleaner_notes
	SET 
		note = $1, 
		updated_at = $2 
	WHERE id = $3 
	AND client_id = $4 	
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		note,
		time.Now(),
		noteID,
		clientID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoteNotFound
	}

	return nil

}

func (r *SQLRepository) Delete(ctx context.Context, noteID uint, clientID uint) error {
	query := `
		DELETE FROM client_cleaner_notes
		WHERE id = $1 
		AND client_id = $2 
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		noteID,
		clientID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoteNotFound
	}

	return nil

}
