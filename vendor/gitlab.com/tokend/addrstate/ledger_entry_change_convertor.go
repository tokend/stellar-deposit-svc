package addrstate

import (
	"gitlab.com/tokend/regources/generated"
	"gitlab.com/tokend/go/xdr"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

var ErrUnexpectedEffect = errors.New("unexpected change effect")

func convertLedgerEntryChange(change regources.LedgerEntryChange) (xdr.LedgerEntryChange, error) {
	switch change.Attributes.ChangeType {
	case xdr.LedgerEntryChangeTypeRemoved:
		var ledgerKey xdr.LedgerKey
		err := xdr.SafeUnmarshalBase64(change.Attributes.Payload, &ledgerKey)
		if err != nil {
			return xdr.LedgerEntryChange{}, errors.Wrap(err, "failed to unmarshal ledger key", logan.F{
				"xdr" : change.Attributes.Payload,
			})
		}
		return xdr.NewLedgerEntryChange(xdr.LedgerEntryChangeType(change.Attributes.ChangeType), ledgerKey)
	case xdr.LedgerEntryChangeTypeCreated, xdr.LedgerEntryChangeTypeUpdated:
		var ledgerEntry xdr.LedgerEntry
		err := xdr.SafeUnmarshalBase64(change.Attributes.Payload, &ledgerEntry)
		if err != nil {
			return xdr.LedgerEntryChange{}, errors.Wrap(err, "failed to unmarshal ledger entry", logan.F{
				"xdr" : change.Attributes.Payload,
			})
		}
		return xdr.NewLedgerEntryChange(xdr.LedgerEntryChangeType(change.Attributes.ChangeType), ledgerEntry)
	default:
		return xdr.LedgerEntryChange{}, errors.Wrap(ErrUnexpectedEffect, "failed to convert ledger entry",
			logan.F{"effect" : change.Attributes.ChangeType})
	}
}
