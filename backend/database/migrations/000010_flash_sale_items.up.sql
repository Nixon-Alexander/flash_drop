CREATE TABLE IF NOT EXISTS flash_sale_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    flash_sale_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,

    flash_price DECIMAL(15,2) NOT NULL,

    stock INT UNSIGNED NOT NULL,

    sold_quantity INT UNSIGNED NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_flash_sale_items_flash_sale
        FOREIGN KEY (flash_sale_id)
        REFERENCES flash_sales(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_flash_sale_items_product
        FOREIGN KEY (product_id)
        REFERENCES products(id),

    UNIQUE KEY uq_flash_sale_product (
        flash_sale_id,
        product_id
    ),

    INDEX idx_flash_sale_items_product (product_id)
);