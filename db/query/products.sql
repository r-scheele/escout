-- name: CheckUserProduct :one
SELECT * FROM products 
WHERE user_id = $1 AND link = $2
LIMIT 1;

-- name: CreateProduct :one
INSERT INTO products (user_id, name, link, base_price, percentage_change, tracking_frequency, created_at, cron_job_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: ListProductsByUserID :many
SELECT * FROM products WHERE user_id = $1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: UpdateProductPrice :one
UPDATE products
SET 
    base_price = $2
WHERE id = $1 
RETURNING *;

-- name: UpdateCronJobId :one
UPDATE products
SET 
    cron_job_id = $2
WHERE id = $1
RETURNING *;

-- name: UpdateProduct :one
UPDATE products 
SET is_active = $2, percentage_change = $3, tracking_frequency = $4, name = $5
WHERE id = $1
RETURNING *;

-- name: UpdateJobStatus :one
UPDATE products
SET 
    is_active = $2
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: GetProductByLinkAndUserID :one
SELECT id, name, link, base_price, percentage_change, created_at
FROM products
WHERE link = $1 AND user_id = $2;

-- name: GetProductsByStore :many
SELECT * FROM products WHERE link LIKE $1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: GetProductsByTimeRange :many
SELECT * FROM products WHERE created_at BETWEEN $1 AND $2 ORDER BY id LIMIT $3 OFFSET $4;

-- name: GetProductsByAveragePrice :many
SELECT product_id, AVG(price) as average_price
FROM price_changes
GROUP BY product_id
HAVING AVG(price) BETWEEN $1 AND $2;
