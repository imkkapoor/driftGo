package transfer

import (
	"context"
	"driftGo/pkg/logger"
	"fmt"

	"github.com/plaid/plaid-go/v35/plaid"
)

/*
ProcessTransferWebhook fetches transfer events from Plaid that haven't been
processed yet and updates transfer statuses accordingly.

The last processed event ID is kept in transfer_event_cursor. The cursor row is
locked for the whole sync, so concurrent webhooks run one at a time instead of
processing the same events twice. If the sync fails partway, the cursor rolls
back and those events are processed again on the next webhook; the status
guards and the idempotent credit leg make that safe.
*/
func (s *Service) ProcessTransferWebhook(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transfer event sync: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cursor := New(tx)
	afterID, err := cursor.LockTransferEventCursor(ctx)
	if err != nil {
		return fmt.Errorf("failed to lock transfer event cursor: %w", err)
	}

	for {
		request := plaid.NewTransferEventSyncRequest(afterID)
		response, _, err := s.client.PlaidApi.TransferEventSync(ctx).TransferEventSyncRequest(*request).Execute()
		if err != nil {
			if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
				logger.Log.Error("Plaid transfer event sync error: ", string(plaidErr.Body()))
			}
			return fmt.Errorf("failed to sync transfer events: %w", err)
		}

		events := response.GetTransferEvents()
		if len(events) == 0 {
			break
		}

		for _, event := range events {
			if err := s.processTransferEvent(ctx, event); err != nil {
				logger.Log.WithError(err).WithField("event_id", event.GetEventId()).Error("Failed to process transfer event")
			}
			afterID = event.GetEventId()
		}

		if err := cursor.UpdateTransferEventCursor(ctx, afterID); err != nil {
			return fmt.Errorf("failed to advance transfer event cursor: %w", err)
		}

		if !response.GetHasMore() {
			break
		}
	}

	return tx.Commit(ctx)
}

/*
statusRank orders the non-terminal path a transfer moves along. Event-driven
updates only ever move a transfer forward on this path.
*/
var statusRank = map[TransferStatus]int{
	TransferStatusPending:          0,
	TransferStatusDebitAuthorized:  1,
	TransferStatusDebitPending:     2,
	TransferStatusDebitSettled:     3,
	TransferStatusCreditAuthorized: 4,
	TransferStatusCreditPending:    5,
	TransferStatusCompleted:        6,
}

func isStopped(status TransferStatus) bool {
	return status == TransferStatusFailed || status == TransferStatusReturned || status == TransferStatusCancelled
}

/*
advanceTo moves a transfer forward to status, ignoring the update if the
transfer is already at or past it, or has stopped (failed/returned/cancelled).
It reports whether the status changed.
*/
func (s *Service) advanceTo(ctx context.Context, transfer *Transfer, status TransferStatus) (bool, error) {
	if isStopped(transfer.Status) || statusRank[transfer.Status] >= statusRank[status] {
		return false, nil
	}
	if status == TransferStatusCompleted {
		return true, s.MarkTransferCompleted(ctx, transfer.ID)
	}
	return true, s.UpdateTransferStatus(ctx, transfer.ID, status, "")
}

/*
stop marks a transfer failed, returned or cancelled. A completed transfer can
still be marked returned, because ACH returns can arrive after settlement.
*/
func (s *Service) stop(ctx context.Context, transfer *Transfer, status TransferStatus, reason string) error {
	if isStopped(transfer.Status) {
		return nil
	}
	if transfer.Status == TransferStatusCompleted && status != TransferStatusReturned {
		return nil
	}
	return s.UpdateTransferStatus(ctx, transfer.ID, status, reason)
}

// processTransferEvent processes a single Plaid transfer event
func (s *Service) processTransferEvent(ctx context.Context, event plaid.TransferEvent) error {
	transferID := event.GetTransferId()
	eventType := event.GetEventType()

	logger.Log.WithField("transfer_id", transferID).WithField("event_type", eventType).Info("Processing transfer event")

	// Try to find the transfer by debit transfer ID first, then credit transfer ID
	transfer, err := s.GetTransferByDebitTransferID(ctx, transferID)
	isDebitLeg := true
	if err != nil {
		if err == ErrTransferNotFound {
			transfer, err = s.GetTransferByCreditTransferID(ctx, transferID)
			if err != nil {
				if err == ErrTransferNotFound {
					logger.Log.WithField("plaid_transfer_id", transferID).Debug("Transfer not found for event, skipping")
					return nil
				}
				return err
			}
			isDebitLeg = false
		} else {
			return err
		}
	}

	leg := "credit"
	if isDebitLeg {
		leg = "debit"
	}

	switch eventType {
	case plaid.TRANSFEREVENTTYPE_PENDING, plaid.TRANSFEREVENTTYPE_POSTED:
		// "posted" means the money is on its way, not that it has arrived, so
		// it never starts the credit leg or completes the transfer.
		if isDebitLeg {
			_, err := s.advanceTo(ctx, transfer, TransferStatusDebitPending)
			return err
		}
		_, err := s.advanceTo(ctx, transfer, TransferStatusCreditPending)
		return err

	case plaid.TRANSFEREVENTTYPE_SETTLED:
		if !isDebitLeg {
			_, err := s.advanceTo(ctx, transfer, TransferStatusCompleted)
			return err
		}
		changed, err := s.advanceTo(ctx, transfer, TransferStatusDebitSettled)
		if err != nil || !changed {
			return err
		}
		logger.Log.WithField("transfer_id", transfer.ID).Info("Debit settled, initiating credit leg")
		return s.InitiateCreditLeg(ctx, transfer.ID)

	case plaid.TRANSFEREVENTTYPE_FAILED:
		return s.stop(ctx, transfer, TransferStatusFailed, fmt.Sprintf("%s leg failed", leg))

	case plaid.TRANSFEREVENTTYPE_RETURNED:
		return s.stop(ctx, transfer, TransferStatusReturned, fmt.Sprintf("%s leg returned", leg))

	case plaid.TRANSFEREVENTTYPE_CANCELLED:
		return s.stop(ctx, transfer, TransferStatusCancelled, fmt.Sprintf("%s leg cancelled", leg))

	default:
		logger.Log.WithField("event_type", eventType).WithField("transfer_id", transferID).Info("Unhandled transfer event type")
	}

	return nil
}
