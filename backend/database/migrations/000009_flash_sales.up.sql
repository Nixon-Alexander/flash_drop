CREATE TABLE IF NOT EXISTS flash_sales (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    name VARCHAR(200) NOT NULL,
    description TEXT,

    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,

    status ENUM(
        'scheduled',
        'active',
        'ended',
        'cancelled'
    ) NOT NULL DEFAULT 'scheduled',

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_flash_sales_time (start_time, end_time),
    INDEX idx_flash_sales_status (status)
);