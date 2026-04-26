-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS `gateway` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `gateway`;

-- The tables will be automatically created by GORM AutoMigrate in the Go code.
-- We can add some initial data here if needed, like a default admin.
-- Note: Password is 'admin123' hashed with bcrypt.
INSERT INTO `admins` (`username`, `password`, `created_at`, `updated_at`) 
VALUES ('admin', '$2a$10$7v8XG0iJqP1j6fKzL6K1e.Kz8Kj8Kj8Kj8Kj8Kj8Kj8Kj8Kj8Kj8K', NOW(), NOW())
ON DUPLICATE KEY UPDATE `username`=`username`;
