package transfer

import (
	"context"
	"driftGo/api/common/utils"
	"driftGo/domain/link"
	"driftGo/domain/user"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plaid/plaid-go/v35/plaid"
)

const testEncryptionKey = "test-encryption-key-0123456789abcdef"

type authCall struct {
	AccessToken string `json:"access_token"`
	AccountID   string `json:"account_id"`
	Type        string `json:"type"`
	User        struct {
		LegalName string `json:"legal_name"`
	} `json:"user"`
}

type fakeEvent struct {
	EventID    int32  `json:"event_id"`
	EventType  string `json:"event_type"`
	TransferID string `json:"transfer_id"`
	Timestamp  string `json:"timestamp"`
}

// fakePlaid stands in for the Plaid Transfer API and records every call made to it.
type fakePlaid struct {
	mu           sync.Mutex
	authCalls    []authCall
	createCount  int
	syncAfterIDs []int32
	events       []fakeEvent
}

func (f *fakePlaid) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/transfer/authorization/create", func(w http.ResponseWriter, r *http.Request) {
		var c authCall
		_ = json.NewDecoder(r.Body).Decode(&c)
		f.mu.Lock()
		f.authCalls = append(f.authCalls, c)
		n := len(f.authCalls)
		f.mu.Unlock()
		writeJSON(w, map[string]any{
			"authorization": map[string]any{
				"id":       "auth-" + string(rune('0'+n)),
				"created":  "2026-01-01T00:00:00Z",
				"decision": "approved",
			},
			"request_id": "req",
		})
	})
	mux.HandleFunc("/transfer/create", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			AccountID string `json:"account_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.createCount++
		f.mu.Unlock()
		writeJSON(w, map[string]any{
			"transfer":   map[string]any{"id": "plaid-transfer-" + body.AccountID},
			"request_id": "req",
		})
	})
	mux.HandleFunc("/transfer/event/sync", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			AfterID int32 `json:"after_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.syncAfterIDs = append(f.syncAfterIDs, body.AfterID)
		out := []fakeEvent{}
		for _, e := range f.events {
			if e.EventID > body.AfterID {
				out = append(out, e)
			}
		}
		f.mu.Unlock()
		writeJSON(w, map[string]any{"transfer_events": out, "has_more": false, "request_id": "req"})
	})
	return mux
}

func (f *fakePlaid) addEvent(eventType, plaidTransferID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, fakeEvent{
		EventID:    int32(len(f.events) + 1),
		EventType:  eventType,
		TransferID: plaidTransferID,
		Timestamp:  "2026-01-01T00:00:00Z",
	})
}

func (f *fakePlaid) creditAuthCalls() []authCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []authCall
	for _, c := range f.authCalls {
		if c.Type == "credit" {
			out = append(out, c)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

type testPerson struct {
	ID int64
	// EncryptedAccountID is the account ID exactly as stored in link_account,
	// which is the form the transfer API expects from clients.
	EncryptedAccountID string
	PlainAccountID     string
	AccessToken        string
}

type env struct {
	svc                 *Service
	plaid               *fakePlaid
	pool                *pgxpool.Pool
	alice, bob, mallory testPerson
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping transfer integration tests")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(ctx, "TRUNCATE users, link_item, link_account, transfer RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	// Reset the event-sync cursor if the table exists.
	_, _ = pool.Exec(ctx, "UPDATE transfer_event_cursor SET last_event_id = 0")

	userSvc := user.NewService(pool)
	linkSvc, err := link.NewService("id", "secret", "sandbox", "", userSvc, pool, testEncryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	fp := &fakePlaid{}
	srv := httptest.NewServer(fp.handler())
	t.Cleanup(srv.Close)

	svc := newTestService(t, userSvc, linkSvc, pool)
	cfg := plaid.NewConfiguration()
	cfg.Servers = plaid.ServerConfigurations{{URL: srv.URL}}
	svc.client = plaid.NewAPIClient(cfg)

	e := &env{svc: svc, plaid: fp, pool: pool}
	e.alice = makePerson(t, userSvc, linkSvc, pool, "alice", "Alice", "Anders")
	e.bob = makePerson(t, userSvc, linkSvc, pool, "bob", "Bob", "Brown")
	e.mallory = makePerson(t, userSvc, linkSvc, pool, "mallory", "Mallory", "Moss")
	return e
}

func makePerson(t *testing.T, us *user.Service, ls *link.Service, pool *pgxpool.Pool, handle, first, last string) testPerson {
	t.Helper()
	ctx := context.Background()
	u, err := us.CreateUser(ctx, "stytch-"+handle, first, last, handle+"@example.com", "active")
	if err != nil {
		t.Fatal(err)
	}
	item, err := ls.CreateLinkItem(ctx, u.ID, "access-"+handle, "item-"+handle, "ins_1", "Test Bank")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ls.CreateLinkAccount(ctx, "acct-"+handle, item.ID, u.ID, "Checking", "", "0000", "checking", "depository"); err != nil {
		t.Fatal(err)
	}
	var enc string
	if err := pool.QueryRow(ctx, "SELECT account_id FROM link_account WHERE user_id = $1", u.ID).Scan(&enc); err != nil {
		t.Fatal(err)
	}
	return testPerson{ID: u.ID, EncryptedAccountID: enc, PlainAccountID: "acct-" + handle, AccessToken: "access-" + handle}
}

func as(p testPerson) context.Context {
	return utils.WithAuthContext(context.Background(), utils.AuthContext{UserID: p.ID})
}

func (e *env) status(t *testing.T, id int64) TransferStatus {
	t.Helper()
	tr, err := e.svc.database.GetTransferByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return tr.Status
}

func (e *env) debitTransferID(t *testing.T, id int64) string {
	t.Helper()
	tr, err := e.svc.database.GetTransferByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return tr.DebitTransferID.String
}

func (e *env) creditTransferID(t *testing.T, id int64) string {
	t.Helper()
	tr, err := e.svc.database.GetTransferByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return tr.CreditTransferID.String
}

func newTestService(t *testing.T, us *user.Service, ls *link.Service, pool *pgxpool.Pool) *Service {
	t.Helper()
	svc, err := NewService("id", "secret", "sandbox", us, ls.GetAccountOwnerAndAccessToken, pool, testEncryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}
