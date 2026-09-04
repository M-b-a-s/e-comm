-- name: ListProducts :many
SELECT * FROM products
ORDER BY id;

-- name: ListProductsByCategory :many
SELECT p.*
FROM products p
JOIN categories c ON c.id = p.category_id
WHERE c.slug = $1
ORDER BY p.is_new DESC, p.id;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: GetProductBySlug :one
SELECT * FROM products WHERE slug = $1;

-- name: GetFeaturedProduct :one
SELECT * FROM products
WHERE is_new = TRUE
ORDER BY id
LIMIT 1;

-- name: GetRecommendedProducts :many
SELECT p.* FROM products p
WHERE p.id = ANY (
    SELECT unnest(rp.recommended_product_ids) FROM products rp WHERE rp.id = $1
);

-- name: CreateProduct :one
INSERT INTO products (
    name,
    price_in_cents,
    slug,
    short_name,
    category_id,
    is_new,
    description,
    features,
    box_includes,
    gallery,
    category_image,
    recommended_product_ids
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name = $2,
    price_in_cents = $3,
    slug = $4,
    short_name = $5,
    category_id = $6,
    is_new = $7,
    description = $8,
    features = $9,
    box_includes = $10,
    gallery = $11,
    category_image = $12,
    recommended_product_ids = $13
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :one
DELETE FROM products
WHERE id = $1
RETURNING *;