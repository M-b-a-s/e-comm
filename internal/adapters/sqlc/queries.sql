-- name: ListProducts :many
SELECT * FROM products;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (name, price_in_cents, quantity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name = $2,
    price_in_cents = $3,
    quantity = $4
WHERE id = $1
RETURNING id, name, price_in_cents, quantity, created_at;

-- name: DeleteProduct :one
DELETE FROM products
WHERE id = $1
RETURNING id, name, price_in_cents, quantity, created_at;