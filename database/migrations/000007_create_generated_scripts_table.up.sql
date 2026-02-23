CREATE TABLE IF NOT EXISTS generated_scripts (
    id CHAR(36) PRIMARY KEY,
    test_procedure_id CHAR(36) NOT NULL,
    framework ENUM('selenium', 'playwright') NOT NULL,
    script_path VARCHAR(512) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT UNSIGNED NOT NULL,
    generation_status ENUM('pending', 'generating', 'completed', 'failed') NOT NULL DEFAULT 'pending',
    error_message TEXT,
    generated_by CHAR(36) NOT NULL,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (test_procedure_id) REFERENCES test_procedures(id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by) REFERENCES users(id) ON DELETE RESTRICT,
    UNIQUE KEY unique_procedure_framework (test_procedure_id, framework)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
