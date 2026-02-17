CREATE TABLE IF NOT EXISTS test_runs (
    id CHAR(36) PRIMARY KEY,
    test_procedure_id CHAR(36) NOT NULL,
    executed_by CHAR(36) NOT NULL,
    status ENUM('pending', 'running', 'passed', 'failed', 'skipped') NOT NULL DEFAULT 'pending',
    notes TEXT,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (test_procedure_id) REFERENCES test_procedures(id) ON DELETE CASCADE,
    FOREIGN KEY (executed_by) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_test_procedure_id (test_procedure_id),
    INDEX idx_executed_by (executed_by),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
