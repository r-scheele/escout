-- name: CreatePriceChange :one
INSERT INTO price_changes (product_id, price, changed_at, created_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetPriceChangeByID :one
SELECT * FROM price_changes WHERE id = $1 LIMIT 1;

-- name: ListPriceChangesByProductID :many
SELECT * FROM price_changes WHERE product_id = $1 ORDER BY changed_at DESC LIMIT $2 OFFSET $3;

-- name: UpdatePriceChange :one
UPDATE price_changes SET price = $2, changed_at = $3 WHERE id = $1 RETURNING *;

-- name: DeletePriceChange :exec
DELETE FROM price_changes WHERE id = $1;

-- name: GetPriceChangesByTimeRange :many
SELECT * FROM price_changes WHERE changed_at BETWEEN $1 AND $2 ORDER BY changed_at DESC LIMIT $3 OFFSET $4;

-- name: GetPriceChangesByPriceRange :many
SELECT * FROM price_changes WHERE price BETWEEN $1 AND $2 ORDER BY changed_at DESC LIMIT $3 OFFSET $4;
