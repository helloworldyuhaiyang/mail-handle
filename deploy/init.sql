-- Set character set and collation
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
SET character_set_connection=utf8mb4;

-- Create forward_targets table
CREATE TABLE IF NOT EXISTS forward_targets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    email VARCHAR(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample data
INSERT INTO forward_targets (name, email) VALUES 
('yang', 'helloworldyang9@gmail.com'),
('张三', 'zhangsan@example.com'),
('李四', 'lisi@example.com')
ON DUPLICATE KEY UPDATE 
    email = VALUES(email),
    updated_at = CURRENT_TIMESTAMP; 