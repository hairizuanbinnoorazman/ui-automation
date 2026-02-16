CREATE TABLE IF NOT EXISTS test_run_assets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    test_run_id BIGINT UNSIGNED NOT NULL,
    asset_type ENUM('image', 'video', 'binary', 'document') NOT NULL,
    asset_path VARCHAR(512) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT UNSIGNED NOT NULL,
    mime_type VARCHAR(128),
    description TEXT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (test_run_id) REFERENCES test_runs(id) ON DELETE CASCADE,
    INDEX idx_test_run_id (test_run_id),
    INDEX idx_asset_type (asset_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
