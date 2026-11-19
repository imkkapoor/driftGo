package transfer

import (
	"context"
	"testing"
)

// Bug 1: after the sender's money settles, the receiver must actually be paid.
func TestCreditLegPaysReceiver(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	tr, err := e.svc.InitiateTransfer(as(e.alice), e.alice.EncryptedAccountID, e.bob.ID, e.bob.EncryptedAccountID, "50.00", "dinner")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}

	e.plaid.addEvent("settled", e.debitTransferID(t, tr.ID))
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatalf("webhook: %v", err)
	}

	credits := e.plaid.creditAuthCalls()
	if len(credits) != 1 {
		t.Fatalf("expected 1 credit authorization to Plaid, got %d; transfer status=%s", len(credits), e.status(t, tr.ID))
	}
	if credits[0].AccountID != e.bob.PlainAccountID || credits[0].AccessToken != e.bob.AccessToken {
		t.Fatalf("credit went to account=%q token=%q, want Bob's", credits[0].AccountID, credits[0].AccessToken)
	}
	if got := e.status(t, tr.ID); got != TransferStatusCreditPending {
		t.Fatalf("status = %s, want credit_pending", got)
	}

	e.plaid.addEvent("settled", e.creditTransferID(t, tr.ID))
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if got := e.status(t, tr.ID); got != TransferStatusCompleted {
		t.Fatalf("status = %s, want completed", got)
	}
}

// Bug 2a: a user must not be able to debit an account they don't own.
func TestCannotSendFromSomeoneElsesAccount(t *testing.T) {
	e := newEnv(t)

	_, err := e.svc.InitiateTransfer(as(e.mallory), e.alice.EncryptedAccountID, e.bob.ID, e.bob.EncryptedAccountID, "500.00", "")
	if err == nil {
		t.Fatalf("Mallory sent money from Alice's account; Plaid debit calls: %+v", e.plaid.authCalls)
	}
	if len(e.plaid.authCalls) != 0 {
		t.Fatalf("expected no Plaid calls, got %+v", e.plaid.authCalls)
	}
}

// Bug 2b: the receiving account must belong to the named receiver.
func TestReceiverAccountMustBelongToReceiver(t *testing.T) {
	e := newEnv(t)

	_, err := e.svc.InitiateTransfer(as(e.alice), e.alice.EncryptedAccountID, e.bob.ID, e.mallory.EncryptedAccountID, "50.00", "")
	if err == nil {
		t.Fatal("transfer addressed to Bob was accepted with Mallory's account as the destination")
	}
}

// Bug 3: each webhook must only process events it hasn't seen, so a finished
// transfer isn't dragged back or re-paid when an unrelated webhook arrives.
func TestWebhookDoesNotReplayOldEvents(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	tr, err := e.svc.InitiateTransfer(as(e.alice), e.alice.EncryptedAccountID, e.bob.ID, e.bob.EncryptedAccountID, "50.00", "")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	debitID := e.debitTransferID(t, tr.ID)
	e.plaid.addEvent("pending", debitID)
	e.plaid.addEvent("settled", debitID)
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatalf("webhook 1: %v", err)
	}
	statusAfterFirst := e.status(t, tr.ID)
	syncCallsBefore := len(e.plaid.syncAfterIDs)

	// An unrelated webhook arrives later.
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatalf("webhook 2: %v", err)
	}

	if first := e.plaid.syncAfterIDs[syncCallsBefore]; first != 2 {
		t.Errorf("second webhook started syncing from after_id=%d, want 2 (last event already seen)", first)
	}
	if got := e.status(t, tr.ID); got != statusAfterFirst {
		t.Errorf("status changed from %s to %s on a webhook with no new events", statusAfterFirst, got)
	}
	if n := len(e.plaid.creditAuthCalls()); n != 1 {
		t.Errorf("receiver payout was started %d times across the two webhooks, want 1", n)
	}
}

// Bug 4: the credit leg must start exactly once, even though Plaid reports
// both "posted" and "settled" for the debit.
func TestCreditLegStartsOnce(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	tr, err := e.svc.InitiateTransfer(as(e.alice), e.alice.EncryptedAccountID, e.bob.ID, e.bob.EncryptedAccountID, "50.00", "")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	debitID := e.debitTransferID(t, tr.ID)
	e.plaid.addEvent("posted", debitID)
	e.plaid.addEvent("settled", debitID)
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatalf("webhook: %v", err)
	}

	if n := len(e.plaid.creditAuthCalls()); n != 1 {
		t.Fatalf("credit leg started %d times, want 1", n)
	}
}

// A finished transfer must not be dragged back by a late or stale event.
func TestCompletedTransferIgnoresStaleEvents(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	tr, err := e.svc.InitiateTransfer(as(e.alice), e.alice.EncryptedAccountID, e.bob.ID, e.bob.EncryptedAccountID, "50.00", "")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	debitID := e.debitTransferID(t, tr.ID)
	e.plaid.addEvent("settled", debitID)
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatal(err)
	}
	e.plaid.addEvent("settled", e.creditTransferID(t, tr.ID))
	e.plaid.addEvent("pending", debitID) // stale, out-of-order event
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatal(err)
	}
	if got := e.status(t, tr.ID); got != TransferStatusCompleted {
		t.Fatalf("status = %s, want completed", got)
	}
}

// ACH returns can arrive after a transfer completes and must still be recorded.
func TestReturnAfterCompletionIsRecorded(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()

	tr, err := e.svc.InitiateTransfer(as(e.alice), e.alice.EncryptedAccountID, e.bob.ID, e.bob.EncryptedAccountID, "50.00", "")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	debitID := e.debitTransferID(t, tr.ID)
	e.plaid.addEvent("settled", debitID)
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatal(err)
	}
	e.plaid.addEvent("settled", e.creditTransferID(t, tr.ID))
	e.plaid.addEvent("returned", debitID)
	if err := e.svc.ProcessTransferWebhook(ctx); err != nil {
		t.Fatal(err)
	}
	if got := e.status(t, tr.ID); got != TransferStatusReturned {
		t.Fatalf("status = %s, want returned", got)
	}
}
