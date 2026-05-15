CREATE TABLE IF NOT EXISTS profiles (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id      VARCHAR(36)  NOT NULL UNIQUE,
    email        VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NULL,
    avatar_url   VARCHAR(500) NULL,
    bio          TEXT         NULL,
    language     VARCHAR(10)  NOT NULL DEFAULT 'en',
    timezone     VARCHAR(50)  NOT NULL DEFAULT 'UTC',
    created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
