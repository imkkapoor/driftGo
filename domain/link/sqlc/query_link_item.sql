-- name: CreateLinkItem :one
INSERT INTO link_item (
    user_id,
    access_token,
    item_id,
    institution_id,
    institution_name
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetLinkItemByID :one
SELECT * FROM link_item
WHERE id = $1;

-- name: GetLinkItemByItemID :one
SELECT * FROM link_item
WHERE item_id = $1;

-- name: GetLinkItemsByUserID :many
SELECT * FROM link_item
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetAccessTokenByAccountID :one
SELECT li.access_token
FROM link_account la
JOIN link_item li ON la.item_id = li.id
WHERE la.account_id = $1;

-- name: DeleteLinkItem :exec
DELETE FROM link_item
WHERE id = $1;

-- name: DeleteLinkItemByItemID :exec
DELETE FROM link_item
WHERE item_id = $1; 