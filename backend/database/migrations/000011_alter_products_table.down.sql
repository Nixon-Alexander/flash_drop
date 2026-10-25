ALTER TABLE payments DROP FOREIGN KEY fk_payments_order;

ALTER TABLE payments
  MODIFY COLUMN order_id BIGINT UNSIGNED NOT NULL;

ALTER TABLE payments
  MODIFY COLUMN payment_method ENUM('bank_transfer','ewallet','virtual_account') NOT NULL;

ALTER TABLE payments
  ADD CONSTRAINT fk_payments_order
  FOREIGN KEY (order_id) REFERENCES orders (id);