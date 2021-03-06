package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/stone-co/the-amazing-ledger/app"
	"github.com/stone-co/the-amazing-ledger/app/domain/entities"
	"github.com/stone-co/the-amazing-ledger/app/domain/vos"
)

func (l *LedgerUseCase) CreateTransaction(ctx context.Context, id uuid.UUID, entries []entities.Entry) error {
	transaction, err := entities.NewTransaction(id, entries...)
	if err != nil {
		return err
	}

	accounts := make([]*entities.CachedAccountInfo, 0, len(entries))

	for _, entry := range entries {
		if entry.Account.Suffix == "*" {
			return app.ErrInvalidAccountStructure
		}

		account := l.cachedAccounts.LoadOrStore(entry.Account.Name())
		accounts = append(accounts, account)

		account.Lock()
		defer account.Unlock()

		if entry.Version == vos.AnyAccountVersion {
			continue
		}

		if entry.Version != account.CurrentVersion {
			return app.ErrInvalidVersion
		}
	}

	for i := range entries {
		entries[i].Version = l.lastVersion.Next()
	}

	if err := l.repository.CreateTransaction(ctx, transaction); err != nil {
		return err
	}

	for i := range accounts {
		accounts[i].CurrentVersion = entries[i].Version
	}

	return nil
}
