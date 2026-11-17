package transfer

import (
	"context"
	"driftGo/pkg/logger"
	"fmt"

	"github.com/plaid/plaid-go/v35/plaid"
)

// ProcessTransferWebhook fetches new transfer events from Plaid using Transfer Event Sync
// and updates transfer statuses accordingly.
// When a debit settles, the credit leg is automatically initiated.
// When a credit settles, the transfer is marked as completed.
func (s *Service) ProcessTransferWebhook(ctx context.Context) error {
	afterID := int32(0)

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
	}

	return nil
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

	switch eventType {
	case plaid.TRANSFEREVENTTYPE_PENDING:
		if isDebitLeg {
			return s.UpdateTransferStatus(ctx, transfer.ID, TransferStatusDebitPending, "")
		}
		return s.UpdateTransferStatus(ctx, transfer.ID, TransferStatusCreditPending, "")

	case plaid.TRANSFEREVENTTYPE_POSTED:
		if isDebitLeg {
			err := s.UpdateTransferStatus(ctx, transfer.ID, TransferStatusDebitSettled, "")
			if err != nil {
				return err
			}
			logger.Log.WithField("transfer_id", transfer.ID).Info("Debit settled, initiating credit leg")
			return s.InitiateCreditLeg(ctx, transfer.ID)
		}
		return s.MarkTransferCompleted(ctx, transfer.ID)

	case plaid.TRANSFEREVENTTYPE_SETTLED:
		if isDebitLeg {
			err := s.UpdateTransferStatus(ctx, transfer.ID, TransferStatusDebitSettled, "")
			if err != nil {
				return err
			}
			logger.Log.WithField("transfer_id", transfer.ID).Info("Debit settled, initiating credit leg")
			return s.InitiateCreditLeg(ctx, transfer.ID)
		}
		return s.MarkTransferCompleted(ctx, transfer.ID)

	case plaid.TRANSFEREVENTTYPE_FAILED:
		failureReason := fmt.Sprintf("Transfer failed: %s", eventType)
		return s.UpdateTransferStatus(ctx, transfer.ID, TransferStatusFailed, failureReason)

	case plaid.TRANSFEREVENTTYPE_RETURNED:
		failureReason := fmt.Sprintf("Transfer returned: %s", eventType)
		return s.UpdateTransferStatus(ctx, transfer.ID, TransferStatusReturned, failureReason)

	case plaid.TRANSFEREVENTTYPE_CANCELLED:
		return s.UpdateTransferStatus(ctx, transfer.ID, TransferStatusCancelled, "Transfer cancelled")

	default:
		logger.Log.WithField("event_type", eventType).WithField("transfer_id", transferID).Info("Unhandled transfer event type")
	}

	return nil
}
