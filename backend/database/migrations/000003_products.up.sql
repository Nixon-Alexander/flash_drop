CREATE TABLE IF NOT EXISTS products (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    category_id BIGINT UNSIGNED NOT NULL,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(220) NOT NULL UNIQUE,

    description TEXT,

    price DECIMAL(15,2) NOT NULL,
    stock INT UNSIGNED NOT NULL DEFAULT 0,

    image_url VARCHAR(500),

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id)
        REFERENCES categories(id),

    INDEX idx_products_category (category_id),
    INDEX idx_products_active (is_active)
);