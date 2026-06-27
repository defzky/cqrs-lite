-- Enable pg_trgm for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Brands table
CREATE TABLE IF NOT EXISTS brands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    price NUMERIC(15,2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    image_url VARCHAR(500) DEFAULT 'https://dummyimage.com/600x400/cccccc/000000&text=No+Image',
    category_id UUID REFERENCES categories(id) ON DELETE RESTRICT,
    brand_id UUID REFERENCES brands(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for search
CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON products USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_description_trgm ON products USING gin (description gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_brand_id ON products(brand_id);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock);

-- Full-text search index
CREATE INDEX IF NOT EXISTS idx_products_search ON products USING gin (
    to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''))
);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_brands_updated_at
    BEFORE UPDATE ON brands
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data
INSERT INTO categories (name, description) VALUES
    ('Electronics', 'Electronic devices and accessories'),
    ('Accessories', 'Computer accessories and peripherals'),
    ('Audio', 'Audio equipment and speakers')
ON CONFLICT (name) DO NOTHING;

INSERT INTO brands (name, description) VALUES
    ('Logitech', 'Computer peripherals manufacturer'),
    ('Razer', 'Gaming peripherals'),
    ('Sony', 'Electronics and entertainment')
ON CONFLICT (name) DO NOTHING;

INSERT INTO products (sku, name, description, price, stock, category_id, brand_id)
SELECT 
    'SKU-001', 
    'Wireless Mouse', 
    '2.4GHz wireless mouse with ergonomic design', 
    150000.00, 
    50,
    (SELECT id FROM categories WHERE name = 'Electronics'),
    (SELECT id FROM brands WHERE name = 'Logitech')
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'SKU-001');

INSERT INTO products (sku, name, description, price, stock, category_id, brand_id)
SELECT 
    'SKU-002', 
    'Mechanical Keyboard', 
    'RGB mechanical keyboard with Cherry MX switches', 
    850000.00, 
    30,
    (SELECT id FROM categories WHERE name = 'Accessories'),
    (SELECT id FROM brands WHERE name = 'Razer')
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'SKU-002');

INSERT INTO products (sku, name, description, price, stock, category_id, brand_id)
SELECT 
    'SKU-003', 
    'Wireless Headphones', 
    'Bluetooth headphones with noise cancellation', 
    1200000.00, 
    20,
    (SELECT id FROM categories WHERE name = 'Audio'),
    (SELECT id FROM brands WHERE name = 'Sony')
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'SKU-003');
