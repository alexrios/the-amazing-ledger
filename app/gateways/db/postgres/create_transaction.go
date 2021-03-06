package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/stone-co/the-amazing-ledger/app"
	"github.com/stone-co/the-amazing-ledger/app/domain/entities"
	"github.com/stone-co/the-amazing-ledger/app/shared/instrumentation/newrelic"
)

func (r *LedgerRepository) CreateTransaction(ctx context.Context, transaction *entities.Transaction) error {
	operation := "Repository.CreateTransaction"
	query := `
		INSERT INTO
			entries (
				id,
				account_class,
				account_group,
				account_subgroup,
				account_id,
	  			operation,
				amount,
				version,
				transaction_id,
				account_suffix
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	defer newrelic.NewDatastoreSegment(ctx, collection, operation, query).End()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var batch pgx.Batch

	for _, entry := range transaction.Entries {
		batch.Queue(
			query,
			entry.ID,
			entry.Account.Class.String(),
			entry.Account.Group,
			entry.Account.Subgroup,
			entry.Account.ID,
			entry.Operation.String(),
			entry.Amount,
			entry.Version,
			transaction.ID,
			entry.Account.Suffix,
		)
	}

	br := tx.SendBatch(ctx, &batch)
	err = br.Close()
	if err != nil {
		// TODO: assuming that is duplicate key.
		return app.ErrIdempotencyKeyViolation
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
