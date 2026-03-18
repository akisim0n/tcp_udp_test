package builders

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type BatchBuilder struct {
	batch   *pgx.Batch
	readers []func(pgx.BatchResults) error
}

func NewBatchBuilder() *BatchBuilder {
	return &BatchBuilder{batch: &pgx.Batch{}}
}

func (b *BatchBuilder) AddQueryRow(retArgs ...any) func(query string, args ...any) {
	return func(query string, args ...any) {
		b.batch.Queue(query, args...)

		b.readers = append(b.readers, func(rows pgx.BatchResults) error {
			return rows.QueryRow().Scan(retArgs...)
		})
	}
}

func (b *BatchBuilder) AddQuery(retRows *pgx.Rows) func(query string, args ...any) {
	return func(query string, args ...any) {
		b.batch.Queue(query, args...)

		b.readers = append(b.readers, func(rows pgx.BatchResults) error {
			var execErr error
			*retRows, execErr = rows.Query()
			return execErr
		})
	}
}

func (b *BatchBuilder) AddExec() func(query string, args ...any) {
	return func(query string, args ...any) {
		b.batch.Queue(query, args...)

		b.readers = append(b.readers, func(rows pgx.BatchResults) error {
			_, execErr := rows.Exec()
			return execErr
		})
	}
}

func (b *BatchBuilder) Run(ctx context.Context, tx pgx.Tx) error {
	res := tx.SendBatch(ctx, b.batch)
	defer func() {
		_ = res.Close()
	}()

	for _, reader := range b.readers {
		if err := reader(res); err != nil {
			return err
		}
	}

	return nil
}
