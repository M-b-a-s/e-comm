-- name: ListCategories :many
SELECT * FROM categories
ORDER BY id;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;