-- name: CreatePriceChange :one
INSERT INTO price_changes (product_id, price, created_at) VALUES ($1, $2, $3) RETURNING *;

-- name: GetPriceChangesForUserAndProduct :many
-- name: GetPriceChangesForUserAndProduct :many
SELECT pc.id, pc.product_id, pc.price, pc.created_at
FROM price_changes pc
JOIN products p ON pc.product_id = p.id
JOIN users u ON u.id = p.user_id
WHERE u.id = $1 AND pc.product_id = $2;


-- name: ListPriceChangesByProductID :many
SELECT * FROM price_changes WHERE product_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdatePriceChange :one
UPDATE price_changes SET price = $2, created_at = $3 WHERE id = $1 RETURNING *;

-- name: DeletePriceChange :exec
DELETE FROM price_changes WHERE id = $1;

-- name: GetPriceChangesByTimeRange :many
SELECT * FROM price_changes WHERE created_at BETWEEN $1 AND $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: GetPriceChangesByPriceRange :many
SELECT * FROM price_changes WHERE price BETWEEN $1 AND $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4;
