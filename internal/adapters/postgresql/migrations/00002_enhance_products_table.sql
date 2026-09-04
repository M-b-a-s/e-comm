-- +goose Up

-- Categories: fixed set of 3 (headphones, speakers, earphones) per PRD §5
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO categories (slug, name) VALUES
    ('headphones', 'Headphones'),
    ('speakers', 'Speakers'),
    ('earphones', 'Earphones')
ON CONFLICT (slug) DO NOTHING;

-- Extend products with catalog/detail fields from the Figma product pages
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS slug TEXT,
    ADD COLUMN IF NOT EXISTS short_name TEXT,
    ADD COLUMN IF NOT EXISTS category_id BIGINT REFERENCES categories(id),
    ADD COLUMN IF NOT EXISTS is_new BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS features TEXT,
    ADD COLUMN IF NOT EXISTS box_includes JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS gallery JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS category_image TEXT,
    ADD COLUMN IF NOT EXISTS recommended_product_ids BIGINT[] NOT NULL DEFAULT '{}';

-- slug is required and unique once populated, but can't be NOT NULL until
-- existing rows (if any) are backfilled, so enforce it via a partial unique
-- index now and tighten to NOT NULL in a later migration once data is seeded.
CREATE UNIQUE INDEX IF NOT EXISTS products_slug_unique_idx
    ON products (slug)
    WHERE slug IS NOT NULL;

CREATE INDEX IF NOT EXISTS products_category_id_idx
    ON products (category_id);

-- quantity/inventory tracking is out of scope per PRD §2 (Non-Goals) and §10
-- (assumed unlimited stock for v1)
ALTER TABLE products
    DROP COLUMN IF EXISTS quantity;

-- +goose Down

DROP INDEX IF EXISTS products_category_id_idx;
DROP INDEX IF EXISTS products_slug_unique_idx;

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 0;

ALTER TABLE products
    DROP COLUMN IF EXISTS recommended_product_ids,
    DROP COLUMN IF EXISTS category_image,
    DROP COLUMN IF EXISTS gallery,
    DROP COLUMN IF EXISTS box_includes,
    DROP COLUMN IF EXISTS features,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS is_new,
    DROP COLUMN IF EXISTS category_id,
    DROP COLUMN IF EXISTS short_name,
    DROP COLUMN IF EXISTS slug;

DROP TABLE IF EXISTS categories;