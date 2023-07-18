// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: products.sql

package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const checkUserProduct = `-- name: CheckUserProduct :one
SELECT id, user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at FROM products 
WHERE user_id = $1 AND link = $2
LIMIT 1
`

type CheckUserProductParams struct {
	UserID int64  `json:"user_id"`
	Link   string `json:"link"`
}

func (q *Queries) CheckUserProduct(ctx context.Context, arg CheckUserProductParams) (Product, error) {
	row := q.db.QueryRow(ctx, checkUserProduct, arg.UserID, arg.Link)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.Link,
		&i.BasePrice,
		&i.PercentageChange,
		&i.TrackingFrequency,
		&i.NotificationThreshold,
		&i.CreatedAt,
	)
	return i, err
}

const createProduct = `-- name: CreateProduct :one
INSERT INTO products (user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at
`

type CreateProductParams struct {
	UserID                int64     `json:"user_id"`
	Name                  string    `json:"name"`
	Link                  string    `json:"link"`
	BasePrice             float64   `json:"base_price"`
	PercentageChange      float64   `json:"percentage_change"`
	TrackingFrequency     int32     `json:"tracking_frequency"`
	NotificationThreshold float64   `json:"notification_threshold"`
	CreatedAt             time.Time `json:"created_at"`
}

func (q *Queries) CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error) {
	row := q.db.QueryRow(ctx, createProduct,
		arg.UserID,
		arg.Name,
		arg.Link,
		arg.BasePrice,
		arg.PercentageChange,
		arg.TrackingFrequency,
		arg.NotificationThreshold,
		arg.CreatedAt,
	)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.Link,
		&i.BasePrice,
		&i.PercentageChange,
		&i.TrackingFrequency,
		&i.NotificationThreshold,
		&i.CreatedAt,
	)
	return i, err
}

const deleteProduct = `-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1
`

func (q *Queries) DeleteProduct(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteProduct, id)
	return err
}

const getProductByID = `-- name: GetProductByID :one
SELECT id, user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at FROM products WHERE id = $1 LIMIT 1
`

func (q *Queries) GetProductByID(ctx context.Context, id int64) (Product, error) {
	row := q.db.QueryRow(ctx, getProductByID, id)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.Link,
		&i.BasePrice,
		&i.PercentageChange,
		&i.TrackingFrequency,
		&i.NotificationThreshold,
		&i.CreatedAt,
	)
	return i, err
}

const getProductByLinkAndUserID = `-- name: GetProductByLinkAndUserID :one
SELECT id, name, link, base_price, percentage_change, created_at
FROM products
WHERE link = $1 AND user_id = $2
`

type GetProductByLinkAndUserIDParams struct {
	Link   string `json:"link"`
	UserID int64  `json:"user_id"`
}

type GetProductByLinkAndUserIDRow struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Link             string    `json:"link"`
	BasePrice        float64   `json:"base_price"`
	PercentageChange float64   `json:"percentage_change"`
	CreatedAt        time.Time `json:"created_at"`
}

func (q *Queries) GetProductByLinkAndUserID(ctx context.Context, arg GetProductByLinkAndUserIDParams) (GetProductByLinkAndUserIDRow, error) {
	row := q.db.QueryRow(ctx, getProductByLinkAndUserID, arg.Link, arg.UserID)
	var i GetProductByLinkAndUserIDRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Link,
		&i.BasePrice,
		&i.PercentageChange,
		&i.CreatedAt,
	)
	return i, err
}

const getProductsByAveragePrice = `-- name: GetProductsByAveragePrice :many
SELECT product_id, AVG(price) as average_price
FROM price_changes
GROUP BY product_id
HAVING AVG(price) BETWEEN $1 AND $2
`

type GetProductsByAveragePriceParams struct {
	Price   pgtype.Numeric `json:"price"`
	Price_2 pgtype.Numeric `json:"price_2"`
}

type GetProductsByAveragePriceRow struct {
	ProductID    int64   `json:"product_id"`
	AveragePrice float64 `json:"average_price"`
}

func (q *Queries) GetProductsByAveragePrice(ctx context.Context, arg GetProductsByAveragePriceParams) ([]GetProductsByAveragePriceRow, error) {
	rows, err := q.db.Query(ctx, getProductsByAveragePrice, arg.Price, arg.Price_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetProductsByAveragePriceRow{}
	for rows.Next() {
		var i GetProductsByAveragePriceRow
		if err := rows.Scan(&i.ProductID, &i.AveragePrice); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProductsByStore = `-- name: GetProductsByStore :many
SELECT id, user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at FROM products WHERE link LIKE $1 ORDER BY id LIMIT $2 OFFSET $3
`

type GetProductsByStoreParams struct {
	Link   string `json:"link"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

func (q *Queries) GetProductsByStore(ctx context.Context, arg GetProductsByStoreParams) ([]Product, error) {
	rows, err := q.db.Query(ctx, getProductsByStore, arg.Link, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Name,
			&i.Link,
			&i.BasePrice,
			&i.PercentageChange,
			&i.TrackingFrequency,
			&i.NotificationThreshold,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProductsByTimeRange = `-- name: GetProductsByTimeRange :many
SELECT id, user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at FROM products WHERE created_at BETWEEN $1 AND $2 ORDER BY id LIMIT $3 OFFSET $4
`

type GetProductsByTimeRangeParams struct {
	CreatedAt   time.Time `json:"created_at"`
	CreatedAt_2 time.Time `json:"created_at_2"`
	Limit       int32     `json:"limit"`
	Offset      int32     `json:"offset"`
}

func (q *Queries) GetProductsByTimeRange(ctx context.Context, arg GetProductsByTimeRangeParams) ([]Product, error) {
	rows, err := q.db.Query(ctx, getProductsByTimeRange,
		arg.CreatedAt,
		arg.CreatedAt_2,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Name,
			&i.Link,
			&i.BasePrice,
			&i.PercentageChange,
			&i.TrackingFrequency,
			&i.NotificationThreshold,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listProductsByUserID = `-- name: ListProductsByUserID :many
SELECT id, user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at FROM products WHERE user_id = $1 ORDER BY id LIMIT $2 OFFSET $3
`

type ListProductsByUserIDParams struct {
	UserID int64 `json:"user_id"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListProductsByUserID(ctx context.Context, arg ListProductsByUserIDParams) ([]Product, error) {
	rows, err := q.db.Query(ctx, listProductsByUserID, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		var i Product
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Name,
			&i.Link,
			&i.BasePrice,
			&i.PercentageChange,
			&i.TrackingFrequency,
			&i.NotificationThreshold,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateProduct = `-- name: UpdateProduct :one
UPDATE products SET name = $2, link = $3, base_price = $4, percentage_change = $5, tracking_frequency = $6, notification_threshold = $7
WHERE id = $1 RETURNING id, user_id, name, link, base_price, percentage_change, tracking_frequency, notification_threshold, created_at
`

type UpdateProductParams struct {
	ID                    int64   `json:"id"`
	Name                  string  `json:"name"`
	Link                  string  `json:"link"`
	BasePrice             float64 `json:"base_price"`
	PercentageChange      float64 `json:"percentage_change"`
	TrackingFrequency     int32   `json:"tracking_frequency"`
	NotificationThreshold float64 `json:"notification_threshold"`
}

func (q *Queries) UpdateProduct(ctx context.Context, arg UpdateProductParams) (Product, error) {
	row := q.db.QueryRow(ctx, updateProduct,
		arg.ID,
		arg.Name,
		arg.Link,
		arg.BasePrice,
		arg.PercentageChange,
		arg.TrackingFrequency,
		arg.NotificationThreshold,
	)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.Link,
		&i.BasePrice,
		&i.PercentageChange,
		&i.TrackingFrequency,
		&i.NotificationThreshold,
		&i.CreatedAt,
	)
	return i, err
}
