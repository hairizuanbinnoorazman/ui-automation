CREATE TABLE IF NOT EXISTS jobs (
    id CHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    config JSON,
    result JSON,
    start_time TIMESTAMP NULL,
    end_time TIMESTAMP NULL,
    duration BIGINT NULL,
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_jobs_type (type),
    INDEX idx_jobs_status (status),
    INDEX idx_jobs_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
