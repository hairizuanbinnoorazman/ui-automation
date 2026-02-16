ALTER TABLE test_procedures
    DROP FOREIGN KEY fk_test_procedure_parent,
    DROP INDEX idx_is_latest,
    DROP INDEX idx_version,
    DROP INDEX idx_parent_id,
    DROP COLUMN parent_id,
    DROP COLUMN is_latest,
    DROP COLUMN version;
