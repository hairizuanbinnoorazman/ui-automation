CREATE TABLE IF NOT EXISTS issue_links (
    id CHAR(36) PRIMARY KEY,
    test_run_id CHAR(36) NOT NULL,
    integration_id CHAR(36) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    title VARCHAR(500),
    status VARCHAR(50),
    url VARCHAR(1000),
    provider VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (test_run_id) REFERENCES test_runs(id) ON DELETE RESTRICT,
    FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE RESTRICT,
    INDEX idx_issue_links_test_run_id (test_run_id),
    INDEX idx_issue_links_integration_id (integration_id),
    UNIQUE INDEX idx_issue_links_unique (test_run_id, integration_id, external_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
