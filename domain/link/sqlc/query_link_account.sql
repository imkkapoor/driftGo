-- name: CreateLinkAccount :one
INSERT INTO link_account (
    account_id,
    item_id,
    user_id,
    name,
    official_name,
    mask,
    subtype,
    type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetLinkAccountByID :one
SELECT * FROM link_account
WHERE id = $1;

-- name: GetLinkAccountByAccountID :one
SELECT * FROM link_account
WHERE account_id = $1;

-- name: GetLinkAccountsByItemID :many
SELECT * FROM link_account
WHERE item_id = $1
ORDER BY created_at DESC;

-- name: GetLinkAccountsByUserID :many
SELECT * FROM link_account
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteLinkAccount :exec
DELETE FROM link_account
WHERE id = $1;

-- name: DeleteLinkAccountByAccountID :exec
DELETE FROM link_account
WHERE account_id = $1;

-- name: DeleteLinkAccountsByItemID :exec
DELETE FROM link_account
WHERE item_id = $1; 