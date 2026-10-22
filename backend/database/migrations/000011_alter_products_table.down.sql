ALTER TABLE products DROP INDEX slug;

ALTER TABLE payments
DROP FOREIGN KEY fk_payments_order;