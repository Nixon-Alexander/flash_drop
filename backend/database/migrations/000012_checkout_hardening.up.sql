ALTER TABLE orders
  ADD COLUMN flash_sale_id BIGINT UNSIGNED NULL AFTER user_id,
  ADD COLUMN expires_at DATETIME NULL AFTER status,
  ADD CONSTRAINT fk_orders_flash_sale FOREIGN KEY (flash_sale_id) REFERENCES flash_sales(id),
  ADD INDEX idx_orders_pending_expiry (status, expires_at);

ALTER TABLE payments
  MODIFY COLUMN order_id VARCHAR(100) NOT NULL,
  MODIFY COLUMN payment_method ENUM('gopay','ovo','shopeepay','permata_va','bca_va','bni_va',
    'bri_va','cimb_va','danamon_va','bsi_va','seabank_va','saqu_va') NOT NULL;