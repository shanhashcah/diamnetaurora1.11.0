package resourceadapter

import (
	"context"
	"fmt"

	"github.com/diamnet/go/amount"
	protocol "github.com/diamnet/go/protocols/aurora"
	auroraContext "github.com/diamnet/go/services/aurora/internal/context"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/hal"
	"github.com/diamnet/go/xdr"
)

// PopulateClaimableBalance fills out the resource's fields
func PopulateClaimableBalance(
	ctx context.Context,
	dest *protocol.ClaimableBalance,
	claimableBalance history.ClaimableBalance,
	ledger *history.Ledger,
) error {
	balanceID, err := xdr.MarshalHex(claimableBalance.BalanceID)
	if err != nil {
		return errors.Wrap(err, "marshalling BalanceID")
	}
	dest.BalanceID = balanceID
	dest.Asset = claimableBalance.Asset.StringCanonical()
	dest.Amount = amount.StringFromInt64(int64(claimableBalance.Amount))
	if claimableBalance.Sponsor.Valid {
		dest.Sponsor = claimableBalance.Sponsor.String
	}
	dest.LastModifiedLedger = claimableBalance.LastModifiedLedger
	dest.Claimants = make([]protocol.Claimant, len(claimableBalance.Claimants))
	for i, c := range claimableBalance.Claimants {
		dest.Claimants[i].Destination = c.Destination
		dest.Claimants[i].Predicate = c.Predicate
	}

	if ledger != nil {
		dest.LastModifiedTime = &ledger.ClosedAt
	}

	lb := hal.LinkBuilder{Base: auroraContext.BaseURL(ctx)}
	self := fmt.Sprintf("/claimable_balances/%s", dest.BalanceID)
	dest.Links.Self = lb.Link(self)
	dest.PT = fmt.Sprintf("%d-%s", claimableBalance.LastModifiedLedger, dest.BalanceID)
	return nil
}
