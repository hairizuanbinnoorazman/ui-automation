-- Remove indexes
DROP INDEX idx_root_lookup ON test_procedures;
DROP INDEX idx_version_is_latest ON test_procedures;

-- Revert version default back to 1
ALTER TABLE test_procedures MODIFY COLUMN version INT UNSIGNED NOT NULL DEFAULT 1;
