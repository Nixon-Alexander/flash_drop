ALTER TABLE orders
  DROP FOREIGN KEY fk_orders_flash_sale,
  DROP INDEX idx_orders_pending_expiry,
  DROP COLUMN expires_at,
  DROP COLUMN flash_sale_id;