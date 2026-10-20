CREATE TABLE IF NOT EXISTS payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    order_id BIGINT UNSIGNED NOT NULL,

    payment_method ENUM(
        'bank_transfer',
        'ewallet',
        'virtual_account'
    ) NOT NULL,

    amount DECIMAL(15,2) NOT NULL,

    status ENUM(
        'pending',
        'paid',
        'failed',
        'expired'
    ) NOT NULL DEFAULT 'pending',

    transaction_reference VARCHAR(100) UNIQUE,

    paid_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_payments_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id),

    INDEX idx_payments_order (order_id),
    INDEX idx_payments_status (status)
);