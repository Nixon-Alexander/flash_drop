ALTER TABLE products DROP INDEX slug;
ALTER TABLE payments
MODIFY COLUMN payment_method ENUM("gopay", "ovo", "shopeepay", "permata_va", "bca_va", "bni_va", "bri_va", "cimb_va", "danamon_va", "bsi_va", "seabank_va", "saqu_va");

ALTER TABLE payments
DROP FOREIGN KEY fk_payments_order;

ALTER TABLE payments
MODIFY COLUMN order_id varchar(100);

ALTER TABLE payments
ADD CONSTRAINT fk_payments_order
FOREIGN KEY (order_id)
REFERENCES orders (order_number);