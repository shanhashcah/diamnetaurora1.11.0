package ingest

import (
	"github.com/diamnet/go/ingest/io"
	"github.com/diamnet/go/support/errors"
)

type auroraChangeProcessor interface {
	io.ChangeProcessor
	// TODO maybe rename to Flush()
	Commit() error
}

type groupChangeProcessors []auroraChangeProcessor

func (g groupChangeProcessors) ProcessChange(change io.Change) error {
	for _, p := range g {
		if err := p.ProcessChange(change); err != nil {
			return errors.Wrapf(err, "error in %T.ProcessChange", p)
		}
	}
	return nil
}

func (g groupChangeProcessors) Commit() error {
	for _, p := range g {
		if err := p.Commit(); err != nil {
			return errors.Wrapf(err, "error in %T.Commit", p)
		}
	}
	return nil
}

type groupTransactionProcessors []auroraTransactionProcessor

func (g groupTransactionProcessors) ProcessTransaction(tx io.LedgerTransaction) error {
	for _, p := range g {
		if err := p.ProcessTransaction(tx); err != nil {
			return errors.Wrapf(err, "error in %T.ProcessTransaction", p)
		}
	}
	return nil
}

func (g groupTransactionProcessors) Commit() error {
	for _, p := range g {
		if err := p.Commit(); err != nil {
			return errors.Wrapf(err, "error in %T.Commit", p)
		}
	}
	return nil
}
