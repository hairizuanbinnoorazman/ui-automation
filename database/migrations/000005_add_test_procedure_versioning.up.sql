ALTER TABLE test_procedures
    ADD COLUMN version INT UNSIGNED NOT NULL DEFAULT 1,
    ADD COLUMN is_latest BOOLEAN DEFAULT TRUE,
    ADD COLUMN parent_id CHAR(36) NULL,
    ADD INDEX idx_parent_id (parent_id),
    ADD INDEX idx_version (version),
    ADD INDEX idx_is_latest (is_latest),
    ADD CONSTRAINT fk_test_procedure_parent FOREIGN KEY (parent_id) REFERENCES test_procedures(id) ON DELETE SET NULL;
