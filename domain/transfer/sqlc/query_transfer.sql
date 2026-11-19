-- name: CreateTransfer :one
INSERT INTO transfer (
    sender_user_id,
    receiver_user_id,
    sender_account_id,
    receiver_account_id,
    amount,
    currency,
    description,
    idempotency_key
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetTransferByID :one
SELECT * FROM transfer
WHERE id = $1;

-- name: GetTransfersByUserID :many
SELECT * FROM transfer
WHERE sender_user_id = $1 OR receiver_user_id = $1
ORDER BY created_at DESC;

-- name: GetTransfersBySenderUserID :many
SELECT * FROM transfer
WHERE sender_user_id = $1
ORDER BY created_at DESC;

-- name: GetTransfersByReceiverUserID :many
SELECT * FROM transfer
WHERE receiver_user_id = $1
ORDER BY created_at DESC;

-- name: GetTransferByDebitTransferID :one
SELECT * FROM transfer
WHERE debit_transfer_id = $1;

-- name: GetTransferByCreditTransferID :one
SELECT * FROM transfer
WHERE credit_transfer_id = $1;

-- name: GetTransfersByStatus :many
SELECT * FROM transfer
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateTransferDebit :exec
UPDATE transfer
SET debit_authorization_id = $2,
    debit_transfer_id = $3,
    status = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateTransferCredit :exec
UPDATE transfer
SET credit_authorization_id = $2,
    credit_transfer_id = $3,
    status = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateTransferStatus :exec
UPDATE transfer
SET status = $2,
    failure_reason = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateTransferCompleted :exec
UPDATE transfer
SET status = 'completed',
    updated_at = CURRENT_TIMESTAMP,
    completed_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: LockTransferEventCursor :one
SELECT last_event_id FROM transfer_event_cursor
WHERE id = 1
FOR UPDATE;

-- name: UpdateTransferEventCursor :exec
UPDATE transfer_event_cursor
SET last_event_id = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 1;
