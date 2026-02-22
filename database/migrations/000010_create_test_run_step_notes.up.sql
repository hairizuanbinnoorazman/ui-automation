CREATE TABLE test_run_step_notes (
    id CHAR(36) PRIMARY KEY,
    test_run_id CHAR(36) NOT NULL,
    step_index INT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (test_run_id) REFERENCES test_runs(id) ON DELETE CASCADE,
    UNIQUE KEY unique_run_step (test_run_id, step_index)
);
